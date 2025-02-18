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
	"fmt"
	"github.com/apache/yunikorn-core/pkg/log"
	"github.com/apache/yunikorn-release/soak/framework"
	"github.com/apache/yunikorn-release/soak/pkg/constants"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var logger *zap.Logger = log.Log(log.Test)

func setK8sContext() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}
	kubeconfigPath := filepath.Join(homeDir, ".kube", "config")
	os.Setenv("KUBECONFIG", kubeconfigPath)
	logger.Info("Set KUBECONFIG", zap.String("path", kubeconfigPath))

	contextCmd := exec.Command("kubectl", "config", "use-context", constants.KindSoakTestCluster)
	contextOutput, err := contextCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to switch kubectl context: %v, output: %s", err, string(contextOutput))
	}
	logger.Info("Kubectl context switch output", zap.String("output", strings.TrimSpace(string(contextOutput))))

	currentContextCmd := exec.Command("kubectl", "config", "current-context")
	_, err = currentContextCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get current context: %v", err)
	}

	return nil
}

func upgradeSchedulerPerConfig(scheduler framework.SchedulerFields) error {
	if err := setK8sContext(); err != nil {
		logger.Fatal("failed to set kubernetes context", zap.Error(err))
		return err
	}

	logger.Info("Scheduler details",
		zap.String("VcoreRequests", scheduler.VcoreRequests),
		zap.String("MemoryRequests", scheduler.MemoryRequests),
		zap.String("VcoreLimits", scheduler.VcoreLimits),
		zap.String("MemoryLimits", scheduler.MemoryLimits),
		zap.String("path", scheduler.Path))

	args := []string{
		"upgrade",
		"yunikorn",
		"yunikorn/yunikorn",
		"-n", "yunikorn",
	}

	var moreArgs []string

	if scheduler.VcoreRequests != "" {
		moreArgs = append(moreArgs, "--set", fmt.Sprintf("resources.requests.cpu=%s", scheduler.VcoreRequests))
	}
	if scheduler.MemoryRequests != "" {
		moreArgs = append(moreArgs, "--set", fmt.Sprintf("resources.requests.memory=%s", scheduler.MemoryRequests))
	}
	if scheduler.VcoreLimits != "" {
		moreArgs = append(moreArgs, "--set", fmt.Sprintf("resources.limits.cpu=%s", scheduler.VcoreLimits))
	}
	if scheduler.MemoryLimits != "" {
		moreArgs = append(moreArgs, "--set", fmt.Sprintf("resources.limits.memory=%s", scheduler.MemoryLimits))
	}

	if len(moreArgs) > 0 {
		args = append(args, moreArgs...)

		cmd := exec.Command("helm", args...)

		logger.Info("Helm command to be executed",
			zap.String("command", fmt.Sprintf("helm %s", strings.Join(args, " "))))

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("helm upgrade failed: %v", err)
		}

		logger.Info("Helm upgrade successful",
			zap.String("command", fmt.Sprintf("helm %s", strings.Join(args, " "))),
			zap.String("output", string(output)))
	}

	if scheduler.Path != "" {
		kubectlArgs := []string{"apply"}
		kubectlArgs = append(kubectlArgs, "-f", scheduler.Path, "-n", "yunikorn")
		kubectlCmd := exec.Command("kubectl", kubectlArgs...)
		logger.Info("Kubectl command to be executed",
			zap.String("command", fmt.Sprintf("kubectl %s", strings.Join(kubectlArgs, " "))))

		kubectlOutput, err := kubectlCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("kubectl apply failed: %v", err)
		}
		logger.Info("Kubectl apply successful", zap.String("output", strings.TrimSpace(string(kubectlOutput))))
	}

	return nil
}

func setNodeScalePerConfig(node framework.NodeFields) error {
	if err := setK8sContext(); err != nil {
		logger.Fatal("failed to set kubernetes context", zap.Error(err))
		return err
	}

	logger.Info("Node details",
		zap.String("path", node.Path),
		zap.String("NodesDesiredCount", node.DesiredCount),
		zap.String("maxCount", node.MaxCount))

	templateContent, err := os.ReadFile(node.Path)
	if err != nil {
		logger.Error("failed to read template file", zap.String("path", node.Path))
		return err
	}
	desiredCount, err := strconv.Atoi(node.DesiredCount)
	if err != nil {
		logger.Error("failed to parse desired count",
			zap.String("desiredCount", node.DesiredCount))
	}

	for i := 0; i < desiredCount; i++ {
		currentNodeName := fmt.Sprintf("kwok-node-%d", i)
		nodeContent := strings.ReplaceAll(string(templateContent), "kwok-node-i", currentNodeName)

		tmpfile, err := os.CreateTemp("", "node-*.yaml")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %v", err)
		}
		defer os.Remove(tmpfile.Name()) // Clean up

		if _, err := tmpfile.WriteString(nodeContent); err != nil {
			return fmt.Errorf("failed to write to temp file: %v", err)
		}
		if err := tmpfile.Close(); err != nil {
			return fmt.Errorf("failed to close temp file: %v", err)
		}

		cmd := exec.Command("kubectl", "apply", "-f", tmpfile.Name())
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to apply node configuration: %v", err)
		}

		logger.Info("Applied node configuration",
			zap.String("nodeName", currentNodeName),
			zap.String("output", string(output)))
	}

	return nil
}
