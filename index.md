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

[![codecov](https://codecov.io/gh/apache/incubator-yunikorn-core/branch/master/graph/badge.svg)](https://codecov.io/gh/apache/incubator-yunikorn-core)
[![Go Report Card](https://goreportcard.com/badge/github.com/apache/incubator-yunikorn-core)](https://goreportcard.com/report/github.com/apache/incubator-yunikorn-core)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Repo Size](https://img.shields.io/github/repo-size/apache/incubator-yunikorn-core)](https://img.shields.io/github/repo-size/apache/incubator-yunikorn-core)

<img src="https://raw.githubusercontent.com/apache/incubator-yunikorn-core/master/images/logo/yunikorn-logo-blue.png" width="200">

----

Apache YuniKorn (Incubating) is a light-weight, universal resource scheduler for container orchestrator systems.
It was created to achieve fine-grained resource sharing for various workloads efficiently on a large scale, multi-tenant,
and cloud-native environment. YuniKorn brings a unified, cross-platform, scheduling experience for mixed workloads that consist
of stateless batch workloads and stateful services. 

YuniKorn now supports K8s and can be deployed as a custom K8s scheduler. YuniKorn's architecture design also allows adding different
shim layer and adopt to different ResourceManager implementation including Apache Hadoop YARN, or any other systems. 

## Architecture

Following chart illustrates the high-level architecture of YuniKorn.

![Architecture](https://raw.githubusercontent.com/apache/incubator-yunikorn-site/master/docs/assets/architecture.png)

YuniKorn consists of the following components spread over multiple code repositories, most of the following projects are written in GoLang.

- _Scheduler core_: the brain of the scheduler, which makes placement decisions (Allocate container X on node Y)
  according to pre configured policies. See more in current repo [yunikorn-core](https://github.com/apache/incubator-yunikorn-core).
  _Scheduler core_ is implemented in a way to be agnostic to scheduler implementation.
- _Scheduler interface_: the common scheduler interface used by shims and the core scheduler.
  Contains the API layer (with GRPC/programming language bindings) which is agnostic to container orchestrator systems like YARN/K8s.
  See more in [yunikorn-scheduler-interface](https://github.com/apache/incubator-yunikorn-scheduler-interface).
- _Resource Manager shims_: allow container orchestrator systems talks to yunikorn-core through scheduler-interface.
   Which can be configured on existing clusters without code change.
   
   Currently, [yunikorn-k8shim](https://github.com/apache/incubator-yunikorn-k8shim) is available for Kubernetes integration. 
   Supporting other Resource Manager is our next priority.
- _Scheduler User Interface_: the YuniKorn web interface for app/queue management.
   See more in [yunikorn-web](https://github.com/apache/incubator-yunikorn-web).

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

#### Performance testing
We love high-performance software, and we made tremendous efforts to make it to support large scale cluster and high-churning tasks. 
Here's the latest [performance test result](https://yunikorn.apache.org/docs/next/performance/evaluate_perf_function_with_kubemark).

#### Deployment model
Yunikorn can be deployed as a K8s custom scheduler, and take over all POD scheduling.
 
#### Verified K8s versions 

| K8s Version         | Supported from version | Support ended |
|---------------------|:----------------------:|:-------------:|
| 1.12.x (or earlier) |           -            |       -       |
| 1.13.x              |         0.8.0          |    0.10.0     |
| 1.14.x              |         0.8.0          |    0.10.0     |
| 1.15.x              |         0.8.0          |    0.10.0     |
| 1.16.x              |         0.10.0         |    0.11.0     |
| 1.17.x              |         0.10.0         |    0.11.0     |
| 1.18.x              |         0.10.0         |    0.11.0     |
| 1.19.x              |         0.11.0         |     1.0.0     |
| 1.20.x              |         0.12.1         |       -       |
| 1.21.x              |         0.12.1         |       -       |
| 1.22.x              |         0.12.2         |       -       |
| 1.23.x              |         0.12.2         |       -       |

### Web UI

YuniKorn has builtin web UIs for queue hierarchies and apps. See below: 

![Web-UI](https://raw.githubusercontent.com/apache/incubator-yunikorn-site/master/docs/assets/yk-ui-screenshots.gif)


## Roadmap

Want to learn more about future of YuniKorn? You can find more information about what are already supported and future plans in the [Road Map](https://yunikorn.apache.org/community/roadmap).

## How to use

The simplest way to run YuniKorn is to build a docker image and then deployed to Kubernetes with a yaml file,
running as a customized scheduler. Then you can run workloads with this scheduler.
See more instructions from the [User Guide](https://yunikorn.apache.org/docs/next/).

## How can I get involved?

Apache YuniKorn (Incubating) community includes engineers from Alibaba, Apple, 
Cloudera, Linkedin, Microsoft, Nvidia, Tencent, Uber, etc. (sorted by alphabet). Want to join the community? 
We welcome any form of contributions, code, documentation or suggestions! 

To get involved, please read following resources.
- Logging an issue or improvement use the [Reporting an issue Guide](https://yunikorn.apache.org/community/reporting_issues).
- Before contributing code or documentation to YuniKorn, please read our [Get Involved](https://yunikorn.apache.org/community/get_involved) guide.
- When you are coding use the [Coding Guidelines](https://yunikorn.apache.org/community/coding_guidelines).
- Please read [How to Contribute](https://yunikorn.apache.org/community/how_to_contribute) to understand the procedure and guidelines of making contributions.
- We have periodic community sync ups in multiple timezones and languages, please find [Events](https://yunikorn.apache.org/community/events) to attend online sync ups.

## Other Resources

**Demo videos**

- Subscribe to [YuniKorn Youtube Channel](https://www.youtube.com/channel/UCDSJ2z-lEZcjdK27tTj_hGw) to get notification about new demos!
- [Running YuniKorn on Kubernetes - a 12 minutes Hello-world demo](https://www.youtube.com/watch?v=cCHVFkbHIzo)
- [YuniKorn configuration hot-refresh introduction](https://www.youtube.com/watch?v=3WOaxoPogDY)
- [Yunikorn scheduling and volumes on Kubernetes](https://www.youtube.com/watch?v=XDrjOkMp3k4)
- [Yunikorn placement rules for applications](https://www.youtube.com/watch?v=DfhJLMjaFH0)

**Communication channels**

- Mailing lists are:
  - for people wanting to contribute to the project: [dev@yunikorn.apache.org](mailto:dev@yunikorn.apache.org) [subscribe](mailto:dev-subscribe@yunikorn.apache.org?subject="send this email to subscribe") [unsubscribe](mailto:dev-unsubscribe@yunikorn.apache.org?subject="send this email to unsubscribe") [archives](https://lists.apache.org/list.html?dev@yunikorn.apache.org)
  - JIRA issue updates: issues@yunikorn.apache.org [subscribe](mailto:issues-subscribe@yunikorn.apache.org?subject="send this email to subscribe") [unsubscribe](mailto:issues-unsubscribe@yunikorn.apache.org?subject="send this email to unsubscribe") [archives](https://lists.apache.org/list.html?issues@yunikorn.apache.org)
  - for review messages and patches in GitHub pull requests reviews@yunikorn.apache.org [subscribe](mailto:reviews-subscribe@yunikorn.apache.org?subject="send this email to subscribe") [unsubscribe](mailto:reviews-unsubscribe@yunikorn.apache.org?subject="send this email to unsubscribe") [archives](https://lists.apache.org/list.html?reviews@yunikorn.apache.org)

- We use [Slack](https://slack.com/) as our collaboration system, you can join us by accessing [this link](https://join.slack.com/t/yunikornworkspace/shared_invite/enQtNzAzMjY0OTI4MjYzLTBmMDdkYTAwNDMwNTE3NWVjZWE1OTczMWE4NDI2Yzg3MmEyZjUyYTZlMDE5M2U4ZjZhNmYyNGFmYjY4ZGYyMGE).
Currently, we have following channels in the workspace: `#yunikorn-dev` and `#yunikorn-user`.

**Blog posts**

Apache based blogs:
- [Apache YuniKorn (Incubating) release announcements](https://blogs.apache.org/yunikorn/)

3rd party blog posts:
- [YuniKorn: a universal resource scheduler](https://blog.cloudera.com/blog/2019/07/yunikorn-a-universal-resource-scheduler/)
