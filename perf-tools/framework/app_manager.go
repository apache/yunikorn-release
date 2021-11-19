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
	"fmt"
	"regexp"
	"time"

	"github.com/apache/incubator-yunikorn-release/perf-tools/constants"
	"github.com/apache/incubator-yunikorn-release/perf-tools/utils"

	"go.uber.org/zap"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AppManager interface {
	Create(schedulerName string, appInfo *AppInfo) error
	Delete(appInfo *AppInfo) error
	RefreshAppStatus(appInfo *AppInfo) error
	RefreshTasksStatusAfterRunning(appInfo *AppInfo) error
	WaitForAppsToBeCleanedUp(appInfos *AppInfo, timeout time.Duration) error
	WaitForAppsToBeSatisfied(appInfos *AppInfo, timeout time.Duration) error
	// create an app and wait for it to be running, refresh tasks status at last
	CreateWaitAndRefreshTasksStatus(schedulerName string, appInfo *AppInfo, timeout time.Duration) error
	// delete an app and wait for it to be cleaned up
	DeleteWait(appInfo *AppInfo, timeout time.Duration) error
}

type DeploymentsAppManager struct {
	kubeClient *utils.KubeClient
	nameRegexp *regexp.Regexp
}

func NewDeploymentsAppManager(kubeClient *utils.KubeClient) AppManager {
	regexp, _ := regexp.Compile(`[_\W]`)
	return &DeploymentsAppManager{
		kubeClient: kubeClient,
		nameRegexp: regexp,
	}
}

func (dam *DeploymentsAppManager) Create(schedulerName string, appInfo *AppInfo) error {
	if len(appInfo.RequestInfos) == 0 {
		return fmt.Errorf("request info not defined for app %s", appInfo.AppID)
	}
	for reqIndex, requestInfo := range appInfo.RequestInfos {
		// init container
		var container apiv1.Container
		if len(appInfo.PodSpec.Containers) > 0 {
			container = appInfo.PodSpec.Containers[0]
		} else {
			container = apiv1.Container{}
			container.Name = constants.DefaultContainerName
			container.Image = constants.DefaultContainerImage
		}
		if requestInfo.RequestResources != nil {
			if container.Resources.Requests == nil {
				container.Resources.Requests = apiv1.ResourceList{}
			}
			for resourceName, resourceValue := range requestInfo.RequestResources {
				container.Resources.Requests[apiv1.ResourceName(resourceName)] = resource.MustParse(resourceValue)
			}
		}
		if requestInfo.LimitResources != nil {
			if container.Resources.Limits == nil {
				container.Resources.Limits = apiv1.ResourceList{}
			}
			for resourceName, resourceValue := range requestInfo.LimitResources {
				container.Resources.Limits[apiv1.ResourceName(resourceName)] = resource.MustParse(resourceValue)
			}
		}
		// init and create deployment
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: appInfo.Namespace,
				Name:      dam.getDeploymentName(appInfo, reqIndex),
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &requestInfo.Number,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						constants.LabelAppID: appInfo.AppID,
					},
				},
				Template: apiv1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							constants.LabelAppID: appInfo.AppID,
							constants.LabelQueue: appInfo.Queue,
						},
						Annotations: appInfo.PodTemplateSpec.ObjectMeta.Annotations,
					},
					Spec: apiv1.PodSpec{
						SchedulerName: schedulerName,
						HostNetwork:   appInfo.PodSpec.HostNetwork,
						Containers: []apiv1.Container{
							container,
						},
						Tolerations:       appInfo.PodSpec.Tolerations,
						NodeSelector:      appInfo.PodSpec.NodeSelector,
						PriorityClassName: requestInfo.PriorityClass,
					},
				},
			},
		}
		err := dam.kubeClient.CreateDeployment(appInfo.Namespace, deployment)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dam *DeploymentsAppManager) getDeploymentName(appInfo *AppInfo, reqIndex int) string {
	normalizedName := dam.nameRegexp.ReplaceAllString(appInfo.AppID, "-")
	return fmt.Sprintf("%s-%d", normalizedName, reqIndex)
}

