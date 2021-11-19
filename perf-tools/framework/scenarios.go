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
	"github.com/apache/incubator-yunikorn-release/perf-tools/utils"
	"go.uber.org/zap"
)

type TestScenarios struct {
	registeredTestScenarios map[string]TestScenario
}

var testScenarios TestScenarios

func init() {
	testScenarios.registeredTestScenarios = make(map[string]TestScenario)
}

func Register(ts TestScenario) {
	testScenarios.registeredTestScenarios[ts.GetName()] = ts
	utils.Logger.Info("register scenario", zap.String("scenarioName", ts.GetName()))
}

func GetRegisteredTestScenarios() map[string]TestScenario {
	return testScenarios.registeredTestScenarios
}

type TestScenario interface {
	GetName() string
	Init(kubeClient *utils.KubeClient, config *Config) error
	Run(results *utils.Results)
}
