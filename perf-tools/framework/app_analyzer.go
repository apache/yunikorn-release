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

package framework

import (
	"math"
	"sort"

	"github.com/TaoYang526/goutils/pkg/profiling"
	"github.com/apache/yunikorn-release/perf-tools/utils"
	"go.uber.org/zap"
)

type AppAnalyzer struct {
	appInfo *AppInfo
}

func NewAppAnalyzer(appInfo *AppInfo) *AppAnalyzer {
	return &AppAnalyzer{appInfo: appInfo}
}

func (aa *AppAnalyzer) GetLastTasks(lastN int) []*TaskStatus {
	taskStatusSlice := make([]*TaskStatus, len(aa.appInfo.TasksStatus))
	i := 0
	for _, taskStatus := range aa.appInfo.TasksStatus {
		taskStatusSlice[i] = taskStatus
		i++
	}
	sort.SliceStable(taskStatusSlice, func(i, j int) bool {
		return taskStatusSlice[i].RunningTime.Before(taskStatusSlice[j].RunningTime)
	})
	if lastN > len(taskStatusSlice) {
		lastN = len(taskStatusSlice)
	}
	return taskStatusSlice[len(taskStatusSlice)-lastN:]
}

func (aa *AppAnalyzer) GetTasksDistributionInfo(nodeInfos map[string]*NodeInfo) *TasksDistributionInfo {
	nodeInfoSlice := make([]*NodeInfo, len(nodeInfos))
	i := 0
	for _, nodeInfo := range nodeInfos {
		nodeInfoSlice[i] = nodeInfo
		i++
	}
	sort.SliceStable(nodeInfoSlice, func(i, j int) bool {
		return len(nodeInfoSlice[i].Tasks) < len(nodeInfoSlice[j].Tasks)
	})
	leastNodeInfo := nodeInfoSlice[0]
	mostNodeInfo := nodeInfoSlice[len(nodeInfoSlice)-1]
	return &TasksDistributionInfo{
		LeastNum:        len(leastNodeInfo.Tasks),
		LeastNumNodeID:  leastNodeInfo.NodeID,
		MostNum:         len(mostNodeInfo.Tasks),
		MostNumNodeID:   mostNodeInfo.NodeID,
		AvgNum:          float64(len(aa.appInfo.TasksStatus)) / float64(len(nodeInfoSlice)),
		SortedNodeInfos: nodeInfoSlice,
	}
}

func (aa *AppAnalyzer) GetTasksDistribution(condType TaskConditionType) [][]*TaskStatus {
	beginTime := aa.appInfo.AppStatus.CreateTime
	endTime := aa.appInfo.AppStatus.RunningTime
	maxSeconds := int(math.Floor(endTime.Sub(beginTime).Seconds()) + 1)
	distribution := make([][]*TaskStatus, maxSeconds+1)
	for _, status := range aa.appInfo.TasksStatus {
		if condTransitionTime := status.GetTransitionTime(condType); condTransitionTime != nil {
			seconds := int(math.Floor(condTransitionTime.Sub(beginTime).Seconds()) + 1)
			if seconds < 0 {
				utils.Logger.Warn("skip invalid task", zap.Any("beginTime", beginTime),
					zap.Any("conditionType", condType), zap.Any("taskStatus", status))
				continue
			}
			distribution[seconds] = append(distribution[seconds], status)
		} else {
			utils.Logger.Warn("skip invalid task", zap.Any("conditionType", condType),
				zap.Any("taskStatus", status))
			continue
		}
	}
	// drop zero tail
	lastNonZeroIndex := -1
	for i := maxSeconds; i >= 0; i-- {
		if len(distribution[i]) != 0 {
			lastNonZeroIndex = i
			break
		}
	}
	return distribution[:lastNonZeroIndex+1]
}

func (aa *AppAnalyzer) GetTimeDistribution(condType TaskConditionType) []int {
	tasksDistribution := aa.GetTasksDistribution(condType)
	timeDistribution := make([]int, len(tasksDistribution))
	for index, tasksInCurSecond := range tasksDistribution {
		timeDistribution[index] = len(tasksInCurSecond)
	}
	return timeDistribution
}

func (aa *AppAnalyzer) GetTasksProfiling() profiling.Profiling {
	beginTime := aa.appInfo.AppStatus.CreateTime
	endTime := aa.appInfo.AppStatus.RunningTime
	prof := profiling.NewProfilingWithTime(beginTime)
	for _, ts := range aa.appInfo.TasksStatus {
		prof.StartExecutionWithTime(ts.CreateTime)
		for _, cond := range ts.Conditions {
			prof.AddCheckpointWithTime(string(cond.CondType), cond.TransitionTime)
		}
		//lastCondTime := ts.Conditions[len(ts.Conditions)-1].TransitionTime
		prof.FinishExecutionWithTime(true, endTime)
	}
	return prof
}
