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
	"gopkg.in/yaml.v3"
	"os"
)

type KubeconfigFields struct {
	Path    string `yaml:"path,omitempty"`
	Context string `yaml:"context,omitempty"`
}

type NodeFields struct {
	Path         string `yaml:"path,omitempty"`
	MaxCount     string `yaml:"maxCount,omitempty"`
	DesiredCount string `yaml:"desiredCount,omitempty"`
}

type JobFields struct {
	Path     string `yaml:"path,omitempty"`
	Count    string `yaml:"count,omitempty"`
	PodCount string `yaml:"podCount,omitempty"`
}

type SchedulerFields struct {
	Path           string `yaml:"path,omitempty"`
	VcoreRequests  string `yaml:"vcoreRequests,omitempty"`
	VcoreLimits    string `yaml:"vcoreLimits,omitempty"`
	MemoryRequests string `yaml:"memoryRequests,omitempty"`
	MemoryLimits   string `yaml:"memoryLimits,omitempty"`
}

type ChaosFields struct {
	Path  string `yaml:"path,omitempty"`
	Count string `yaml:"count,omitempty"`
}

type Template struct {
	Kubeconfig KubeconfigFields  `yaml:"kubeconfig,omitempty"`
	Node       []NodeFields      `yaml:"node,omitempty"`
	Job        []JobFields       `yaml:"job,omitempty"`
	Scheduler  []SchedulerFields `yaml:"scheduler,omitempty"`
	Chaos      []ChaosFields     `yaml:"chaos,omitempty"`
}

type TestCaseParams map[string]interface{}

type Prom struct {
	Query      string `yaml:"query,omitempty"`
	Expression string `yaml:"expression,omitempty"`
	Value      string `yaml:"value,omitempty"`
	Op         string `yaml:"op,omitempty"`
}

type Metrics struct {
	SchedulerRestarts  int    `yaml:"schedulerRestarts,omitempty"`
	MaxAllocationDelay string `yaml:"maxAllocationDelay,omitempty"`
	Prom               []Prom `yaml:"prom,omitempty"`
}

type Threshold struct {
	MaxRuntime     string  `yaml:"maxRuntime,omitempty"`
	PendingPods    int     `yaml:"pendingPods,omitempty"`
	DetectDeadlock bool    `yaml:"detectDeadlock,omitempty"`
	Metrics        Metrics `yaml:"metrics,omitempty"`
}

type TestCase struct {
	Name                    string         `yaml:"name,omitempty"`
	Params                  TestCaseParams `yaml:"params,omitempty"`
	Schedule                string         `yaml:"schedule,omitempty"`
	Runs                    int            `yaml:"runs,omitempty"`
	Labels                  []string       `yaml:"labels,omitempty"`
	Threshold               Threshold      `yaml:"threshold,omitempty"`
	ClusterLoaderConfigPath string         `yaml:"clusterLoaderConfigPath,omitempty"`
}

type Test struct {
	Name      string     `yaml:"name,omitempty"`
	Template  Template   `yaml:"template,omitempty"`
	TestCases []TestCase `yaml:"testCases,omitempty"`
}

type Config struct {
	Tests []Test `yaml:"tests,omitempty"`
}

func InitConfig(configFile string) (*Config, error) {
	yamlContent, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read config file: %s ", err.Error())
	}
	conf := Config{}
	err = yaml.Unmarshal(yamlContent, &conf)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse config file: %s ", err.Error())
	}
	return &conf, nil
}
