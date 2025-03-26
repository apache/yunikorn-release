#!/usr/bin/env bash

# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Constants
SOAK_TEST_CLUSTER='soak-test-cluster'

# create a kind cluster
kind create cluster --name $SOAK_TEST_CLUSTER

# install YuniKorn scheduler on kind Cluster
helm repo add yunikorn https://apache.github.io/yunikorn-release
helm repo update
kubectl create namespace yunikorn
# TODO: allow to install a customized YuniKorn version to run the soak test
helm install yunikorn yunikorn/yunikorn --namespace yunikorn

# Deploy kwok in a Cluster
helm repo add kwok https://kwok.sigs.k8s.io/charts/
helm upgrade --namespace kube-system --install kwok kwok/kwok
helm upgrade --install kwok kwok/stage-fast

# Install Helm chart for autoscaler with Kwok provider
helm repo add autoscaler https://kubernetes.github.io/autoscaler
helm repo update
helm upgrade --install autoscaler autoscaler/cluster-autoscaler --set cloudProvider=kwok --set "autoDiscovery.clusterName"="kind-${SOAK_TEST_CLUSTER}" --set "extraArgs.enforce-node-group-min-size"=true

# Install the perf-tests repository and compile the clusterloader2 binary
git clone git@github.com:kubernetes/perf-tests.git
cd perf-tests/clusterloader2
go build -o clusterloader2 ./cmd
