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

[![codecov](https://codecov.io/gh/apache/yunikorn-core/branch/master/graph/badge.svg)](https://codecov.io/gh/apache/yunikorn-core)
[![Go Report Card](https://goreportcard.com/badge/github.com/apache/yunikorn-core)](https://goreportcard.com/report/github.com/apache/yunikorn-core)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Repo Size](https://img.shields.io/github/repo-size/apache/yunikorn-core)](https://img.shields.io/github/repo-size/apache/yunikorn-core)

<img src="https://raw.githubusercontent.com/apache/yunikorn-site/master/static/img/logo/yunikorn-logo-blue.png" width="200">

----

[Apache YuniKorn](https://yunikorn.apache.org/) is a light-weight, universal resource scheduler for container orchestrator systems.
It was created to achieve fine-grained resource sharing for various workloads efficiently on a large scale, multi-tenant,
and cloud-native environment. YuniKorn brings a unified, cross-platform, scheduling experience for mixed workloads that consist
of stateless batch workloads and stateful services. 

YuniKorn now supports K8s and can be deployed as a custom K8s scheduler. YuniKorn's architecture design also allows adding different
shim layer and adopt to different ResourceManager implementation including Apache Hadoop YARN, or any other systems. 

## Feature highlights

- App-aware scheduling
- Hierarchy Resource Queues
- Job Ordering and Queuing
- Resource fairness
- Resource Reservation
- Throughput

Read the complete list of features from [here](https://yunikorn.apache.org/docs/get_started/core_features).

## Web UI

YuniKorn has builtin web UIs for queue hierarchies and apps. See below:

![Web-UI](https://raw.githubusercontent.com/apache/yunikorn-site/master/docs/assets/yk-ui-screenshots.gif)

## Supported K8s versions

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
| 1.20.x              |         0.12.1         |     1.2.0     |
| 1.21.x              |         0.12.1         |       -       |
| 1.22.x              |         0.12.2         |       -       |
| 1.23.x              |         0.12.2         |       -       |
| 1.24.x              |         1.0.0          |       -       |
| 1.25.x              |         1.2.0          |       -       |
| 1.26.x              |         1.2.0          |       -       |

## Useful links

- [Get started](https://yunikorn.apache.org/docs/)
- [Project roadmap](https://yunikorn.apache.org/community/roadmap)
- [Performance](https://yunikorn.apache.org/docs/performance/evaluate_perf_function_with_kubemark)
- [Get involved](https://yunikorn.apache.org/community/get_involved)
- [Other resources](https://yunikorn.apache.org/community/events)

