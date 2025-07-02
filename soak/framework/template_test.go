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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplateProcessor(t *testing.T) {
	// Create a temporary template file using string literals for parameter names
	templateContent := `{{$totalPods := DefaultParam "CL2_SCHEDULER_THROUGHPUT_PODS" 10}}
{{$defaultQps := DefaultParam "CL2_DEFAULTQPS" 5}}
{{$threshold := DefaultParam "CL2_SCHEDULERTHROUGHPUTTHRESHOLD" 15}}

name: test-config
totalPods: {{$totalPods}}
qps: {{$defaultQps}}
threshold: {{$threshold}}
calculatedValue: {{Add $totalPods 100}}
`

	tempDir, err := os.MkdirTemp("", "template_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	templatePath := filepath.Join(tempDir, "test_template.yaml")
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// With parameters
	testCase := TestCase{
		Name: "test-case",
		Params: TestCaseParams{
			"numPods":                       5000,
			"defaultQps":                    10,
			"schedulerThroughputThreshold":  20,
		},
	}

	params := BuildParameterMap(testCase)

	processor := NewTemplateProcessor(params)

	result, err := processor.ProcessConfigFile(templatePath)
	if err != nil {
		t.Fatalf("Failed to process template: %v", err)
	}

	// Verify results
	expectedValues := []string{
		"totalPods: 5000",    // Should use numPods parameter
		"qps: 10",            // Should use defaultQps parameter
		"threshold: 20",      // Should use schedulerThroughputThreshold parameter
		"calculatedValue: 5100", // Should be numPods + 100
	}

	for _, expected := range expectedValues {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected result to contain '%s', but got:\n%s", expected, result)
		}
	}
}

func TestBuildParameterMap(t *testing.T) {
	testCase := TestCase{
		Name: "test-case",
		Params: TestCaseParams{
			"nodesMaxCount":     1000,
			"nodesDesiredCount": 20,
			"numPods":           5000,
			"customParam":       "test-value",
		},
	}

	params := BuildParameterMap(testCase)

	// Original parameters must be preserved
	if params["nodesMaxCount"] != 1000 {
		t.Errorf("Expected nodesMaxCount to be 1000, got %v", params["nodesMaxCount"])
	}

	if params["customParam"] != "test-value" {
		t.Errorf("Expected customParam to be 'test-value', got %v", params["customParam"])
	}

	// CL2 parameters must be created
	if params["CL2_NODES_MAX_COUNT"] != 1000 {
		t.Errorf("Expected CL2_NODES_MAX_COUNT to be 1000, got %v", params["CL2_NODES_MAX_COUNT"])
	}

	if params["CL2_SCHEDULER_THROUGHPUT_PODS"] != 5000 {
		t.Errorf("Expected CL2_SCHEDULER_THROUGHPUT_PODS to be 5000, got %v", params["CL2_SCHEDULER_THROUGHPUT_PODS"])
	}
}

func TestDefaultParamFunction(t *testing.T) {
	params := map[string]interface{}{
		"existingParam": 42,
		"CL2_TEST_PARAM": "test-value",
	}

	processor := NewTemplateProcessor(params)

	// Test existing parameter
	result := processor.defaultParam("existingParam", 10)
	if result != 42 {
		t.Errorf("Expected 42, got %v", result)
	}

	// Test non-existing parameter (should return default)
	result = processor.defaultParam("nonExistingParam", 99)
	if result != 99 {
		t.Errorf("Expected 99, got %v", result)
	}

	// Test CL2 style parameter
	result = processor.defaultParam("CL2_TEST_PARAM", "default")
	if result != "test-value" {
		t.Errorf("Expected 'test-value', got %v", result)
	}
}

