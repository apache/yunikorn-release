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
)

type NodeInfo struct {
	NodeID            string
	Tasks             map[string]*TaskStatus
	Capacity          *resources.Resource
	AllocatedResource ResourceInfo
}

type ResourceInfo struct {
	NodeResourceBefore *resources.Resource
	TasksTotalResource *resources.Resource
	NumTasks           int
}

func NewNodeInfo(nodeID string, capacityRes *resources.Resource, allocatedRes *resources.Resource) *NodeInfo {
	return &NodeInfo{
		NodeID:   nodeID,
		Tasks:    make(map[string]*TaskStatus),
		Capacity: capacityRes,
		AllocatedResource: ResourceInfo{
			NodeResourceBefore: allocatedRes,
			TasksTotalResource: resources.NewResource(),
		},
	}
}

func (ni *NodeInfo) AddTask(task *TaskStatus) {
	ni.Tasks[task.TaskID] = task
	ni.AllocatedResource.AddTaskResource(task.RequestResources)
}

func (ni *NodeInfo) ClearTasks() {
	ni.Tasks = make(map[string]*TaskStatus)
	ni.AllocatedResource.ClearTaskResources()
}

func (ni *NodeInfo) GetSummary() string {
	return fmt.Sprintf("nodeID=%s, numTasks=%d, capacity=%+v, allocatedRes=%+v",
		ni.NodeID, len(ni.Tasks), ni.Capacity, ni.AllocatedResource)
}

func (ri *ResourceInfo) AddTaskResource(taskReqResources *resources.Resource) {
	if ri.TasksTotalResource == nil {
		ri.TasksTotalResource = resources.NewResource()
	}
	ri.TasksTotalResource.AddTo(taskReqResources)
	ri.NumTasks++
}

func (ri *ResourceInfo) ClearTaskResources() {
	ri.NumTasks = 0
	ri.TasksTotalResource = resources.NewResource()
}

func (ri *ResourceInfo) GetResource() *resources.Resource {
	resource := ri.NodeResourceBefore.Clone()
	resource.AddTo(ri.TasksTotalResource)
	return resource
}
