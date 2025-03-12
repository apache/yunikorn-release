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

package setup

import (
	"github.com/apache/yunikorn-release/soak/framework"
	"go.uber.org/zap"
	"testing"
)

func TestSetAutoScalerPerConfig(t *testing.T) {
	conf, err := framework.InitConfig("test_conf.yaml")
	if err != nil {
		logger.Fatal("failed to parse config", zap.Error(err))
	}
	logger.Info("config successfully loaded", zap.Any("conf", conf))

	for _, test := range conf.Tests {
		if len(test.Template.Node) > 0 {
			for _, nodeTemplate := range test.Template.Node {
				err := setAutoscalerPerConfig(nodeTemplate)
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}