func TestNumericParameterToStringConversion(t *testing.T) {
	// Numeric parameters should be converted to strings when template default is string
	templateContent := `max-count: {{DefaultParam "MAX_COUNT" "100"}}
min-count: {{DefaultParam "MIN_COUNT" "1"}}
desired-count: {{DefaultParam "DESIRED_COUNT" "10"}}`

	// Create temp directory and file
	tempDir, err := os.MkdirTemp("", "numeric_conversion_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	templatePath := filepath.Join(tempDir, "numeric_template.yaml")
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// Test with numeric parameters (as they would come from conf.yaml)
	numericParams := map[string]interface{}{
		"MAX_COUNT":     1000,  // numeric value
		"MIN_COUNT":     20,    // numeric value
		"DESIRED_COUNT": 50,    // numeric value
	}

	processor := NewTemplateProcessor(numericParams)
	result, err := processor.ProcessConfigFile(templatePath)
	if err != nil {
		t.Fatalf("Failed to process template: %v", err)
	}

	// Verify the numeric parameters were converted to strings
	// No quotes in YAML output
	expectedValues := []string{
		"max-count: 1000",
		"min-count: 20",
		"desired-count: 50",
	}

	for _, expected := range expectedValues {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected result to contain '%s', but got:\n%s", expected, result)
		}
	}

	// Verify no numeric formatting issues
	if strings.Contains(result, "e+") || strings.Contains(result, "E+") {
		t.Errorf("Result contains scientific notation, should be plain numbers: %s", result)
	}
}

func TestStringParameterToNumericConversion(t *testing.T) {
	// String parameters should be converted to numbers when template default is numeric
	templateContent := `replicas: {{DefaultParam "REPLICA_COUNT" 5}}
timeout: {{DefaultParam "TIMEOUT_SECONDS" 30}}
calculated: {{Add (DefaultParam "BASE_VALUE" 100) 50}}`

	// Create temp directory and file
	tempDir, err := os.MkdirTemp("", "string_to_numeric_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	templatePath := filepath.Join(tempDir, "numeric_default_template.yaml")
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// String parameters
	stringParams := map[string]interface{}{
		"REPLICA_COUNT":   "10",   // string value
		"TIMEOUT_SECONDS": "60",   // string value
		"BASE_VALUE":      "200",  // string value for math operation
	}

	processor := NewTemplateProcessor(stringParams)
	result, err := processor.ProcessConfigFile(templatePath)
	if err != nil {
		t.Fatalf("Failed to process template: %v", err)
	}

	// Verify the string parameters were converted to numbers (no quotes in YAML output)
	expectedValues := []string{
		"replicas: 10",
		"timeout: 60",
		"calculated: 250", // 200 + 50
	}

	for _, expected := range expectedValues {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected result to contain '%s', but got:\n%s", expected, result)
		}
	}

	// Verify no quotes around numeric values
	unexpectedValues := []string{
		`replicas: "10"`,
		`timeout: "60"`,
		`calculated: "250"`,
	}

	for _, unexpected := range unexpectedValues {
		if strings.Contains(result, unexpected) {
			t.Errorf("Result should not contain quoted numeric value '%s', but got:\n%s", unexpected, result)
		}
	}
}

