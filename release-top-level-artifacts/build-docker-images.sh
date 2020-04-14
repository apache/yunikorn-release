#!/bin/bash

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

DOCKER_REPOSITORY_NAME="yunikorn"
DOCKER_IMAGE_VERSION="latest"

display_help() {
    echo "Usage: $0 [option...]" >&2
    echo
    echo "   -r, --repository           docker repository name"
    echo "   -v, --version              docker image version"
    echo
    exit 1
}

while :
do
  case "$1" in
    -r | --repository)
      if [ $# -ne 0 ]; then
        DOCKER_REPOSITORY_NAME="$2"
      fi
      shift 2
      ;;
    -v | --version)
      DOCKER_IMAGE_VERSION="$2"
      shift 2
      ;;
    -h | --help)
      display_help
      exit 0
      ;;
    --)
      shift
      break
      ;;
    -*)
      echo "Error: Unknown option: $1" >&2
      display_help
      exit 1
      ;;
    *)
      break
      ;;
    esac
done

echo "start to build docker images for yunikorn"
echo "build options"
echo "  - repository: ${DOCKER_REPOSITORY_NAME}"
echo "  - version: ${DOCKER_IMAGE_VERSION}"

CURRENT_DIR="$( cd "$(dirname "$0")" >/dev/null 2>&1 || exit ; pwd -P )"
cd "${CURRENT_DIR}"/k8shim && \
  make image REGISTRY="${DOCKER_REPOSITORY_NAME}" \
  VERSION="${DOCKER_IMAGE_VERSION}" && \
  cd - || exit

cd "${CURRENT_DIR}"/web && \
  make image TAG="${DOCKER_REPOSITORY_NAME}"/yunikorn-web \
  VERSION="${DOCKER_IMAGE_VERSION}" && \
  cd - || exit
