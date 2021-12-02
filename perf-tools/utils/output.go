package utils

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

import (
	"fmt"
	"strings"
)

type VerificationStatus int

const (
	SUCCEEDED VerificationStatus = iota
	FAILED
)

type Reporter interface {
	GenerateReport()
}

type Results struct {
	ScenarioResults []*ScenarioResult
}

type ScenarioResult struct {
	Name          string
	Status        VerificationStatus
	Verifications []*Verification
}

type Verification struct {
	Deep             int
	Name             string
	Status           VerificationStatus
	Description      string
	SubVerifications []*Verification
	Parent           *Verification
}

func NewResults() *Results {
	return &Results{
		ScenarioResults: make([]*ScenarioResult, 0),
	}
}

func (r *Results) CreateScenarioResults(scenarioName string) *ScenarioResult {
	scenarioResult := &ScenarioResult{
		Name:          scenarioName,
		Status:        SUCCEEDED,
		Verifications: make([]*Verification, 0),
	}
	r.ScenarioResults = append(r.ScenarioResults, scenarioResult)
	return scenarioResult
}

func (sr *ScenarioResult) AddVerification(name, description string, status VerificationStatus) {
	verification := &Verification{
		Deep:        1,
		Name:        name,
		Description: description,
		Status:      status,
	}
	sr.Verifications = append(sr.Verifications, verification)
	if status == FAILED {
		sr.Status = FAILED
	}
}

func (sr *ScenarioResult) AddVerificationGroup(name, description string) *Verification {
	verification := &Verification{
		Deep:             1,
		Name:             name,
		Description:      description,
		SubVerifications: make([]*Verification, 0),
	}
	sr.Verifications = append(sr.Verifications, verification)
	return verification
}

func (vg *Verification) AddSubVerificationGroup(name, description string) *Verification {
	subVerification := &Verification{
		Deep:        vg.Deep + 1,
		Name:        name,
		Description: description,
		Parent:      vg,
	}
	vg.SubVerifications = append(vg.SubVerifications, subVerification)
	return subVerification
}

func (vg *Verification) AddAssertSubVerification(result bool, name, description string) *Verification {
	var status VerificationStatus
	if result {
		status = SUCCEEDED
	} else {
		status = FAILED
	}
	return vg.AddSubVerification(name, description, status)
}

func (vg *Verification) AddErrorSubVerification(err error, name, description string) *Verification {
	var status VerificationStatus
	if err == nil {
		status = SUCCEEDED
	} else {
		status = FAILED
		if description != "" {
			description += ","
		}
		description += " err: " + strings.TrimSpace(err.Error())
	}
	return vg.AddSubVerification(name, description, status)
}

func (vg *Verification) AddSubVerification(name, description string, status VerificationStatus) *Verification {
	subVerification := &Verification{
		Deep:        vg.Deep + 1,
		Name:        name,
		Status:      status,
		Description: description,
		Parent:      vg,
	}
	vg.SubVerifications = append(vg.SubVerifications, subVerification)
	if status == FAILED {
		parentVer := subVerification.Parent
		for parentVer != nil {
			parentVer.Status = FAILED
			parentVer = parentVer.Parent
		}
	}
	//Logger.Info(getVerificationStatusInfo(subVerification))
	return subVerification
}

func (vg *Verification) IsFailed() bool {
	return vg.Status == FAILED
}

func (r *Results) RefreshStatus() {
	for _, scenarioResult := range r.ScenarioResults {
		status := SUCCEEDED
		for _, v := range scenarioResult.Verifications {
			v.RefreshStatus()
			if v.Status == FAILED {
				status = FAILED
			}
		}
		scenarioResult.Status = status
	}
}

func (v *Verification) RefreshStatus() {
	if len(v.SubVerifications) > 0 {
		status := SUCCEEDED
		for _, subV := range v.SubVerifications {
			subV.RefreshStatus()
			if subV.Status == FAILED {
				status = FAILED
			}
		}
		v.Status = status
	}
}

func (r *Results) String() string {
	outputInfo := ""
	for _, scenarioResult := range r.ScenarioResults {
		outputInfo += fmt.Sprintf("%s\n", getStatusInfo("Scenario: "+scenarioResult.Name, scenarioResult.Status))
		for _, v := range scenarioResult.Verifications {
			outputInfo += v.String()
		}
	}
	return outputInfo
}

func (v *Verification) String() string {
	outputInfo := fmt.Sprintf("%s%s\n", getDeepPrefix(v.Deep), getVerificationStatusInfo(v))
	if len(v.SubVerifications) > 0 {
		for _, subV := range v.SubVerifications {
			outputInfo += subV.String()
		}
	}
	return outputInfo
}

func getStatusInfo(name string, status VerificationStatus) string {
	statusStr := ""
	switch status {
	case FAILED:
		statusStr = "FAILED"
	case SUCCEEDED:
		statusStr = "SUCCEEDED"
	}
	return fmt.Sprintf("%s [%s]", name, statusStr)
}

func getVerificationStatusInfo(v *Verification) string {
	statusInfo := getStatusInfo(v.Name, v.Status)
	if v.Description != "" {
		statusInfo += fmt.Sprintf(" (%s)", v.Description)
	}
	return statusInfo
}

func getDeepPrefix(deep int) string {
	return strings.Repeat("    ", deep)
}
