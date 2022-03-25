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
	"time"

	"github.com/apache/yunikorn-release/perf-tools/constants"
	"github.com/apache/yunikorn-release/perf-tools/framework"
	"github.com/apache/yunikorn-release/perf-tools/utils"
	"go.uber.org/zap"
)

const ThroughputScenarioName = "throughput"

type ThroughputScenario struct {
	kubeClient   *utils.KubeClient
	commonConf   *framework.CommonConfig
	scenarioConf *ThroughputScenarioConfig
}

type ThroughputScenarioConfig struct {
	CleanUpDelayMs int
	SchedulerNames []string
	Cases          []*ThroughputCaseConfig
}

type ThroughputCaseConfig struct {
	Description    string
	RequestConfigs []*RequestConfig
}

func init() {
	framework.Register(&ThroughputScenario{})
}

func (ts *ThroughputScenario) GetName() string {
	return ThroughputScenarioName
}

func (ts *ThroughputScenario) Init(kubeClient *utils.KubeClient, conf *framework.Config) error {
	ts.kubeClient = kubeClient
	ts.commonConf = conf.Common
	ts.scenarioConf = &ThroughputScenarioConfig{}
	return LoadScenarioConf(conf, ts.GetName(), ts.scenarioConf)
}

func (ts *ThroughputScenario) Run(results *utils.Results) {
	scenarioResults := results.CreateScenarioResults(ts.GetName())
	schedulerNames := ts.scenarioConf.SchedulerNames
	maxWaitTime := time.Duration(ts.commonConf.MaxWaitSeconds) * time.Second
	var appManager framework.AppManager
	var appInfo *framework.AppInfo
	var appAnanyzer *framework.AppAnalyzer
	// make sure app is cleaned up when error occurred
	defer func() {
		CleanupApp(appManager, appInfo, maxWaitTime)
	}()

	for caseIndex, testCase := range ts.scenarioConf.Cases {
		verGroupName := fmt.Sprintf("Case-%d", caseIndex)
		verGroupDescription := fmt.Sprintf("%+v", testCase.Description)
		caseVerification := scenarioResults.AddVerificationGroup(verGroupName, verGroupDescription)
		utils.Logger.Info("[Prepare] add verification group", zap.String("name", verGroupName),
			zap.String("description", verGroupDescription))
		// init app info & app manager
		requestInfos := ConvertToRequestInfos(testCase.RequestConfigs)
		appInfo = framework.NewAppInfo(ts.commonConf.Namespace, ThroughputScenarioName, ts.commonConf.Queue,
			requestInfos, ts.commonConf.PodTemplateSpec, ts.commonConf.PodSpec)
		appManager = framework.NewDeploymentsAppManager(ts.kubeClient)
		appAnanyzer = framework.NewAppAnalyzer(appInfo)

		// test for different schedulers
		cumulativeDistributions := make(map[string][]int, len(schedulerNames))
		for _, schedulerName := range schedulerNames {
			utils.Logger.Info("start testing for scheduler " + schedulerName)
			schedulerVerification := caseVerification.AddSubVerificationGroup(
				fmt.Sprintf("test for %s", schedulerName), verGroupDescription)

			// create app and wait for it to be running
			utils.Logger.Info("[Testing] create an app and wait for it to be running, refresh app status at last",
				zap.String("appID", appInfo.AppID))
			beginTime := time.Now().Truncate(time.Second)
			err := appManager.CreateWaitAndRefreshTasksStatus(schedulerName, appInfo, maxWaitTime)
			if err != nil {
				utils.Logger.Error("failed to create/wait/refresh app", zap.Error(err))
				schedulerVerification.AddSubVerification("test app", err.Error(), utils.FAILED)
				return
			}
			utils.Logger.Info("all requirements of this app are satisfied",
				zap.String("appID", appInfo.AppID),
				zap.Duration("elapseTime", time.Since(beginTime)))

			// calculate scheduled time distribution and its cumulative distribution
			scheduledTimeDistribution := appAnanyzer.GetTimeDistribution(framework.PodScheduled)
			cumulativeDistributions[schedulerName] = getCumulativeDistribution(scheduledTimeDistribution)

			if ts.scenarioConf.CleanUpDelayMs > 0 {
				utils.Logger.Info("wait for a while before cleaning up test apps",
					zap.String("schedulerName", schedulerName),
					zap.Any("cleanUpDelayMs", ts.scenarioConf.CleanUpDelayMs))
				time.Sleep(time.Millisecond * time.Duration(ts.scenarioConf.CleanUpDelayMs))
			}

			// delete this app and wait for it to be cleaned up
			utils.Logger.Info("[Cleanup] delete this app then wait for it to be cleaned up",
				zap.String("appID", appInfo.AppID))
			err = appManager.DeleteWait(appInfo, maxWaitTime)
			if err != nil {
				utils.Logger.Error("failed to delete/wait app", zap.Error(err))
				schedulerVerification.AddSubVerification("cleanup app", err.Error(), utils.FAILED)
				return
			}

			description := fmt.Sprintf("seconds: %d, throughputDistribution: %+v",
				len(scheduledTimeDistribution), scheduledTimeDistribution)
			schedulerVerification.AddSubVerification("get scheduled time distribution", description, utils.SUCCEEDED)
		}
		// draw chart
		linePoints := utils.GetLinePoints(cumulativeDistributions)
		chartFileName := fmt.Sprintf("%s-case%d-%d", ThroughputScenarioName,
			caseIndex, appInfo.GetDesiredNumTasks())
		chart := &utils.Chart{
			Title:      "Scheduling Throughput",
			XLabel:     "Seconds",
			YLabel:     "Number of Scheduled Pods",
			Width:      constants.ChartWidth,
			Height:     constants.ChartHeight,
			LinePoints: linePoints,
			SvgFile:    ts.commonConf.OutputPath + "/" + chartFileName + constants.ChartFileSuffix,
		}
		err := utils.DrawChart(chart)
		outputName := "output chart"
		if err != nil {
			caseVerification.AddSubVerification(outputName,
				fmt.Sprintf("failed to draw chart: %s", err.Error()),
				utils.FAILED)
			return
		}
		caseVerification.AddSubVerification(outputName, chart.SvgFile, utils.SUCCEEDED)
	}
}

func getCumulativeDistribution(data []int) []int {
	cumulativeDistribution := make([]int, len(data))
	cumulativeNum := 0
	for i := range data {
		cumulativeNum += data[i]
		cumulativeDistribution[i] = cumulativeNum
	}
	return cumulativeDistribution
}
