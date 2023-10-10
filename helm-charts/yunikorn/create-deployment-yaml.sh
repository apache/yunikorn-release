#!/bin/bash
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