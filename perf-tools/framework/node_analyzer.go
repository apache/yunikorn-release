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

	"github.com/apache/yunikorn-core/pkg/common/resources"
	"github.com/apache/yunikorn-release/perf-tools/utils"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeAnalyzer struct {
	kubeClient       *utils.KubeClient
	allocatableNodes map[string]*NodeInfo
	nodeSelector     string
}

func NewNodeAnalyzer(kubeClient *utils.KubeClient, nodeSelector string) *NodeAnalyzer {
	return &NodeAnalyzer{
		kubeClient:   kubeClient,
		nodeSelector: nodeSelector,
	}
}

// InitNodeInfosBeforeTesting records the snapshot of schedulable and ready nodes before testing
func (na *NodeAnalyzer) InitNodeInfosBeforeTesting() error {
	// get resource map for schedulable nodes
	nodes, err := na.kubeClient.GetNodes(&metav1.ListOptions{LabelSelector: na.nodeSelector})
	if nodes == nil || err != nil {
		return err
	}
	if len(nodes.Items) == 0 {
		return fmt.Errorf("no selected nodes in k8s cluster, nodeSelector=%s", na.nodeSelector)
	}
	na.allocatableNodes = make(map[string]*NodeInfo)
	for _, nodeItem := range nodes.Items {
		if nodeItem.Spec.Unschedulable {
			utils.Logger.Debug("skip unschedulable node", zap.String("nodeID", nodeItem.Name))
			continue
		}
		if !IsNodeReady(&nodeItem) {
			utils.Logger.Debug("skip not-ready node", zap.String("nodeID", nodeItem.Name))
			continue
		}
		// allocatable resource is the upper limit of allocated resource for pods
		capacityRes := ParseResourceFromResourceList(&nodeItem.Status.Allocatable)
		nodeInfo := NewNodeInfo(nodeItem.Name, capacityRes, resources.NewResource())
		na.allocatableNodes[nodeItem.Name] = nodeInfo
	}
	return nil
}

// CalculateAllocatedResource calculate allocated resource for nodes,
// which may be rather time-consuming when there are numerous pods in the cluster,
// so this should be called only if necessary!
func (na *NodeAnalyzer) CalculateAllocatedResource() {
	utils.Logger.Info("start loading all pods for calculating allocated resource")
	podList, _ := na.kubeClient.GetPods("", utils.GetEverythingListOptions())
	utils.Logger.Info(fmt.Sprintf("loaded %d pods", len(podList.Items)))
	for _, pod := range podList.Items {
		if pod.Spec.NodeName == "" {
			continue
		}
		if node, ok := na.allocatableNodes[pod.Spec.NodeName]; ok {
			node.AllocatedResource.NodeResourceBefore.AddTo(GetPodRequestResource(&pod))
		}
	}
	utils.Logger.Info("calculated allocated resource successfully")
}

func (na *NodeAnalyzer) ClearApps() {
	for _, nodeInfo := range na.allocatableNodes {
		nodeInfo.ClearTasks()
	}
}

// AnalyzeApp updates the state of nodes according to app status
func (na *NodeAnalyzer) AnalyzeApp(appInfo *AppInfo) {
	for _, taskStatus := range appInfo.TasksStatus {
		if nodeInfo, ok := na.allocatableNodes[taskStatus.NodeID]; ok {
			nodeInfo.AddTask(taskStatus)
		}
	}
}

func (na *NodeAnalyzer) GetAllocatableNodes() map[string]*NodeInfo {
	return na.allocatableNodes
}

func (na *NodeAnalyzer) GetScheduledNodes() map[string]*NodeInfo {
	relatedNodeInfos := make(map[string]*NodeInfo)
	for nodeID, nodeInfo := range na.allocatableNodes {
		if len(nodeInfo.Tasks) > 0 {
			relatedNodeInfos[nodeID] = nodeInfo
		}
	}
	return relatedNodeInfos
}

func (na *NodeAnalyzer) GetTotalAllocatableResource() *resources.Resource {
	totalAllocatableResource := resources.NewResource()
	for _, nodeInfo := range na.allocatableNodes {
		nodeAllocatedRes := nodeInfo.AllocatedResource.GetResource()
		nodeAllocatableRes := resources.Sub(nodeInfo.Capacity, nodeAllocatedRes)
		totalAllocatableResource.AddTo(nodeAllocatableRes)
	}
	return totalAllocatableResource
}

func (na *NodeAnalyzer) GetNodeResourceDistribution(tasksDistribution [][]*TaskStatus,
	resourceName string) [10][]int {
	// init node resources, key(string): <NodeID>, value([]int64): <CapacityResourceValue>, <AllocatedResourceValue>
	nodeResources := make(map[string][]int64)
	for nodeID, nodeInfo := range na.allocatableNodes {
		nodeResources[nodeID] = []int64{int64(nodeInfo.Capacity.Resources[resourceName]),
			int64(nodeInfo.AllocatedResource.GetResource().Resources[resourceName])}
	}
	// statistic the number of nodes in 10 buckets with different resource utilization levels
	// ([0%,10%), [10%,20%), ..., [90%,100%]) in every second
	var buckets [10][]int
	for _, tasksInCurSecond := range tasksDistribution {
		for _, taskStatus := range tasksInCurSecond {
			nodeResources[taskStatus.NodeID][1] += int64(taskStatus.RequestResources.Resources[resourceName])
		}
		// calculate current distribution
		bucketsInThisSecond := calculateResourceDistribution(nodeResources)
		for level, num := range bucketsInThisSecond {
			buckets[level] = append(buckets[level], num)
		}
	}
	return buckets
}

func calculateResourceDistribution(resources map[string][]int64) [10]int {
	var buckets [10]int
	for _, resourceValues := range resources {
		if len(resourceValues) == 2 {
			ratio := float32(resourceValues[1]*10) / float32(resourceValues[0])
			bucketIndex := int(ratio)
			if bucketIndex == 10 {
				bucketIndex = 9
			}
			buckets[bucketIndex] += 1
		} else {
			panic(fmt.Errorf("Error resources: %+v ", resourceValues))
		}
	}
	return buckets
}

func ParseResourceFromResourceList(resourceList *v1.ResourceList) *resources.Resource {
	resourceMap := make(map[string]resources.Quantity)
	for name, value := range *resourceList {
		switch name {
		case v1.ResourceMemory:
			resourceMap[resources.MEMORY] = resources.Quantity(value.ScaledValue(resource.Mega))
		case v1.ResourceCPU:
			resourceMap[resources.VCORE] = resources.Quantity(value.MilliValue())
		default:
			resourceMap[name.String()] = resources.Quantity(value.Value())
		}
	}
	return resources.NewResourceFromMap(resourceMap)
}

// IsNodeReady returns true if a node is ready, false otherwise.
func IsNodeReady(node *v1.Node) bool {
	for _, c := range node.Status.Conditions {
		if c.Type == v1.NodeReady {
			return c.Status == v1.ConditionTrue
		}
	}
	return false
}

func GetPodRequestResource(pod *v1.Pod) *resources.Resource {
	resource := resources.NewResource()
	for _, container := range pod.Spec.Containers {
		resource.AddTo(ParseResourceFromResourceList(&container.Resources.Requests))
	}
	return resource
}
