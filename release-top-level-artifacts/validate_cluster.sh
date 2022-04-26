#!/usr/bin/env bash
#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# config check
function config_check() {
  FAIL=0
  if [ ! -r "${KIND_CONFIG}" ]; then
    echo "  kind config not found"
    FAIL=1
  fi
  if [ ! -d "${HELMCHART}" ]; then
    echo "  helm chart directory not found"
    FAIL=1
  fi
  return ${FAIL}
}

# docker check
function docker_check() {
  DOCKER_UP=`docker version | grep "^Server:"`
  if [ -z "${DOCKER_UP}" ]; then
    echo "docker daemon must be running"
    return 1
  fi
}

# tool check
function tool_check() {
  FAIL=0
  if ! command -v kind &> /dev/null
  then
    echo "  kind must be installed and on the path"
    FAIL=1
  fi
  if ! command -v kubectl &> /dev/null
  then
    echo "  kubectl must be installed and on the path"
    FAIL=1
  fi
  if ! command -v helm &> /dev/null
  then
    echo "  helm must be installed and on the path"
    FAIL=1
  fi
  if ! command -v docker &> /dev/null
  then
    echo "  docker must be installed and on the path"
    FAIL=1
  fi
  return ${FAIL}
}

# show run details 
function run_detail() {
  echo "Creating kind test cluster"
  echo "  Apache YuniKorn version: ${VERSION}"
  echo "  helm chart directory:    ${HELMCHART}"
  echo "  kind cluster config:     ${KIND_CONFIG}"
  echo "  kubernetes image:        ${KIND_IMAGE}"
  echo "  Registry name:           ${REGISTRY}"
}

# usage message
function usage() {
  NAME=`basename "$0"`
  echo "You must enter exactly 1 command line argument"
  echo "  ${NAME} K8s-VERSION"
  echo "K8s-VERSION: the numeric version of the K8s release, example: 1.22.4"
  echo
  echo "Overrides for settings via shell variables"
  echo "  REGISTRY=local ${NAME} 1.22.4"
  echo
  echo "variable names with default values:"
  echo "  VERSION,      default: latest"
  echo "  REGISTRY,     default: 'apache'"
  echo "  KIND_CONFIG,  default: './kind.yaml'"
  echo "  HELMCHART,    default: './helm-charts/yunikorn'"
}

# remove kind cluster ion failure
function remove_cluster() {
	echo "Removing kind cluster"
	kind delete cluster --name yk8s
	exit 1
}

# print message on how to cleanup cluster
function cleanup() {
	echo
	echo "To clean up the cluster after use execute the following command:"
	echo "  kind delete cluster --name yk8s"
	echo
}

# tool check: run before input check to make sure all tools are available
tool_check
if [ $? -eq 1 ]; then
  echo "tool check failed please install required tool(s) before continuing." 
  echo
  usage
  exit 1
fi

# input check need at least the versions
if [ $# -ne 1 ]; then
  usage
  exit 1
fi

K8S=$1
KIND_IMAGE=kindest/node:v${K8S}
VERSION="${VERSION:-latest}"
REGISTRY="${REGISTRY:-apache}"
KIND_CONFIG="${KIND_CONFIG:-./kind.yaml}"
HELMCHART="${HELMCHART:-./helm-charts/yunikorn}"

# show details for the run
run_detail

# check if docker is up prevents kind failures
docker_check
if [ $? -eq 1 ]; then
  exit 1
fi

# check the configs
config_check
if [ $? -eq 1 ]; then
  exit 1
fi

kind create cluster --name yk8s --image ${KIND_IMAGE} --config=${KIND_CONFIG}
if [ $? -eq 1 ]; then
  exit 1
fi
echo
echo "Pre-Loading docker images..."
echo
kind load docker-image ${REGISTRY}/yunikorn:admission-${VERSION} --name yk8s >/dev/null 2>&1
if [ $? -eq 1 ]; then
	echo "Pre-Loading Admission Controller image failed, aborting"
  remove_cluster
fi
kind load docker-image ${REGISTRY}/yunikorn:scheduler-${VERSION} --name yk8s >/dev/null 2>&1
if [ $? -eq 1 ]; then
	echo "Pre-Loading scheduler image failed, aborting"
  remove_cluster
fi
kind load docker-image ${REGISTRY}/yunikorn:web-${VERSION} --name yk8s >/dev/null 2>&1
if [ $? -eq 1 ]; then
	echo "Pre-Loading web image failed, aborting"
  remove_cluster
fi

kubectl config use-context kind-yk8s
if [ $? -eq 1 ]; then
	echo "Kubernetes context switch failed, aborting"
  remove_cluster
fi

kubectl create namespace yunikorn
if [ $? -eq 1 ]; then
	echo "Namespace creation failed, aborting"
  remove_cluster
fi
echo
echo "Deploying helm chart..."
helm install yunikorn ${HELMCHART} --namespace yunikorn \
    --set image.repository=${REGISTRY}/yunikorn \
    --set image.tag=scheduler-${VERSION} \
    --set image.pullPolicy=IfNotPresent \
    --set admissionController.image.repository=${REGISTRY}/yunikorn \
    --set admissionController.image.tag=admission-${VERSION} \
    --set admissionController.image.pullPolicy=IfNotPresent \
    --set web.image.repository=${REGISTRY}/yunikorn \
    --set web.image.tag=web-${VERSION} \
    --set web.image.pullPolicy=IfNotPresent
echo
echo "Waiting for helm deployment to finish..."
kubectl wait --for=condition=available --timeout=150s deployment/yunikorn-scheduler -n yunikorn
if [ $? -eq 1 ]; then
	cleanup
  exit 1
fi

kubectl wait --for=condition=ready --timeout=150s pod -l app=yunikorn -n yunikorn
if [ $? -eq 1 ]; then
	cleanup
  exit 1
fi

echo
echo "Setting up port forwarding for REST (9080) and web UI (9889)"
kubectl port-forward svc/yunikorn-service 9889:9889 -n yunikorn >/dev/null 2>&1 &
kubectl port-forward svc/yunikorn-service 9080:9080 -n yunikorn >/dev/null 2>&1 &
cleanup
