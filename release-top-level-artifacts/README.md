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
----
Apache YuniKorn (incubating) is a light-weighted, universal resource scheduler for container orchestrator systems.
It was created to achieve fine-grained resource sharing for various workloads efficiently on a large scale, multi-tenant,
and cloud-native environment. YuniKorn brings a unified, cross-platform scheduling experience for mixed workloads consists
of stateless batch workloads and stateful services.

## Build

Run the script `build-docker-images.sh` to build docker images.

```shell script
# specify the docker repo
./build-docker-images.sh -r <REPO_NAME> -v <VERSION>

# for example, the following command will
# build 3 docker images like below:
#  foo/yunikorn-scheduler-k8s:0.8.0
#  foo/yunikorn-scheduler-admission-controller:0.8.0
#  foo/yunikorn-web:0.8.0
./build-docker-images.sh -r foo -v 0.8.0
```

## Run YuniKorn on an existing K8s cluster

The simplest way to run YuniKorn is to use our helm charts,
you can find the templates in the release package `helm-charts`.
There are a few prerequisites:
1. A existing K8s cluster is up and running.
2. Helm chart client is installed.

then simply run command:

```shell script
helm install ./yunikorn
```

For more instructions, please refer to [User Guide](https://github.com/apache/incubator-yunikorn-core/blob/master/docs/user-guide.md#quick-start).
