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
	"strings"
	"time"

	"github.com/TaoYang526/goutils/pkg/profiling"
	"github.com/apache/yunikorn-release/perf-tools/framework"
	"github.com/apache/yunikorn-release/perf-tools/utils"
	"go.uber.org/zap"
)

const E2EPerfScenarioName = "e2e_perf"

type E2EPerfScenario struct {
	kubeClient   *utils.KubeClient
	commonConf   *framework.CommonConfig
	scenarioConf *E2EPerfScenarioConfig
}

type E2EPerfScenarioConfig struct {
	CleanUpDelayMs     int
	ShowNumOfLastTasks int
	Cases              []*E2EPerfCaseConfig
}

type E2EPerfCaseConfig struct {
	Description    string
	SchedulerName  string
	RequestConfigs []*RequestConfig
}

func init() {
	framework.Register(&E2EPerfScenario{})
}

func (ts *E2EPerfScenario) GetName() string {
	return E2EPerfScenarioName
}

func (ts *E2EPerfScenario) Init(kubeClient *utils.KubeClient, conf *framework.Config) error {
	ts.kubeClient = kubeClient
	ts.commonConf = conf.Common
	ts.scenarioConf = &E2EPerfScenarioConfig{}
	return LoadScenarioConf(conf, ts.GetName(), ts.scenarioConf)
}

func (eps *E2EPerfScenario) Run(results *utils.Results) {
	scenarioResults := results.CreateScenarioResults(eps.GetName())
	maxWaitTime := time.Duration(eps.commonConf.MaxWaitSeconds) * time.Second
	var appManager framework.AppManager
	var appInfo *framework.AppInfo
	// make sure app is cleaned up when error occurred
	defer func() {
		CleanupApp(appManager, appInfo, maxWaitTime)
	}()

	for caseIndex, testCase := range eps.scenarioConf.Cases {
		verGroupName := fmt.Sprintf("Case-%d", caseIndex)
		verGroupDescription := fmt.Sprintf("%+v", testCase.Description)
		caseVerification := scenarioResults.AddVerificationGroup(verGroupName, verGroupDescription)
		utils.Logger.Info("[Prepare] add verification group",
			zap.Int("caseIndex", caseIndex),
			zap.String("name", verGroupName),
			zap.String("description", verGroupDescription))
		// init app info, app manager, app analyzer and node analyzer
		requestInfos := ConvertToRequestInfos(testCase.RequestConfigs)
		appInfo = framework.NewAppInfo(eps.commonConf.Namespace, E2EPerfScenarioName, eps.commonConf.Queue,
			requestInfos, eps.commonConf.PodTemplateSpec, eps.commonConf.PodSpec)
		appManager = framework.NewDeploymentsAppManager(eps.kubeClient)
		appAnalyzer := framework.NewAppAnalyzer(appInfo)
		nodeAnalyzer := framework.NewNodeAnalyzer(eps.kubeClient, eps.commonConf.NodeSelector)

		schedulerName := testCase.SchedulerName

		// prepare nodes
		err := nodeAnalyzer.InitNodeInfosBeforeTesting()
		if err != nil {
			utils.Logger.Error("failed to init nodes", zap.Error(err))
			caseVerification.AddSubVerification("init nodes", err.Error(), utils.FAILED)
			return
		}
		utils.Logger.Info("[Prepare] init nodes", zap.Int("numNodes", len(nodeAnalyzer.GetAllocatableNodes())))

		// create app and wait for it to be running
		utils.Logger.Info("[Testing] create an app and wait for it to be running, refresh tasks status at last",
			zap.String("appID", appInfo.AppID))
		beginTime := time.Now().Truncate(time.Second)
		err = appManager.CreateWaitAndRefreshTasksStatus(schedulerName, appInfo, maxWaitTime)
		if err != nil {
			utils.Logger.Error("failed to create/wait/refresh app", zap.Error(err))
			caseVerification.AddSubVerification("test app", err.Error(), utils.FAILED)
			return
		}
		utils.Logger.Info("all requirements of this app are satisfied",
			zap.String("appID", appInfo.AppID),
			zap.Duration("elapseTime", time.Since(beginTime)))

		// fulfill nodes info after testing
		nodeAnalyzer.AnalyzeApp(appInfo)
		scheduledNodes := nodeAnalyzer.GetScheduledNodes()
		utils.Logger.Info("got related nodes", zap.Int("numScheduledNodes", len(scheduledNodes)))

		// analyze-1: print slow tasks (optional)
		if eps.scenarioConf.ShowNumOfLastTasks > 0 {
			slowTasksStatus := appAnalyzer.GetLastTasks(eps.scenarioConf.ShowNumOfLastTasks)
			utils.Logger.Info(fmt.Sprintf("[Analyze] Show last %d tasks: ", len(slowTasksStatus)))
			for _, task := range slowTasksStatus {
				utils.Logger.Info("task status",
					zap.String("taskID", task.TaskID),
					zap.String("nodeID", task.NodeID),
					zap.Duration("to-running-duration", task.RunningTime.Sub(task.CreateTime)),
					zap.Time("createTime", task.CreateTime),
					zap.Time("runningTime", task.RunningTime))
			}
		}
		// analyze-2: print tasks distribution on nodes
		tasksDistributionInfo := appAnalyzer.GetTasksDistributionInfo(scheduledNodes)
		utils.Logger.Info("[Analyze] tasks distribution info on nodes",
			zap.Int("LeastNum", tasksDistributionInfo.LeastNum),
			zap.String("LeastNumNodeID", tasksDistributionInfo.LeastNumNodeID),
			zap.Int("MostNum", tasksDistributionInfo.MostNum),
			zap.String("MostNumNodeID", tasksDistributionInfo.MostNumNodeID),
			zap.Float64("AvgNumPerNode", tasksDistributionInfo.AvgNum),
			zap.Int("NumTasks", len(appInfo.TasksStatus)),
			zap.Int("NumScheduledNodes", len(scheduledNodes)))
		utils.Logger.Info("node with least number of tasks", zap.Any("summary",
			tasksDistributionInfo.SortedNodeInfos[0].GetSummary()))
		utils.Logger.Info("node with most number of tasks", zap.Any("summary",
			tasksDistributionInfo.SortedNodeInfos[len(tasksDistributionInfo.SortedNodeInfos)-1].GetSummary()))

		// profiling
		prof := appAnalyzer.GetTasksProfiling()
		if prof.GetCount() > 0 {
			utils.Logger.Info("[Analyze] time statistics for pod conditions")
			statsTableFilePath := fmt.Sprintf("%s/%s-case%d-timecost-stat.txt",
				eps.commonConf.OutputPath, eps.GetName(), caseIndex)
			statsOutputName := "time statistics"
			stats := prof.GetTimeStatistics()
			statsTable := ParseTableFromStatistic(stats)
			if err := statsTable.Output(statsTableFilePath); err != nil {
				caseVerification.AddSubVerification(statsOutputName,
					fmt.Sprintf("failed to output %s: %s", statsOutputName, err.Error()),
					utils.FAILED)
				return
			}
			caseVerification.AddSubVerification(statsOutputName, statsTableFilePath, utils.SUCCEEDED)
			statsTable.Print()
			utils.Logger.Info("[Analyze] QPS statistics for pod conditions")
			qpsStatsTableFilePath := fmt.Sprintf("%s/%s-case%d-qps-stat.txt",
				eps.commonConf.OutputPath, eps.GetName(), caseIndex)
			qpsStatsOutputName := "QPS statistics"
			qpsStat, err := prof.GetQPSStatistics()
			if err != nil {
				caseVerification.AddSubVerification(qpsStatsOutputName,
					fmt.Sprintf("failed to output %s: %s", qpsStatsOutputName, err.Error()),
					utils.FAILED)
			}
			qpsStatsTable := ParseTableFromQPSStatistics(qpsStat, framework.GetOrderedTaskConditionTypes())
			if err := qpsStatsTable.Output(qpsStatsTableFilePath); err != nil {
				caseVerification.AddSubVerification(qpsStatsOutputName,
					fmt.Sprintf("failed to output %s: %s", qpsStatsOutputName, err.Error()),
					utils.FAILED)
				return
			}
			caseVerification.AddSubVerification(qpsStatsOutputName, qpsStatsTableFilePath, utils.SUCCEEDED)
			qpsStatsTable.Print()
		}

		if eps.scenarioConf.CleanUpDelayMs > 0 {
			utils.Logger.Info("wait for a while before cleaning up test apps",
				zap.String("schedulerName", schedulerName),
				zap.Any("cleanUpDelayMs", eps.scenarioConf.CleanUpDelayMs))
			time.Sleep(time.Millisecond * time.Duration(eps.scenarioConf.CleanUpDelayMs))
		}

		// delete this app and wait for it to be cleaned up
		utils.Logger.Info("[Cleanup] delete this app then wait for it to be cleaned up",
			zap.String("appID", appInfo.AppID))
		err = appManager.DeleteWait(appInfo, maxWaitTime)
		if err != nil {
			utils.Logger.Error("failed to delete/wait app", zap.Error(err))
			caseVerification.AddSubVerification("cleanup app", err.Error(), utils.FAILED)
			return
		}
	}
}