func TestAutoscalerNodeTemplateProcessing(t *testing.T) {
	// Test the exact scenario used in autoscaler setup
	nodeTemplateContent := `apiVersion: v1
kind: Node
metadata:
	 annotations:
	   cluster-autoscaler.kwok.nodegroup/max-count: "{{DefaultParam "MAX_COUNT" "100"}}"
	   cluster-autoscaler.kwok.nodegroup/min-count: "{{DefaultParam "MIN_COUNT" "1"}}"
	   cluster-autoscaler.kwok.nodegroup/desired-count: "{{DefaultParam "DESIRED_COUNT" "10"}}"
	 name: kwok-node`

	// Create temp directory and file
	tempDir, err := os.MkdirTemp("", "autoscaler_node_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	templatePath := filepath.Join(tempDir, "node_template.yaml")
	err = os.WriteFile(templatePath, []byte(nodeTemplateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// Simulate the exact parameters passed from autoscaler setup
	// These are already converted to strings using fmt.Sprintf("%v", val)
	nodeParams := map[string]interface{}{
		"MAX_COUNT":     "1000", // string value (converted from numeric config)
		"MIN_COUNT":     "20",   // string value (converted from numeric config)
		"DESIRED_COUNT": "50",   // string value (converted from numeric config)
	}

	processor := NewTemplateProcessor(nodeParams)
	result, err := processor.ProcessConfigFile(templatePath)
	if err != nil {
		t.Fatalf("Failed to process node template: %v", err)
	}

	t.Logf("Processed node template result:\n%s", result)

	// Verify the parameters were substituted correctly as quoted strings
	expectedValues := []string{
		`max-count: "1000"`,
		`min-count: "20"`,
		`desired-count: "50"`,
	}

	for _, expected := range expectedValues {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected result to contain '%s', but got:\n%s", expected, result)
		}
	}
}

func TestNodeTemplateProcessing(t *testing.T) {
	// Test processing of node template similar to kwok-node-template.yaml
	nodeTemplateContent := `apiVersion: v1
kind: Node
metadata:
	 annotations:
	   cluster-autoscaler.kwok.nodegroup/max-count: {{DefaultParam "MAX_COUNT" "100"}}
	   cluster-autoscaler.kwok.nodegroup/min-count: {{DefaultParam "MIN_COUNT" "1"}}
	   cluster-autoscaler.kwok.nodegroup/desired-count: {{DefaultParam "DESIRED_COUNT" "10"}}
	 name: kwok-node
`

	// Create temp directory and file
	tempDir, err := os.MkdirTemp("", "node_template_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	templatePath := filepath.Join(tempDir, "node_template.yaml")
	err = os.WriteFile(templatePath, []byte(nodeTemplateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// Test with specific node parameters
	nodeParams := map[string]interface{}{
		"MAX_COUNT":     "1000",
		"MIN_COUNT":     "20",
		"DESIRED_COUNT": "50",
	}

	processor := NewTemplateProcessor(nodeParams)
	result, err := processor.ProcessConfigFile(templatePath)
	if err != nil {
		t.Fatalf("Failed to process node template: %v", err)
	}

	// Verify the parameters were substituted correctly
	expectedValues := []string{
		"max-count: 1000",
		"min-count: 20",
		"desired-count: 50",
	}

	for _, expected := range expectedValues {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected result to contain '%s', but got:\n%s", expected, result)
		}
	}
}

func TestErrorHandling(t *testing.T) {
	processor := NewTemplateProcessor(nil)

	t.Run("mathOp with invalid operation", func(t *testing.T) {
		result := processor.add(10, 5)
		if result != 15.0 {
			t.Errorf("Expected 15, got %v", result)
		}

		_, err := processor.mathOp(10, 5, "invalid")
		if err == nil {
			t.Error("Expected error for invalid operation, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "unknown operation") {
			t.Errorf("Expected 'unknown operation' error, got: %v", err)
		}

		_, err = processor.mathOp(10, 0, "/")
		if err == nil {
			t.Error("Expected error for division by zero, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "division by zero") {
			t.Errorf("Expected 'division by zero' error, got: %v", err)
		}
	})

	t.Run("toString with unsupported type", func(t *testing.T) {
		result, err := processor.toString(42)
		if err != nil {
			t.Errorf("Expected no error for int, got: %v", err)
		}
		if result != "42" {
			t.Errorf("Expected '42', got '%s'", result)
		}

		complex := complex(1, 2)
		_, err = processor.toString(complex)
		if err == nil {
			t.Error("Expected error for complex number, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "unsupported type") {
			t.Errorf("Expected 'unsupported type' error, got: %v", err)
		}

		safeResult := processor.safeToString(complex)
		if safeResult == "" {
			t.Error("Expected non-empty string from safeToString")
		}
	})
}