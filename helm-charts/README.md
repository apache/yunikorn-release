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
# Apache YuniKorn (Incubating) - A Universal Scheduler

Apache YuniKorn (Incubating) is a light-weight, universal resource scheduler for container orchestrator systems.
It was created to achieve fine-grained resource sharing for various workloads efficiently on a large scale, multi-tenant,
and cloud-native environment. YuniKorn brings a unified, cross-platform, scheduling experience for mixed workloads that consist
of stateless batch workloads and stateful services. 

YuniKorn now supports K8s and can be deployed as a custom K8s scheduler. YuniKorn's architecture design also allows adding different
shim layer and adopt to different ResourceManager implementation including Apache Hadoop YARN, or any other systems. 

## Feature highlights

### Main scheduling features:

- Features to support both batch jobs and long-running/stateful services
- Hierarchy queues with min/max resource quotas.
- Resource fairness between queues, users and apps.
- Cross-queue preemption based on fairness.
- Customized resource types (like GPU) scheduling support.
- Rich placement constraints support.
- Automatically map incoming container requests to queues by policies. 
- Node partition: partition cluster to sub-clusters with dedicated quota/ACL management.

### Integration with K8s:

The `k8shim` provides the integration for K8s scheduling and supported features include: 

- _Predicates:_ All kinds of predicates such as node-selector, pod affinity/anti-affinity, taint/tolerant, etc.
- _Persistent volumes:_ We have verified hostpath, EBS, NFS, etc. 
- _K8s namespace awareness:_ YuniKorn support hierarchical of queues, does it mean you need to give up K8s namespace? Answer is no, with simple config, YuniKorn is able to 
 support automatically map K8s namespaces to YuniKorn queues. All K8s-namespace-related ResourceQuota, permissions will be still valid.
- _Metrics:_ Prometheus, Grafana integration.
- _Cluster AutoScaler_: YuniKorn can nicely work with Cluster AutoScaler (https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler) to drive cluster scales up and down.
- _K8s Events_: YuniKorn also integrated with K8s events, so lots of information can be retrieved by using `kubectl describe pod`.

#### Deployment model
Yunikorn can be deployed as a K8s custom scheduler, and take over all POD scheduling. 
An enhancement is open to improve coexistence behaviour of the YuniKorn scheduler with other Kubernetes schedulers,
like the default scheduler, in the cluster: [Co-existing with other K8s schedulers](https://issues.apache.org/jira/browse/YUNIKORN-16). 
 
#### Verified K8s versions 

| K8s Version   | Support?  |
| ------------- |:-------------:|
| 1.12.x (or earlier) | X |
| 1.13.x | √ |
| 1.14.x | √ |
| 1.15.x | √ |
| 1.16.x | To be verified |
| 1.17.x | To be verified |

## Istalling the chart
```
helm repo add yunikorn  https://apache.github.io/incubator-yunikorn-release
helm repo update 
helm install yunikorn/yunikorn
```
## Configuration
The following table lists the configurable parameters of the YuniKorn chart and their default values.

| Parameter                         | Description                                                    | Default                                     |
| ---                               | ---                                                            | ---                                         |
| `imagePullSecrets`                | Docker repository secrets                                      | ` `  
| `serviceAccount`                  | Service account name                                           | `yunikorn-admin`  
| `image.repository`                | Scheduler image repository                                     | `apache/yunikorn` 
| `image.tag`                       | Scheduler image tag                                            | `scheduler-latest` 
| `image.pullPolicy`                | Scheduler image pull policy                                    | `Always`  
| `web_image.repository`            | web app image repository                                       | `apache/yunikorn` 
| `web_image.tag`                   | web app image tag                                              | `web-latest` 
| `web_image.pullPolicy`            | web app image pull policy                                      | `Always`  
| `service.port`                    | Port of the scheduler service                                  | `9080` 
| `service.port_web`                | Port of the web application service                            | `9889`  
| `resources.requests.cpu`          | CPU resource requests                                          | `200m`  
| `resources.requests.memory`       | Memory resource requests                                       | `1Gi`  
| `resources.limits.cpu`            | CPU resource limit                                             | `4`  
| `resources.limits.memory`         | Memory resource limit                                          | `2Gi` 
| `embedAdmissionController`        | Flag for enabling/disabling admission controller               | `true` 


 

 



