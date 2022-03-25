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
	"time"

	"github.com/apache/yunikorn-core/pkg/common/resources"

	apiv1 "k8s.io/api/core/v1"
)

type AppInfo struct {
	Namespace       string
	AppID           string
	Queue           string
	RequestInfos    []*RequestInfo
	PodSpec         apiv1.PodSpec
	PodTemplateSpec apiv1.PodTemplateSpec
	AppStatus       AppStatus
	TasksStatus     map[string]*TaskStatus
}

type RequestInfo struct {
	Number           int32
	PriorityClass    string
	RequestResources map[string]string
	LimitResources   map[string]string
}

type AppStatus struct {
	CreateTime time.Time
	// running time of the last task
	RunningTime time.Time
	DesiredNum  int
	CreatedNum  int
	ReadyNum    int
}

type TaskStatus struct {
	TaskID           string
	CreateTime       time.Time
	RunningTime      time.Time
	NodeID           string
	RequestResources *resources.Resource
	Conditions       []*TaskCondition
}

type TaskCondition struct {
	CondType       TaskConditionType
	TransitionTime time.Time
}

type TasksDistributionInfo struct {
	LeastNum        int
	LeastNumNodeID  string
	MostNum         int
	MostNumNodeID   string
	AvgNum          float64
	SortedNodeInfos []*NodeInfo
}

type TaskConditionType string

const (
	PodCreated      TaskConditionType = "PodCreated"
	PodScheduled    TaskConditionType = "PodScheduled"
	PodStarted      TaskConditionType = "PodStarted"
	PodInitialized  TaskConditionType = "Initialized"
	PodReady        TaskConditionType = "Ready"
	ContainersReady TaskConditionType = "ContainersReady"
)

func NewRequestInfo(number int32, priorityClass string, requestResources, limitResources map[string]string) *RequestInfo {
	return &RequestInfo{
		Number:           number,
		PriorityClass:    priorityClass,
		RequestResources: requestResources,
		LimitResources:   limitResources,
	}
}

func NewAppInfo(namespace, appID, queue string, requestInfos []*RequestInfo,
	podTemplateSpec apiv1.PodTemplateSpec, podSpec apiv1.PodSpec) *AppInfo {
	return &AppInfo{
		Namespace:       namespace,
		AppID:           appID,
		Queue:           queue,
		RequestInfos:    requestInfos,
		PodTemplateSpec: podTemplateSpec,
		PodSpec:         podSpec,
	}
}

func NewTaskStatus(taskID, nodeID string, createTime, runningTime time.Time,
	requestResources *resources.Resource, conditions []*TaskCondition) *TaskStatus {
	return &TaskStatus{
		TaskID:           taskID,
		CreateTime:       createTime,
		RunningTime:      runningTime,
		NodeID:           nodeID,
		RequestResources: requestResources,
		Conditions:       conditions,
	}
}

func (appInfo *AppInfo) SetAppStatus(desiredNum, createdNum, readyNum int) {
	appInfo.AppStatus.DesiredNum = desiredNum
	appInfo.AppStatus.CreatedNum = createdNum
	appInfo.AppStatus.ReadyNum = readyNum
}

func (appInfo *AppInfo) GetDesiredNumTasks() int32 {
	var desiredNum int32
	for _, requestInfo := range appInfo.RequestInfos {
		desiredNum += requestInfo.Number
	}
	return desiredNum
}

func (taskStatus *TaskStatus) GetTransitionTime(condType TaskConditionType) *time.Time {
	for _, cond := range taskStatus.Conditions {
		if cond.CondType == condType {
			return &cond.TransitionTime
		}
	}
	return nil
}

func GetOrderedTaskConditionTypes() []TaskConditionType {
	return []TaskConditionType{PodCreated, PodScheduled,
		PodStarted, PodInitialized, PodReady, ContainersReady}
}
