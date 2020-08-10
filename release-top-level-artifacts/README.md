<!--
#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
-->

# Apache YuniKorn (Incubating)
Apache YuniKorn (Incubating) is a light-weight, universal resource scheduler for container orchestrator systems.
It was created to achieve fine-grained resource sharing for various workloads efficiently on a large scale, multi-tenant,
and cloud-native environment. YuniKorn brings a unified, cross-platform scheduling experience for mixed workloads consists
of stateless batch workloads and stateful services.

## Build pre-requisites
These instructions provided are tailored to the source release.
Details on how to set up a full development environment can be found in [Building YuniKorn](https://yunikorn.apache.org/docs/next/developer_guide/build).

General requirement for building YuniKorn images from this release:
* Make
* Docker 

### Yunikorn Scheduler
The scheduler and shim are build as one set of artifacts and have one requirement:
* Go 1.11 or later

### Yunikorn web UI
The project requires a number of external tools to be installed before the build and development.
A build requires the following tools to be installed:
* Node.js 10.16.2
* Angular CLI 8.3.19
* yarn 1.21

NOTE: the scheduler can be used without a web UI build or deployed.

## Building
Run the `make` command to build docker images.

```shell script
make
```
The command will generate the following three docker images in the local docker repository:
* apache/yunikorn:scheduler-0.9.0
* apache/yunikorn:admission-0.9.0
* apache/yunikorn:web-0.9.0 

## Deploying YuniKorn 
The simplest way to run YuniKorn is to use the provided helm charts, you can find the templates in the release 
package `helm-charts`.
There are a few prerequisites:
1. An existing K8s cluster is up and running.
2. Helm chart client is installed.

If you have a cluster, and the helm client you can simply run:
```shell script
helm install ./yunikorn
```

## Customising the build
The `make` command will pass on the following two variables:
* VERSION
* REGISTRY
These variables can be used to generate customised build: 
```shell script
VERSION="0.9.1" REGISTRY="internal" make
```

The values defined in the helm charts assume a default build without changes to the `VERSION` or `REGISTRY`. 
Once you have built your own docker images, you will need to replace the docker image name in the helm chart templates.
Open `helm-charts/yunikorn/values.yaml` and replace the docker image information with ones you built.

For more instructions, please refer to [User Guide](https://yunikorn.apache.org/docs/next/).

## Deploying a convenience build
Apache YuniKorn (incubating) provides a convenience release with pre-build docker images and helm charts.
These can be accessed via the [downloads page](https://yunikorn.apache.org/community/download) and instructions are 
located in the [User Guide](https://yunikorn.apache.org/docs/next/).