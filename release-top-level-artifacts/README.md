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

# Apache YuniKorn
Apache YuniKorn is a light-weight, universal resource scheduler for container orchestrator systems.
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
* Go 1.16 or later

### Yunikorn web UI
The YuniKorn web UI uses a two stage docker build with predefined images.
All dependencies are included in the image.

NOTE: the scheduler can be used without a web UI build or deployed.

## Building
Run the `make` command to build docker images.

```shell script
make
```
The local build only generates an image for one processor architecture.
The default for the architecture is the local processor type that is retrieved via the `uname -m` command.
The architecture can be overridden by setting the shell variable `HOST_ARCH`.
For details on how to change the architectures see the processing of `HOST_ARCH` in the `Makefile` included in the `k8shim` directory.

The command will generate the following four docker images in the local docker repository:
* apache/yunikorn:scheduler-_amd64_-latest
* apache/yunikorn:scheduler-plugin-_amd64_-latest
* apache/yunikorn:admission-_amd64_-latest
* apache/yunikorn:web-_amd64_-latest

Note: the naming of the images assumes a processor architecture that maps to the docker architecture _amd64_.

## Verifying the release
A script and configuration to create a simple cluster using the locally built images is provided in this release archive.
Follow the instructions in [Building](#building) to create the local docker images.
The script for validating the release and the build use the same defaults.

After the images have been created run the script for more instructions and to list the tools required for validating
the release:
```shell
./validate_cluster.sh
```
The `kind` cluster created using the included config is a small, but fully functional Kubernetes cluster, with
Apache YuniKorn deployed.

## Deploying YuniKorn
The simplest way to run Apache YuniKorn is to use the provided helm charts, you can find the templates in the release
package `helm-charts`.
There are a few prerequisites:
1. An existing K8s cluster is up and running.
2. Helm chart client is installed.

If you have a cluster, and the helm client you can simply run helm from the root directory:
```shell script
helm install yunikorn ./helm-charts/yunikorn
```

## Customising the build
The `make` command will pass on the following variables:
* VERSION
* REGISTRY
* HOST_ARCH
These variables can be used to generate customised build:
```shell script
HOST_ARCH="arm64" VERSION="1.1.0" REGISTRY="internal" make
```

These same variables can be set and will be picked up by the validate_cluster script.

The values defined in the helm charts assume a default build without changes to the variables.
Once you have built your own docker images, you will need to replace the docker image name in the helm chart templates.
Open `helm-charts/yunikorn/values.yaml` and replace the docker image information with ones you built.

For more instructions, please refer to [User Guide](https://yunikorn.apache.org/docs/).

## Deploying a convenience build
Apache YuniKorn provides a convenience release with pre-built Docker images and Helm charts.
These can be accessed via the [downloads page](https://yunikorn.apache.org/community/download) and instructions are
located in the [User Guide](https://yunikorn.apache.org/docs/).

The convenience build images are multi-architecture images. Supported architectures are `amd64` and `arm64v8`.

## Reproducible builds
Building YuniKorn from source generates reproducible build artifacts which
depend only on the version of YuniKorn built and the go compiler version used.

This release was compiled by the official release manager using Go version `@GO_VERSION@`
and generated binary artifacts with the following SHA-512 checksums:

### linux/amd64
```
@AMD64_BINARIES@
```

### linux/arm64v8
```
@ARM64_BINARIES@
```

## Testing the build
Running the unit tests is supported via the make command.
It will run the tests for all parts of YuniKorn in order:
```shell script
make test
```

### Yunikorn Scheduler
The scheduler tests runs in two parts: the core, and the k8shim.
There are no tests for the scheduler-interface.

Unit testing for the scheduler has no additional pre-requisites.

### Yunikorn web UI
The project requires a number of external tools to be installed for test and development.
A non image build requires the following tools to be installed:
* Node.js 16.14.2
* Angular CLI 13.3.0
* yarn 1.22

Running unit tests adds the following requirements:
* Karma
* json-server

Please check the [documentation](https://yunikorn.apache.org/docs/) for further details.

