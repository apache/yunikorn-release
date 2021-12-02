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
	"context"
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/tools/clientcmd"

	"go.uber.org/zap"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubeClient struct {
	clientSet *kubernetes.Clientset
	configs   *rest.Config
}

func NewKubeClient(KubeConfigFile string) (*KubeClient, error) {
	kubeConfigFile := os.ExpandEnv(KubeConfigFile)
	fmt.Println(KubeConfigFile)
	if kubeConfigFile == "" {
		return nil, fmt.Errorf("specified kubeconfig file %s not found", kubeConfigFile)
	}
	restClientConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFile)
	if err != nil {
		return nil, err
	}
	configuredClient := kubernetes.NewForConfigOrDie(restClientConfig)
	return &KubeClient{
		clientSet: configuredClient,
		configs:   restClientConfig,
	}, nil
}

func GetListOptions(selectLabels map[string]string) *metav1.ListOptions {
	labelSelector := metav1.LabelSelector{MatchLabels: selectLabels}
	return &metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}
}

func GetEverythingListOptions() *metav1.ListOptions {
	return &metav1.ListOptions{LabelSelector: labels.Everything().String()}
}

func (kc *KubeClient) GetPods(namespace string, listOptions *metav1.ListOptions) (*apiv1.PodList, error) {
	return kc.clientSet.CoreV1().Pods(namespace).List(context.TODO(), *listOptions)
}

func (kc *KubeClient) GetNodes(listOptions *metav1.ListOptions) (*apiv1.NodeList, error) {
	return kc.clientSet.CoreV1().Nodes().List(context.TODO(), *listOptions)
}

func (kc *KubeClient) GetConfigMap(namespace string, name string, getOptions *metav1.GetOptions) (*apiv1.ConfigMap, error) {
	return kc.clientSet.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, *getOptions)
}

func (kc *KubeClient) CreateDeployment(namespace string, deployment *appsv1.Deployment) error {
	deploymentsClient := kc.clientSet.AppsV1().Deployments(namespace)
	Logger.Debug("creating deployment...")
	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	Logger.Debug("created deployment", zap.String("deploymentName", result.GetObjectMeta().GetName()))
	return err
}

func (kc *KubeClient) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	deploymentsClient := kc.clientSet.AppsV1().Deployments(namespace)
	return deploymentsClient.Get(context.TODO(), name, metav1.GetOptions{})
}

func (kc *KubeClient) DeleteDeployment(namespace, name string) error {
	deploymentsClient := kc.clientSet.AppsV1().Deployments(namespace)
	deletePolicy := metav1.DeletePropagationForeground
	return deploymentsClient.Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

// GetDeploymentInfo return basic information of deployment: (createTime, [desired, created, ready] replicas, error)
func (kc *KubeClient) GetDeploymentInfo(namespace, appID string) (time.Time, []int, error) {
	deployment, err := kc.GetDeployment(namespace, appID)
	if err != nil || deployment == nil {
		return time.Time{}, nil, err
	}
	return deployment.CreationTimestamp.Time, []int{int(*deployment.Spec.Replicas), int(deployment.Status.Replicas),
		int(deployment.Status.ReadyReplicas)}, nil
}

func (kc *KubeClient) GetConfigs() *rest.Config {
	return kc.configs
}

func (kc *KubeClient) GetClientSet() *kubernetes.Clientset {
	return kc.clientSet
}
