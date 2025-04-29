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
	"github.com/apache/yunikorn-release/soak/autoscaler"
	"github.com/apache/yunikorn-release/soak/framework"
	"github.com/apache/yunikorn-release/soak/logger"
	"go.uber.org/zap"
)

const (
	ConfigFileName = "conf.yaml"
)

func main() {
	conf, err := framework.InitConfig(ConfigFileName)
	log := logger.Logger
	if err != nil {
		log.Fatal("failed to parse config", zap.Error(err))
	}
	log.Info("config successfully loaded", zap.Any("conf", conf))

	// Register scenarios
	a := autoscaler.New(conf)
	if a != nil {
		framework.Register(a)
	}

	for _, ts := range framework.GetRegisteredTestScenarios() {
		err = ts.Init()
		if err != nil {
			log.Fatal("failed to initialize scenario", zap.String("scenarioName", ts.GetName()),
				zap.Error(err))
		}

		reportDirs, err := ts.Run()
		if err != nil {
			log.Error("failed to run scenario", zap.String("scenarioName", ts.GetName()),
				zap.Error(err))
		}
		log.Info("Reports are generated for scenario",
			zap.String("scenarioName", ts.GetName()),
			zap.Strings("reportDirectories", reportDirs))

	}

}
