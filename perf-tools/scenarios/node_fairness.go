/*
 Licensed to the Apache Software Foundation (ASF) under one
 or more contributor license agreements.  See the NOTICE file
 distributed with this work for additional information
 regarding copyright ownership.  The ASF licenses this file
 to you under the Apache License, Version 2.0 (the
 "License"); you may not use this file except in compliance
 with the License.  You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package scenarios

import (
	"fmt"
	"strconv"
	"time"

	"github.com/apache/incubator-yunikorn-core/pkg/common/resources"
	"github.com/apache/incubator-yunikorn-release/perf-tools/constants"
	"github.com/apache/incubator-yunikorn-release/perf-tools/utils"
	v1 "k8s.io/api/core/v1"

	"github.com/apache/incubator-yunikorn-release/perf-tools/framework"
	"go.uber.org/zap"
)

const NodeFairnessScenarioName = "node_fairness"

type NodeFairnessScenario struct {
	kubeClient   *utils.KubeClient
	commonConf   *framework.CommonConfig
	scenarioConf *NodeFairnessScenarioConfig
}

type NodeFairnessScenarioConfig struct {
	SchedulerNames []string
	Cases          []NodeFairnessCaseConfig
}

type NodeFairnessCaseConfig struct {
	NumPodsPerNode     int
	AllocatePercentage int
	ResourceName       string
}

func init() {
	framework.Register(&NodeFairnessScenario{})
}

func (nfs *NodeFairnessScenario) GetName() string {
	return NodeFairnessScenarioName
}

func (nfs *NodeFairnessScenario) Init(kubeClient *utils.KubeClient, conf *framework.Config) error {
	nfs.kubeClient = kubeClient
	nfs.commonConf = conf.Common
	nfs.scenarioConf = &NodeFairnessScenarioConfig{}
	return LoadScenarioConf(conf, nfs.GetName(), nfs.scenarioConf)
}

func (nfs *NodeFairnessScenario) Run(results *utils.Results) {
	scenarioResults := results.CreateScenarioResults(nfs.GetName())
	maxWaitTime := time.Duration(nfs.commonConf.MaxWaitSeconds) * time.Second
	var appManager framework.AppManager
	var appInfo *framework.AppInfo
	// make sure app is cleaned up when error occurred
	defer func() {
		CleanupApp(appManager, appInfo, maxWaitTime)
	}()

	// init node analyzer and calculate allocated resource for nodes
	nodeAnalyzer := framework.NewNodeAnalyzer(nfs.kubeClient, nfs.commonConf.NodeSelector)
	err := nodeAnalyzer.InitNodeInfosBeforeTesting()
	if err != nil {
		utils.Logger.Error("failed to init nodes", zap.Error(err))
		scenarioResults.AddVerification("init nodes", err.Error(), utils.FAILED)
		return
	}
	utils.Logger.Info("[Prepare] init nodes", zap.Int("numNodes", len(nodeAnalyzer.GetAllocatableNodes())))
	nodeAnalyzer.CalculateAllocatedResource()

	for caseIndex, testCase := range nfs.scenarioConf.Cases {
		verGroupName := fmt.Sprintf("Case-%d", caseIndex)
		verGroupDescription := fmt.Sprintf("%+v", testCase)
		caseVerification := scenarioResults.AddVerificationGroup(verGroupName, verGroupDescription)
		utils.Logger.Info("add verification group", zap.String("name", verGroupName),
			zap.String("description", verGroupDescription))

		nodeAnalyzer.ClearApps()
		totalAllocatableResource := nodeAnalyzer.GetTotalAllocatableResource()
		var resourceUnit string
		ykResourceName := testCase.ResourceName
		if testCase.ResourceName == v1.ResourceMemory.String() {
			resourceUnit = "Mi"
		} else if testCase.ResourceName == v1.ResourceCPU.String() {
			resourceUnit = "m"
			ykResourceName = resources.VCORE
		}
		totalAllocatableResourceValue, ok := totalAllocatableResource.Resources[ykResourceName]
		if !ok {
			caseVerification.AddSubVerification("Unknown resource name",
				fmt.Sprintf("resourceName=%s, totalAllocatableResource=%v",
					ykResourceName, totalAllocatableResource), utils.FAILED)
			return
		}

		// init expected number of pods
		allocatableNodes := nodeAnalyzer.GetAllocatableNodes()
		expectedNumPods := testCase.NumPodsPerNode * len(allocatableNodes)
		expectedPodResource := int64(totalAllocatableResourceValue) * int64(testCase.AllocatePercentage) /
			int64(expectedNumPods*100)
		requestResources := make(map[string]string)
		requestResources[testCase.ResourceName] = fmt.Sprintf("%d"+resourceUnit, expectedPodResource)
		utils.Logger.Info("start testing",
			zap.Int("numAllocatableNodes", len(allocatableNodes)),
			zap.Int("expectedNumPods", expectedNumPods),
			zap.Any("requestResources", requestResources))

		// init app info & app manager
		requestInfo := framework.NewRequestInfo(int32(expectedNumPods), "", requestResources, nil)
		appInfo = framework.NewAppInfo(nfs.commonConf.Namespace, NodeFairnessScenarioName, nfs.commonConf.Queue,
			[]*framework.RequestInfo{requestInfo}, nfs.commonConf.PodTemplateSpec, nfs.commonConf.PodSpec)
		appManager = framework.NewDeploymentsAppManager(nfs.kubeClient)
		appAnalyzer := framework.NewAppAnalyzer(appInfo)

		// test for different schedulers
		for _, schedulerName := range nfs.scenarioConf.SchedulerNames {
			utils.Logger.Info("start testing for scheduler " + schedulerName)
			schedulerVerification := caseVerification.AddSubVerificationGroup(
				fmt.Sprintf("test for %s", schedulerName), verGroupDescription)

			// prepare nodes
			nodeAnalyzer.ClearApps()
			utils.Logger.Info("[Prepare] init nodes", zap.Int("numNodes", len(nodeAnalyzer.GetAllocatableNodes())))

			// create app and wait for it to be running
			utils.Logger.Info("create an app and wait for it to be running, refresh tasks status at last",
				zap.String("appID", appInfo.AppID))
			beginTime := time.Now().Truncate(time.Second)
			err = appManager.CreateWaitAndRefreshTasksStatus(schedulerName, appInfo, maxWaitTime)
			if err != nil {
				utils.Logger.Error("failed to create/wait/refresh app", zap.Error(err))
				schedulerVerification.AddSubVerification("test app", err.Error(), utils.FAILED)
				return
			}
			utils.Logger.Info("all requirements of this app are satisfied", zap.String("appID", appInfo.AppID),
				zap.Duration("elapseTime", time.Since(beginTime)))

			// analyze
			nodeDistribution := nodeAnalyzer.GetNodeResourceDistribution(
				appAnalyzer.GetTasksDistribution(framework.PodScheduled), ykResourceName)

			table := parseTableFromNodeDistribution(nodeDistribution)
			tableFilePath := fmt.Sprintf("%s/%s-case%d-node-distribution.txt",
				nfs.commonConf.OutputPath, nfs.GetName(), caseIndex)
			tableOutputName := "output node distribution timeline table"
			utils.Logger.Info(tableOutputName)
			table.Print()
			if err := table.Output(tableFilePath); err != nil {
				schedulerVerification.AddSubVerification(tableOutputName,
					fmt.Sprintf("failed to output node distribution timeline table: %s", err.Error()),
					utils.FAILED)
				return
			}
			schedulerVerification.AddSubVerification(tableOutputName, tableFilePath, utils.SUCCEEDED)

			// prepare line points
			var linePoints []interface{}
			for i, v := range nodeDistribution {
				typeName := "bucket-" + strconv.Itoa(i)
				linePoints = append(linePoints, typeName, utils.GetPointsFromSlice(v))
			}
			// draw chart
			chartFileName := fmt.Sprintf("%s-case%d-%s-%d-%d", nfs.GetName(), caseIndex, schedulerName,
				testCase.NumPodsPerNode, testCase.AllocatePercentage)
			chart := &utils.Chart{
				Title:      "Node Fairness",
				XLabel:     "Seconds",
				YLabel:     "Number of Nodes",
				Width:      constants.ChartWidth,
				Height:     constants.ChartHeight,
				LinePoints: linePoints,
				SvgFile:    nfs.commonConf.OutputPath + "/" + chartFileName + constants.ChartFileSuffix,
			}
			err = utils.DrawChart(chart)
			outputName := "output node distribution timeline chart"
			if err != nil {
				schedulerVerification.AddSubVerification(outputName,
					fmt.Sprintf("failed to draw chart: %s", err.Error()),
					utils.FAILED)
				return
			}
			schedulerVerification.AddSubVerification(outputName, chart.SvgFile, utils.SUCCEEDED)

			// delete this app and wait for it to be cleaned up
			utils.Logger.Info("delete this app then wait for it to be cleaned up",
				zap.String("appID", appInfo.AppID))
			err = appManager.DeleteWait(appInfo, maxWaitTime)
			if err != nil {
				utils.Logger.Error("failed to delete/wait app", zap.Error(err))
				schedulerVerification.AddSubVerification("cleanup app", err.Error(), utils.FAILED)
				return
			}
		}
	}
}

func parseTableFromNodeDistribution(nodeDistribution [10][]int) *utils.Table {
	var data [][]string
	for bucketIndex, bucketData := range nodeDistribution {
		rowData := make([]string, len(bucketData)+1)
		rowData[0] = fmt.Sprintf("bucket-%d", bucketIndex)
		for i, dataItem := range bucketData {
			rowData[i+1] = strconv.Itoa(dataItem)
		}
		data = append(data, rowData)
	}
	headers := make([]string, len(nodeDistribution[0])+1)
	headers[0] = "buckets"
	for i := 0; i < len(nodeDistribution[0]); i++ {
		headers[i+1] = strconv.Itoa(i)
	}
	return &utils.Table{
		Headers: headers,
		Data:    data,
	}
}
