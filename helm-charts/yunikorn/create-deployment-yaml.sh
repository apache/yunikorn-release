#!/bin/bash
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

# This file is used to generate some YAML configuration files
# which can be directly used by k8s without the need for helm.

mkdir plugin
mkdir deployment
helm template yunikorn . -f values.yaml --output-dir ./deployment
helm template yunikorn ./ -f values.yaml --set enableSchedulerPlugin=true --output-dir ./plugin
mv ./plugin/yunikorn/templates/deployment.yaml ./deployment/yunikorn/templates/plugin.yaml
rm -r plugin

folder_path="./deployment/yunikorn/templates"
target_words=("Helm" "helm" "chart" "annotations" "release")

for file in "$folder_path"/*; do
  if [ -f "$file" ]; then
    for word in "${target_words[@]}"; do
      grep -v "$word" "$file" > "$file.tmp" && mv "$file.tmp" "$file"
    done
    mv "$file" "./deployment/"
  fi
done

rm -r ./deployment/yunikorn