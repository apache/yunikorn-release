<!--
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to you under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

### Prerequisites
- [Helm](https://github.com/helm/helm#install)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Docker](https://docs.docker.com/get-docker/)
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [Kwok](https://kwok.sigs.k8s.io/docs/user/installation/)
- [autoscaler](https://kubernetes.github.io/autoscaler)
- [Go](https://golang.org/doc/install) (required for building clusterloader2)
- [Git](https://git-scm.com/downloads) (required for cloning repositories)

## Setup Scripts

### Complete Initial Setup
Sets up the entire soak test environment including YuniKorn, Kwok, autoscaler, and clusterloader2:
```bash
./soak/scripts/initial_setup.sh
```

### Install clusterloader2 Only
If you only need to install the clusterloader2 binary:
```bash
./soak/scripts/install_clusterloader2.sh
```
### Manual clusterloader2 Installation
If you prefer to install clusterloader2 manually, follow the official guide:
https://github.com/kubernetes/perf-tests/blob/master/clusterloader2/docs/GETTING_STARTED.md#clusterloader2
