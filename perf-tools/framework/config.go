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
	"io/ioutil"

	apiv1 "k8s.io/api/core/v1"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Common    *CommonConfig
	Scenarios map[string]interface{}
}

type CommonConfig struct {
	KubeConfigFile  string
	SchedulerName   string
	MaxWaitSeconds  int
	Queue           string
	Namespace       string
	OutputRootPath  string
	OutputPath      string
	NodeSelector    string
	PodSpec         apiv1.PodSpec
	PodTemplateSpec apiv1.PodTemplateSpec
}

func InitConfig(configFile string) (*Config, error) {
	yamlContent, err := ioutil.ReadFile(configFile)
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
