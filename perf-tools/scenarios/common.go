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

package scenarios

import (
	"fmt"
	"time"

	"github.com/apache/yunikorn-release/perf-tools/framework"
	"github.com/apache/yunikorn-release/perf-tools/utils"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

func LoadScenarioConf(conf *framework.Config, scenarioName string, scenarioConf interface{}) error {
	rawScenarioConf := conf.Scenarios[scenarioName]
	if rawScenarioConf == nil {
		return fmt.Errorf("failed to load %s scenario config", scenarioName)
	}
	err := mapstructure.Decode(rawScenarioConf, scenarioConf)
	if err != nil {
		return fmt.Errorf("failed to parse %s scenario config: %s", scenarioName, err.Error())
	}
	utils.Logger.Info("initialized scenario config", zap.String("scenarioName", scenarioName),
		zap.Any("conf", scenarioConf))
	return nil
}

func CleanupApp(appManager framework.AppManager, appInfo *framework.AppInfo, maxWaitTime time.Duration) {
	if appManager != nil && appInfo != nil {
		utils.Logger.Info("make sure app is cleaned up", zap.Any("appID", appInfo.AppID))
		if err := appManager.DeleteWait(appInfo, maxWaitTime); err != nil {
			utils.Logger.Info("failed to cleanup app", zap.Error(err))
		}
	}
}

type RequestConfig struct {
	NumPods          int32
	Repeat           int
	PriorityClass    string
	RequestResources map[string]string
	LimitResources   map[string]string
}

func ConvertToRequestInfos(requestConfigs []*RequestConfig) []*framework.RequestInfo {
	requestInfos := make([]*framework.RequestInfo, 0)
	for _, requestConfig := range requestConfigs {
		for i := 0; i < requestConfig.Repeat; i++ {
			requestInfos = append(requestInfos, framework.NewRequestInfo(requestConfig.NumPods,
				requestConfig.PriorityClass, requestConfig.RequestResources, requestConfig.LimitResources))
		}
	}
	return requestInfos
}