func (dam *DeploymentsAppManager) Delete(appInfo *AppInfo) error {
	for i := 0; i < len(appInfo.RequestInfos); i++ {
		err := dam.kubeClient.DeleteDeployment(appInfo.Namespace, dam.getDeploymentName(appInfo, i))
		if err != nil {
			return err
		}
	}
	return nil
}

func (dam *DeploymentsAppManager) RefreshAppStatus(appInfo *AppInfo) error {
	var summaryMetrics [3]int
	firstCreateTime := time.Time{}
	for i := 0; i < len(appInfo.RequestInfos); i++ {
		createTime, metrics, err := dam.kubeClient.GetDeploymentInfo(
			appInfo.Namespace, dam.getDeploymentName(appInfo, i))
		if err != nil {
			utils.Logger.Info("failed to refresh app status", zap.Error(err))
			continue
		}
		if firstCreateTime.IsZero() || firstCreateTime.After(createTime) {
			firstCreateTime = createTime
		}
		summaryMetrics[0] += metrics[0]
		summaryMetrics[1] += metrics[1]
		summaryMetrics[2] += metrics[2]
	}
	appInfo.SetAppStatus(summaryMetrics[0], summaryMetrics[1], summaryMetrics[2])
	return nil
}

func (dam *DeploymentsAppManager) RefreshTasksStatusAfterRunning(appInfo *AppInfo) error {
	podList, err := dam.kubeClient.GetPods(appInfo.Namespace,
		utils.GetListOptions(map[string]string{constants.LabelAppID: appInfo.AppID}))
	if err != nil {
		return err
	}
	tasksStatus := make(map[string]*TaskStatus)
	maxRunningTime := time.Time{}
	firstCreateTime := time.Time{}
	for _, pod := range podList.Items {
		createTime := pod.CreationTimestamp.Time
		startTime := pod.Status.StartTime.Time
		requestResources := ParseResourceFromResourceList(&pod.Spec.Containers[0].Resources.Requests)
		// init conditions map
		condMap := make(map[TaskConditionType]*TaskCondition)
		condMap[PodCreated] = &TaskCondition{
			CondType:       PodCreated,
			TransitionTime: pod.CreationTimestamp.Time,
		}
		condMap[PodStarted] = &TaskCondition{
			CondType:       PodStarted,
			TransitionTime: startTime,
		}
		for _, cond := range pod.Status.Conditions {
			condMap[TaskConditionType(cond.Type)] = &TaskCondition{
				CondType:       TaskConditionType(cond.Type),
				TransitionTime: cond.LastTransitionTime.Time,
			}
		}

		// set running time from the last condition (ContainersReady)
		var runningTime time.Time
		if readyCond, ok := condMap[ContainersReady]; ok {
			runningTime = readyCond.TransitionTime
		} else {
			utils.Logger.Fatal("unexpected conditions", zap.Any("Conditions", condMap))
		}
		// transfer to ordered conditions
		orderedCondTypes := GetOrderedTaskConditionTypes()
		conditions := make([]*TaskCondition, len(orderedCondTypes))
		for idx, condType := range orderedCondTypes {
			if cond, ok := condMap[condType]; ok {
				conditions[idx] = cond
			} else {
				utils.Logger.Fatal("unknown condition", zap.Any("condType", condType),
					zap.Any("podName", pod.Name))
			}
		}
		taskStatus := NewTaskStatus(pod.Name, pod.Spec.NodeName,
			createTime, runningTime, requestResources, conditions)
		tasksStatus[taskStatus.TaskID] = taskStatus
		// update maxRunningTime
		if runningTime.After(maxRunningTime) {
			maxRunningTime = runningTime
		}
		if firstCreateTime.IsZero() || createTime.Before(firstCreateTime) {
			firstCreateTime = createTime
		}
	}
	appInfo.TasksStatus = tasksStatus
	appInfo.AppStatus.RunningTime = maxRunningTime
	appInfo.AppStatus.CreateTime = firstCreateTime
	return nil
}

