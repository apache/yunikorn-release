<!--
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 -->
# Apache YuniKorn - A Universal Scheduler

Apache YuniKorn is a light-weight, universal resource scheduler for container orchestrator systems.
It was created to achieve fine-grained resource sharing for various workloads efficiently on a large scale, multi-tenant, and cloud-native environment. YuniKorn brings a unified, cross-platform, scheduling experience for mixed workloads that consist of AI, Machine Learning, stateless batch workloads and stateful services. 

YuniKorn now supports K8s and can be deployed as a custom K8s scheduler. YuniKorn's architecture design also allows adding different shim layer and adopt to different ResourceManager implementation including Apache Hadoop YARN, or any other systems. 

## Feature highlights

- Features to support both AI, Machine Learning or batch jobs and long-running/stateful services.
- Hierarchical queues with guaranteed/maximum resource quotas and applications.
- User and group quotas configurable for each  queue with maximum applicatins and resources.
- Resource fairness between queues, users and apps.
- Scheduling policies configurable per queue: FIFO, priority and state based.
- Cross-queue preemption based on fairness.
- Automatically map incoming container requests to queues by policies. 
- Node partition: partition cluster to sub-clusters with dedicated quota/ACL management.
- Fully compatible with K8s predicates, events, PV/PVC and admin commands.
- Support for [Cluster AutoScaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler) and [Karpenter](https://karpenter.sh/) to drive cluster scale up and down. 

## Deployment model
YuniKorn can be deployed with [helm-charts](https://artifacthub.io/packages/helm/yunikorn/yunikorn) on an existing K8s cluster. It can be deployed with or without the admission controller. When the admission controller is enabled, YuniKorn will be the primary scheduler that takes over the resource scheduling (the admission controller runs a mutation webhook that automatically mutates pod spec's schedulerName to yunikorn); when it is disabled, user needs to manually change the schedulerName to `yunikorn` in order to get apps scheduled by YuniKorn.

## Supported K8s versions 

| K8s Version         | Support? |
|---------------------|:--------:|
| 1.23.x (or earlier) |    X     |
| 1.24.x              |    √     |
| 1.25.x              |    √     |
| 1.26.x              |    √     |
| 1.27.x              |    √     |
| 1.28.x              |    √     |
| 1.29.x              |    √     |

## Installing the chart
```
helm repo add yunikorn  https://apache.github.io/yunikorn-release
helm repo update 
helm install yunikorn yunikorn/yunikorn
```
## Configuration
The following table lists the configurable parameters of the YuniKorn chart and their default values:

| Parameter                                       | Description                                                 | Default                         |
| ----------------------------------------------- | ----------------------------------------------------------- | ------------------------------- |
| `imagePullSecrets`                              | Docker repository secrets                                   | `[]`                            |
| `serviceAccount`                                | Service account name                                        | `yunikorn-admin`                |
| `hostNetwork`                                   | Whether scheduler should run in the host network            | `false`                         |
| `image.repository`                              | Scheduler image repository                                  | `apache/yunikorn`               |
| `image.tag`                                     | Scheduler image tag                                         | `scheduler-latest`              |
| `image.pullPolicy`                              | Scheduler image pull policy                                 | `Always`                        |
| `pluginImage.repository`                        | Scheduler plugin image repository                           | `apache/yunikorn`               |
| `pluginImage.tag`                               | Scheduler plugin image tag                                  | `scheduler-plugin-latest`       |
| `pluginImage.pullPolicy`                        | Scheduler plugin image pull policy                          | `Always`                        |
| `podLabels`                                     | Scheduler pod labels                                        | `{}`                            |
| `podAnnotations`                                | Scheduler pod annotations                                   | `{}`                            |
| `admissionController.podLabels`                 | Admission controller pod labels                             | `{}`                            |
| `admissionController.podAnnotations`            | Admission controller pod annotations                        | `{}`                            |
| `admissionController.replicaCount`              | Admission controller replicas to be deployed                | `1`                             |
| `admissionController.serviceAccount`            | Admission controller service account name                   | `yunikorn-admission-controller` |
| `admissionController.image.repository`          | Admission controller image repository                       | `apache/yunikorn`               |
| `admissionController.image.tag`                 | Admission controller image tag                              | `admission-latest`              |
| `admissionController.image.pullPolicy`          | Admission controller image pull policy                      | `Always`                        |
| `admissionController.hostNetwork`               | Whether admission controller should run in the host network | `true`                          |
| `admissionController.resources.requests.cpu`    | Admission controller CPU resource requests                  | `100m`                          |
| `admissionController.resources.requests.memory` | Admission controller memory resource requests               | `500Mi`                         |
| `admissionController.resources.limits.cpu`      | Admission controller CPU resource limit                     | `500m`                          |
| `admissionController.resources.limits.memory`   | Admission controller memory resource limit                  | `500Mi`                         |
| `admissionController.nodeSelector`              | Admission controller deployment nodeSelector(s)             | `{}`                            |
| `admissionController.tolerations`               | Admission controller deployment tolerations                 | `[]`                            |
| `admissionController.affinity`                  | Admission controller deployment affinity                    | `{}`                            |
| `admissionController.topologySpreadConstraints` | Admission controller deployment topology spread constraints | `[]`                            |
| `admissionController.service.type`              | Admission controller service type                           | `ClusterIP`                     |
| `admissionController.priorityClassName`         | Admission controller pod priority                           | `""`                            |
| `priorityClassName`                             | Scheduler pod priority                                      | `""`                            |
| `service.type`                                  | Scheduler service type                                      | `ClusterIP`                     |
| `service.port`                                  | Port of the scheduler service                               | `9080`                          |
| `service.portWeb`                               | Port of the web application service                         | `9889`                          |
| `resources.requests.cpu`                        | CPU resource requests                                       | `200m`                          |
| `resources.requests.memory`                     | Memory resource requests                                    | `1Gi`                           |
| `resources.limits.cpu`                          | CPU resource limit                                          | `4`                             |
| `resources.limits.memory`                       | Memory resource limit                                       | `2Gi`                           |
| `web.image.repository`                          | Web app image repository                                    | `apache/yunikorn`               |
| `web.image.tag`                                 | Web app image tag                                           | `web-latest`                    |
| `web.image.pullPolicy`                          | Web app image pull policy                                   | `Always`                        |
| `web.resources.requests.cpu`                    | Web app CPU resource requests                               | `100m`                          |
| `web.resources.requests.memory`                 | Web app memory resource requests                            | `100Mi`                         |
| `web.resources.limits.cpu`                      | Web app CPU resource limit                                  | `200m`                          |
| `web.resources.limits.memory`                   | Web app memory resource limit                               | `500Mi`                         |
| `embedAdmissionController`                      | Flag for enabling/disabling the admission controller        | `true`                          |
| `enableSchedulerPlugin`                         | Flag for enabling/disabling scheduler plugin mode           | `false`                         |
| `enableWebService`                              | Flag for enabling/disabling web service                     | `true`                          |
| `nodeSelector`                                  | Scheduler deployment nodeSelector(s)                        | `{}`                            |
| `tolerations`                                   | Scheduler deployment tolerations                            | `[]`                            |
| `affinity`                                      | Scheduler deployment affinity                               | `{}`                            |
| `yunikornDefaults`                              | Data for the `yunikorn-defaults` ConfigMap                  | `{}`                            |


These parameters can be passed in via helm's `--set` and `--values (-f)` options, such as `--set resources.requests.cpu=2000m` or `-f myvalues.yaml`.

In addition to these parameters, YuniKorn supports reading most of its runtime configuration directly from two ConfigMaps in the installed namespace: `yunikorn-defaults` and `yunikorn-configs`. The `yunikorn-defaults` ConfigMap is rendered by Helm based on the contents of the `yunikornDefaults` parameter. The `yunikorn-configs` ConfigMap is entirely ignored by Helm, and can be populated (or not) using external mechanisms. At runtime, YuniKorn will evaluate both ConfigMaps, using the contents of `yunikorn-configs` as an override to `yunikorn-defaults`.

## Deprecated Configuration
The following configuration parameters are deprecated and will be ignored in a future release of YuniKorn:

| Parameter                               | Description                                         | Replacement                                                       |
|-----------------------------------------|-----------------------------------------------------|-------------------------------------------------------------------|
| `operatorPlugins`                       | Scheduler operator plugins                          | `service.operatorPlugins` ConfigMap entry                         |
| `placeholderImage`                      | Docker image of the placeholder pods                | `service.placeholderImage` ConfigMap entry                        |
| `admissionController.processNamespaces` | List of namespace regexes to process (empty=ALL)    | `admissionController.filtering.processNamespaces` ConfigMap entry |
| `admissionController.bypassNamespaces`  | List of namespace regexes to bypass (empty=NONE)    | `admissionController.filtering.bypassNamespaces` ConfigMap entry  |
| `admissionController.labelNamespaces`   | List of namespace regexes to label (empty=ALL)      | `admissionController.filtering.labelNamespaces` ConfigMap entry   |
| `admissionController.noLabelNamespaces` | List of namespace regexes not to label (empty=NONE) | `admissionController.filtering.noLabelNamespaces` ConfigMap entry |
| `configuration`                         | YAML-formatted queue configuration                  | `queues.yaml` ConfigMap entry                                     |

Currently, if both the deprecated parameter and the replacement ConfigMap entry are specified, the ConfigMap entry will take precedence.
