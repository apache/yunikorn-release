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
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"path/filepath"
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

func setAutoscalerPerConfig(node framework.NodeFields) error {
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
		logger.Error("failed to read template file", zap.Error(err))
		return err
	}

	var nodeTemplate map[string]interface{}
	err = yaml.Unmarshal(templateContent, &nodeTemplate)
	if err != nil {
		logger.Error("failed to parse template YAML", zap.Error(err))
		return err
	}

	metadata, ok := nodeTemplate["metadata"].(map[string]interface{})
	if !ok {
		logger.Error("invalid metadata format in node template")
		return fmt.Errorf("invalid metadata format in node template")
	}

	annotations, ok := metadata["annotations"].(map[string]interface{})
	if !ok {
		logger.Error("invalid annotations format in node template")
		return fmt.Errorf("invalid annotations format in node template")
	}

	annotations["cluster-autoscaler.kwok.nodegroup/max-count"] = node.MaxCount
	annotations["cluster-autoscaler.kwok.nodegroup/min-count"] = node.DesiredCount
	annotations["cluster-autoscaler.kwok.nodegroup/desired-count"] = node.DesiredCount

	autoscalerConfigmapPath := "../../templates/autoscaler-configmap.yaml"

	autoscalerConfigmap, err := os.ReadFile(autoscalerConfigmapPath)
	if err != nil {
		logger.Error("failed to read autoscaler configmap template", zap.Error(err))
		return err
	}

	var autoscalerNodeList map[string]interface{}
	err = yaml.Unmarshal(autoscalerConfigmap, &autoscalerNodeList)
	if err != nil {
		logger.Error("failed to parse autoscalerConfigmap YAML", zap.Error(err))
		return err
	}
	logger.Info("Autoscaler Node List", zap.Any("autoscalerNodeList", autoscalerNodeList))

	var itemsSlice []interface{}
	itemsSlice = append(itemsSlice, nodeTemplate)
	autoscalerNodeList["items"] = itemsSlice

	autoscalerNodeListYaml, err := yaml.Marshal(autoscalerNodeList)
	if err != nil {
		logger.Error("failed to convert updated autoscalerNodeList to YAML", zap.Error(err))
		return err
	}
	logger.Info("Encoded autoscalerNodeListYaml", zap.Any("autoscalerNodeListYaml", autoscalerNodeListYaml))

	updatedAcCmTempFile, err := os.CreateTemp("", "updated-autoscaler-configmap-temp.yaml")
	if err != nil {
		logger.Error("failed to create updated-autoscaler-configmap-temp file", zap.Error(err))
		return err
	}

	updatedAcCmTempFilePath := updatedAcCmTempFile.Name()
	defer os.Remove(updatedAcCmTempFilePath)

	if _, err = updatedAcCmTempFile.Write(autoscalerNodeListYaml); err != nil {
		updatedAcCmTempFile.Close()
		logger.Error("failed to write to updated-autoscaler-configmap-temp file", zap.Error(err))
		return err
	}
	if err = updatedAcCmTempFile.Close(); err != nil {
		logger.Error("failed to close updated-autoscaler-configmap-temp file", zap.Error(err))
		return err
	}

	// Delete the default autoscaler configMap
	deleteConfigMapCmd := exec.Command("kubectl", "delete", "cm", "kwok-provider-templates")
	deleteConfigMapCmdOutput, err := deleteConfigMapCmd.CombinedOutput()
	if err != nil {
		logger.Error("fail to delete configmap", zap.Error(err))
		return err
	}
	logger.Info(string(deleteConfigMapCmdOutput))

	// Create a new autoscaler configMap
	createConfigMapCmd := exec.Command("kubectl", "create", "cm", "kwok-provider-templates",
		"--from-file=templates="+updatedAcCmTempFilePath)
	createConfigMapCmdOutput, err := createConfigMapCmd.CombinedOutput()
	if err != nil {
		logger.Error("fail to create new configmap", zap.Error(err))
		return err
	}
	logger.Info(string(createConfigMapCmdOutput))

	// Restart the autoscaler pod after updating the configmap
	restartAutoscalerPodCmd := exec.Command("kubectl", "rollout", "restart", "deployment", "autoscaler-kwok-cluster-autoscaler")
	restartAutoscalerPodCmdOutput, err := restartAutoscalerPodCmd.CombinedOutput()
	if err != nil {
		logger.Error("failed to restart autoscaler deployment", zap.Error(err))
		return err
	}
	logger.Info("Restarted autoscaler deployment", zap.String("output", string(restartAutoscalerPodCmdOutput)))

	logger.Info("Successfully set up kwok provider cluster autoscaler for desiredNodeCount and MaxNodeCount")

	return nil
}
