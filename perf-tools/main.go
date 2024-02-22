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

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/apache/yunikorn-release/perf-tools/utils"

	"github.com/apache/yunikorn-release/perf-tools/framework"
	_ "github.com/apache/yunikorn-release/perf-tools/scenarios"
	"go.uber.org/zap"
)

const (
	DateTimeLayout      = "20060102150405"
	ConfigFileName      = "conf.yaml"
	OutputDirNamePrefix = "YK-PERF"
	DefaultLoggingLevel = 0
)

type CommandLineConfig struct {
	ConfigFilePath string
	ScenarioNames  string
}

var commandLineConfig *CommandLineConfig

func init() {
	// test options
	configFile := flag.String("config", "",
		"absolute path to the config file for performance tests")
	scenarioNames := flag.String("scenarios", "",
		"The comma separated names of scenarios which are expected to run")
	flag.Parse()
	commandLineConfig = &CommandLineConfig{
		ConfigFilePath: *configFile,
		ScenarioNames:  *scenarioNames,
	}
}

func main() {
	configFilePath := commandLineConfig.ConfigFilePath
	if configFilePath == "" {
		err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			} else if info.Name() == ConfigFileName {
				configFilePath = path
				return io.EOF
			}
			return nil
		})
		if err != io.EOF {
			utils.Logger.Fatal("failed to find config file", zap.Error(err))
		}
		if configFilePath == "" {
			utils.Logger.Fatal("can't find required config file for integration testing")
		}
	}
	conf, err := framework.InitConfig(configFilePath)
	if err != nil {
		utils.Logger.Fatal("failed to initialize config", zap.Error(err))
	}
	// init kubeClient
	kubeClient, err := utils.NewKubeClient(conf.Common.KubeConfigFile)
	if err != nil {
		utils.Logger.Fatal("failed to initialize kube-client", zap.Error(err))
	}
	// prepare expected test scenarios due to optional flag "scenarios",
	// set to all registered test scenarios if not configured.
	expectedTestScenarios := make([]framework.TestScenario, 0)
	if commandLineConfig.ScenarioNames != "" {
		for _, scenarioName := range strings.Split(commandLineConfig.ScenarioNames, ",") {
			if ts := framework.GetRegisteredTestScenarios()[scenarioName]; ts != nil {
				expectedTestScenarios = append(expectedTestScenarios, ts)
			} else {
				utils.Logger.Fatal("can't find specified scenario",
					zap.String("specifiedScenarioName", scenarioName),
					zap.Any("registeredTestScenarios", framework.GetRegisteredTestScenarios()))
			}
		}
	} else {
		for _, ts := range framework.GetRegisteredTestScenarios() {
			expectedTestScenarios = append(expectedTestScenarios, ts)
		}
	}
	// prepare output directory
	outputTime := time.Now().Format(DateTimeLayout)
	conf.Common.OutputPath = fmt.Sprintf("%s/%s-%s-%s",
		conf.Common.OutputRootPath, OutputDirNamePrefix, commandLineConfig.ScenarioNames, outputTime)
	err = os.Mkdir(conf.Common.OutputPath, os.ModePerm)
	if err != nil {
		utils.Logger.Fatal("failed to create output directory",
			zap.String("outputPath", conf.Common.OutputPath), zap.Error(err))
	}
	// init expected test scenarios first
	for _, testScenario := range expectedTestScenarios {
		err := testScenario.Init(kubeClient, conf)
		if err != nil {
			utils.Logger.Fatal("failed to initialize scenario",
				zap.String("scenarioName", testScenario.GetName()),
				zap.Error(err))
		}
	}
	// run expected test scenarios
	results := utils.NewResults()
	for _, testScenario := range expectedTestScenarios {
		testScenario.Run(results)
	}
	utils.Logger.Info("all tests have been done, generate report")
	results.RefreshStatus()
	fmt.Println(results.String())
}