func (dam *DeploymentsAppManager) WaitForAppsToBeCleanedUp(appInfo *AppInfo, timeout time.Duration) error {
	startTime := time.Now()
	i := 1
	return waitForCondition(func() bool {
		err := dam.RefreshAppStatus(appInfo)
		if err != nil {
			return true
		}
		if appInfo.AppStatus.DesiredNum != 0 || appInfo.AppStatus.CreatedNum != 0 || appInfo.AppStatus.ReadyNum != 0 {
			if time.Since(startTime) > 60*time.Duration(i)*time.Second {
				utils.Logger.Info("still waiting for app to be cleaned up",
					zap.String("appID", appInfo.AppID),
					zap.Duration("timeout", timeout),
					zap.Duration("elapseTime", time.Since(startTime)),
					zap.Int("desiredNum", appInfo.AppStatus.DesiredNum),
					zap.Int("createdNum", appInfo.AppStatus.CreatedNum),
					zap.Int("readyNum", appInfo.AppStatus.ReadyNum))
				i++
			}
			return false
		}
		utils.Logger.Info("app is cleaned up", zap.String("appID", appInfo.AppID),
			zap.Any("appStatus", appInfo.AppStatus))
		return true
	}, 1*time.Second, timeout)
}

func (dam *DeploymentsAppManager) WaitForAppsToBeSatisfied(appInfo *AppInfo, timeout time.Duration) error {
	startTime := time.Now()
	i := 1
	return waitForCondition(func() bool {
		err := dam.RefreshAppStatus(appInfo)
		if err != nil {
			return true
		}
		if appInfo.AppStatus.DesiredNum == 0 || appInfo.AppStatus.DesiredNum != appInfo.AppStatus.ReadyNum {
			if time.Since(startTime) > 5*time.Duration(i)*time.Second {
				utils.Logger.Info("still waiting for app to be running",
					zap.String("appID", appInfo.AppID),
					zap.Duration("timeout", timeout),
					zap.Duration("elapseTime", time.Since(startTime)),
					zap.Int("desiredNum", appInfo.AppStatus.DesiredNum),
					zap.Int("readyNum", appInfo.AppStatus.ReadyNum))
				i++
			}
			return false
		}
		return true
	}, 1*time.Second, timeout)
}

func (dam *DeploymentsAppManager) CreateWaitAndRefreshTasksStatus(schedulerName string, appInfo *AppInfo,
	timeout time.Duration) error {
	err := dam.Create(schedulerName, appInfo)
	if err != nil {
		return fmt.Errorf("failed to create app: %s", err.Error())
	}
	// wait for this app to be running (all pods are scheduled to be running)
	err = dam.WaitForAppsToBeSatisfied(appInfo, timeout)
	if err != nil {
		return fmt.Errorf("failed to wait for this app to be running: %s", err.Error())
	}
	// refresh task status
	err = dam.RefreshTasksStatusAfterRunning(appInfo)
	if err != nil {
		return fmt.Errorf("failed to refresh task status: %s", err.Error())
	}
	return nil
}

func (dam *DeploymentsAppManager) DeleteWait(appInfo *AppInfo,
	timeout time.Duration) error {
	err := dam.Delete(appInfo)
	if err != nil {
		return fmt.Errorf("failed to delete app: %s", err.Error())
	}
	// wait for this app to be cleaned up
	err = dam.WaitForAppsToBeCleanedUp(appInfo, timeout)
	if err != nil {
		return fmt.Errorf("failed to wait for this app to be cleaned up: %s", err.Error())
	}
	return nil
}

// copied from yunikorn-k8shim to avoid importing too many dependencies
func waitForCondition(eval func() bool, interval time.Duration, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if eval() {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for condition")
		}

		time.Sleep(interval)
	}
}
