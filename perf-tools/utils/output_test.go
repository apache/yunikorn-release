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

package utils

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestResults(t *testing.T) {
	results := NewResults()
	s1 := results.CreateScenarioResults("s1")
	s1.AddVerification("s1-v1", "des...", SUCCEEDED)
	s2 := results.CreateScenarioResults("s2")
	s2.AddVerification("s2-v1", "des...", SUCCEEDED)
	s2.AddVerification("s2-v2", "des...", SUCCEEDED)
	s2vg1 := s2.AddVerificationGroup("s2-vg1", "")
	s2vg1.AddSubVerification("s2-vg1-1", "des...", SUCCEEDED)
	s2vg1.AddSubVerification("s2-vg1-2", "des...", SUCCEEDED)
	s2vg2 := s2.AddVerificationGroup("s2-vg2", "")
	s2vg2.AddSubVerification("s2-vg2-1", "des...", SUCCEEDED)
	s2vg2.AddSubVerification("s2-vg2-2", "des...", FAILED)
	s2vg2.AddSubVerification("s2-vg2-3", "des...", SUCCEEDED)
	s3 := results.CreateScenarioResults("s3")
	s3vg1 := s3.AddVerificationGroup("s3-vg1", "")
	s3vg1sub1 := s3vg1.AddSubVerificationGroup("s3-vg1-1", "")
	s3vg1sub1.AddSubVerification("s3-vg1-1-1", "", FAILED)

	results.RefreshStatus()
	assert.Equal(t, s2vg2.Status, FAILED)
	assert.Equal(t, s2.Status, FAILED)
	assert.Equal(t, s3.Status, FAILED)
	assert.Equal(t, s3vg1.Status, FAILED)
	assert.Equal(t, s3vg1sub1.Status, FAILED)

	t.Log("\n" + results.String())
}
