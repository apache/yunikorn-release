//
// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

module github.com/apache/yunikorn-release

go 1.16

require (
	github.com/TaoYang526/goutils v0.0.0-20210209083921-039008f0a57d
	github.com/apache/yunikorn-core v0.0.0-20220325135453-73d55282f052
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/olekukonko/tablewriter v0.0.0-20170122224234-a0225b3f23b5
	go.uber.org/zap v1.13.0
	golang.org/x/crypto v0.0.0-20220817201139-bc19a97f63c8 // indirect
	golang.org/x/net v0.0.0-20220812174116-3211cb980234 // indirect
	golang.org/x/sys v0.0.0-20220817070843-5a390386f1f2 // indirect
	golang.org/x/tools v0.1.12 // indirect
	gonum.org/v1/plot v0.9.0
	gopkg.in/yaml.v2 v2.2.8
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.20.11
	k8s.io/apimachinery v0.20.11
	k8s.io/client-go v0.20.11
)
