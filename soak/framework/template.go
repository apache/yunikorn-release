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

 Inspired from https://github.com/kubernetes/perf-tests/tree/master/clusterloader2/pkg/config
*/

package framework

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

// TemplateProcessor handles Go template processing for clusterloader and node template config files
type TemplateProcessor struct {
	params map[string]interface{}
}

func NewTemplateProcessor(params map[string]interface{}) *TemplateProcessor {
	return &TemplateProcessor{
		params: params,
	}
}

func (tp *TemplateProcessor) ProcessConfigFile(configPath string) (string, error) {
	templateContent, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config template file %s: %w", configPath, err)
	}

	// Create a new template with custom functions
	tmpl := template.New(filepath.Base(configPath)).Funcs(tp.getTemplateFunctions())

	newTmpl, err := tmpl.Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", configPath, err)
	}

	var buf bytes.Buffer
	err = newTmpl.Execute(&buf, tp.params)
	if err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", configPath, err)
	}

	return buf.String(), nil
}

// ProcessAndWriteConfigFile processes a template and writes the result to a new file
func (tp *TemplateProcessor) ProcessAndWriteConfigFile(templatePath, outputPath string) error {
	processedContent, err := tp.ProcessConfigFile(templatePath)
	if err != nil {
		return err
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	// Write processed content to output file
	err = os.WriteFile(outputPath, []byte(processedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write processed config to %s: %w", outputPath, err)
	}

	return nil
}

// getTemplateFunctions returns custom template functions used in clusterloader configs
func (tp *TemplateProcessor) getTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"DefaultParam": tp.defaultParam,
		"Param":        tp.param,
		"Add":          tp.add,
		"Sub":          tp.sub,
		"Mul":          tp.mul,
		"Div":          tp.div,
		"ToString":     tp.safeToString,
		"ToInt":        tp.toInt,
		"ToFloat":      tp.toFloat,
	}
}

// safeToString is a helper that handles errors from toString
func (tp *TemplateProcessor) safeToString(val interface{}) string {
	result, err := tp.toString(val)
	if err != nil {
		// For internal use, just return a simple string representation
		return fmt.Sprintf("%v", val)
	}
	return result
}

// defaultParam returns the parameter value if it exists, otherwise returns the default value
func (tp *TemplateProcessor) defaultParam(paramName interface{}, defaultValue interface{}) interface{} {
	paramStr := tp.safeToString(paramName)

	// Check if parameter exists in our params map
	if val, exists := tp.params[paramStr]; exists {
		// Convert the parameter value to match the type of the default value
		convertedVal := tp.convertToType(tp.safeToString(val), defaultValue)

		// If the default value is a string, ensure the output is also a string
		// This is important for YAML values that need to be quoted
		// Tested
		if _, isStringDefault := defaultValue.(string); isStringDefault {
			return tp.safeToString(convertedVal)
		}

		return convertedVal
	}

	// Check if parameter exists as environment variable (clusterloader2 style)
	if envVal := os.Getenv(paramStr); envVal != "" {
		convertedVal := tp.convertToType(envVal, defaultValue)

		// If the default value is a string, output should also be a string
		if _, isStringDefault := defaultValue.(string); isStringDefault {
			return tp.safeToString(convertedVal)
		}

		return convertedVal
	}

	return defaultValue
}

// param returns the parameter value, panics if not found
func (tp *TemplateProcessor) param(paramName interface{}) interface{} {
	paramStr := tp.safeToString(paramName)

	if val, exists := tp.params[paramStr]; exists {
		return val
	}

	if envVal := os.Getenv(paramStr); envVal != "" {
		return envVal
	}

	panic(fmt.Sprintf("required parameter %s not found", paramStr))
}

// convertToType attempts to convert a string value to the same type as the reference value
func (tp *TemplateProcessor) convertToType(value string, reference interface{}) interface{} {
	switch reference.(type) {
	case int, int32, int64:
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	case float32, float64:
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	case bool:
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return value // Default to string if conversion fails
}

// Math operations for templates
func (tp *TemplateProcessor) add(a, b interface{}) interface{} {
	result, err := tp.mathOp(a, b, "+")
	if err != nil {
		// In templates, we need to return a value even on error
		return fmt.Sprintf("ERROR: %v", err)
	}
	return result
}

func (tp *TemplateProcessor) sub(a, b interface{}) interface{} {
	result, err := tp.mathOp(a, b, "-")
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}
	return result
}

func (tp *TemplateProcessor) mul(a, b interface{}) interface{} {
	result, err := tp.mathOp(a, b, "*")
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}
	return result
}

func (tp *TemplateProcessor) div(a, b interface{}) interface{} {
	result, err := tp.mathOp(a, b, "/")
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}
	return result
}

// mathOp performs math operations on two values
func (tp *TemplateProcessor) mathOp(a, b interface{}, op string) (float64, error) {
	aVal := tp.toFloat(a)
	bVal := tp.toFloat(b)

	switch op {
	case "+":
		return aVal + bVal, nil
	case "-":
		return aVal - bVal, nil
	case "*":
		return aVal * bVal, nil
	case "/":
		if bVal == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return aVal / bVal, nil
	default:
		return 0, fmt.Errorf("unknown operation: %s", op)
	}
}

// Type conversion functions
func (tp *TemplateProcessor) toString(val interface{}) (string, error) {
	if val == nil {
		return "", nil
	}

	switch v := val.(type) {
	case string:
		return v, nil
	case int, int32, int64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%g", v), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		// This is safer as it makes it explicit when an unsupported type is used
		return "", fmt.Errorf("unsupported type for conversion to string: %T", val)
	}
}

func (tp *TemplateProcessor) toInt(val interface{}) int {
	switch v := val.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case string:
		if intVal, err := strconv.Atoi(v); err == nil {
			return intVal
		}
	}
	return 0
}

func (tp *TemplateProcessor) toFloat(val interface{}) float64 {
	switch v := val.(type) {
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	case string:
		if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
			return floatVal
		}
	}
	return 0.0
}

// BuildParameterMap builds a parameter map from test case configuration
func BuildParameterMap(testCase TestCase) map[string]interface{} {
	params := make(map[string]interface{})

	// Since TestCaseParams is now a map[string]interface{}, directly copy all parameters
	for key, value := range testCase.Params {
		params[key] = value

		// Also add with common clusterloader2 environment variable naming convention
		envName := "CL2_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		params[envName] = value
	}

	// Add some common parameter mappings
	if numPods, exists := params["numPods"]; exists {
		params["CL2_SCHEDULER_THROUGHPUT_PODS"] = numPods
	}
	if nodesMaxCount, exists := params["nodesMaxCount"]; exists {
		params["CL2_NODES_MAX_COUNT"] = nodesMaxCount
	}
	if nodesDesiredCount, exists := params["nodesDesiredCount"]; exists {
		params["CL2_NODES_DESIRED_COUNT"] = nodesDesiredCount
	}

	return params
}