func ParseTableFromStatistic(statistics *profiling.TimeStatistics) *utils.Table {
	var data [][]string
	for _, stat := range statistics.StagesTime {
		rowData := []string{stat.From, stat.To, strconv.Itoa(stat.Count),
			stat.TotalTime.String(), stat.AvgTime.String(),
			fmt.Sprintf("%.2f", stat.Percentage)}
		data = append(data, rowData)
	}
	return &utils.Table{
		Headers: []string{"From", "To", "Count", "TotalTime", "AvgTime", "Percentage"},
		Data:    data,
	}
}

func ParseTableFromQPSStatistics(statistics *profiling.QPSStatistics, orderedKeys []framework.TaskConditionType) *utils.Table {
	var data [][]string
	if len(orderedKeys) > 0 {
		for _, key := range orderedKeys {
			if stat := getStageQPS(statistics.StagesQPS, ")"+string(key)); stat != nil {
				rowData := []string{stat.To, strconv.Itoa(stat.MaxQPS),
					fmt.Sprintf("%.2f", stat.AvgQPS), fmt.Sprintf("%+v", stat.Samples)}
				data = append(data, rowData)
			}
		}
	} else {
		for _, stat := range statistics.StagesQPS {
			rowData := []string{stat.To, strconv.Itoa(stat.MaxQPS),
				fmt.Sprintf("%.2f", stat.AvgQPS), fmt.Sprintf("%+v", stat.Samples)}
			data = append(data, rowData)
		}
	}
	return &utils.Table{
		Headers: []string{"To", "MaxQPS", "AvgQPS", "Samples"},
		Data:    data,
	}
}

func getStageQPS(qpsMap map[string]*profiling.StageQPS, suffix string) *profiling.StageQPS {
	for key, item := range qpsMap {
		if strings.HasSuffix(key, suffix) {
			return item
		}
	}
	return nil
}
