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

import json
import os
import shutil
import git
import tarfile
import distutils.dir_util

def build_release():
    tools_dir = os.path.dirname(os.path.realpath(__file__))

    # load configs
    config_file = os.path.join(tools_dir, "release-configs.json")
    with open (config_file) as configs:
        data = json.load(configs)

    release_meta = data["release"]
    release_package_name = "apache-yunikorn-{0}-incubating-src".format(release_meta["version"])
    repo_list = data["repositories"]

    print("release meta info:")
    print(" - main version: %s" % release_meta["version"])
    print(" - release cadidate: %s" % release_meta["release-candidate-version"])
    print(" - release package name: %s" % release_package_name)

    staging_dir = os.path.join(os.path.dirname(tools_dir), "staging")
    release_base = os.path.join(staging_dir, release_package_name)
    release_top_path = os.path.join(os.path.dirname(tools_dir),
        "release-top-level-artifacts")
    helm_chart_path = os.path.join(os.path.dirname(tools_dir),
        "helm-charts")

    # setup artifacts in the relase base dir
    setup_base_dir(release_top_path, helm_chart_path, release_base)

    # download source code from github repo
    for repo_meta in repo_list:
        dowload_sourcecode(release_base, repo_meta)

    # generate tarball
    tarball_name = release_package_name + ".tar.gz"
    tarball_path = os.path.join(staging_dir, tarball_name)
    print("creating tarball %s" % tarball_path)
    with tarfile.open(tarball_path, "w:gz") as tar:
        tar.add(release_base, arcname=release_package_name, filter=exclude_files)
        tar.addfile(tarfile.TarInfo("LICENSE"), open(os.path.join(release_top_path, "LICENSE")))

def exclude_files(tarinfo):
    file_name = os.path.basename(tarinfo.name)
    exclude = [".git", ".github", ".gitignore", ".asf.yaml", ".golangci.yml", ".helmignore", "LICENSE"]
    if file_name in exclude:
        print("exclude file from tarball %s" % tarinfo.name)
        return None
    return tarinfo

def setup_base_dir(release_top_path, helm_path, base_path):
    print("setting up base dir for release artifacts, path: %s" % base_path)
    if os.path.exists(base_path):
      raise Exception("staging dir %s already exist, please remove it and retry" % base_path)

    # setup base dir
    os.makedirs(base_path)
    # copy top level artifacts
    for file in os.listdir(release_top_path):
        org = os.path.join(release_top_path, file)
        dest = os.path.join(base_path, file)
        print("copying files: %s ===> %s" % (org, dest))
        shutil.copy2(org, dest)
    # copy helm charts
    print("helm patch: %s, base path: %s", helm_path, base_path)
    distutils.dir_util.copy_tree(helm_path, os.path.join(base_path, "helm-charts"))

def dowload_sourcecode(base_path, repo_meta):
    print("downloading source code")
    print("repository info:")
    print(" - repository: %s " % repo_meta["repository"])
    print(" - description: %s " % repo_meta["description"])
    print(" - tag: %s " % repo_meta["tag"])
    repo = git.Repo.clone_from(
        url=repo_meta["repository"],
        to_path=os.path.join(base_path, repo_meta["alias"]))
    repo.git.checkout(repo_meta["tag"])

    # avoid pulling dependencies from github,
    # add replace to go mod files to make sure it builds locally
    update_dep_ref(repo_meta["name"], os.path.join(base_path, repo_meta["alias"]))

def update_dep_ref_k8shim(local_repo_path):
    print("updating dependency for k8shim")
    # K8shim is depdending on yunikorn-core and scheduler-interface
    mod_file = os.path.join(local_repo_path, "go.mod")
    with open(mod_file, "a") as file_object:
        file_object.write("\n")
        file_object.write("replace github.com/apache/incubator-yunikorn-core => ../core \n")
        file_object.write("replace github.com/apache/incubator-yunikorn-scheduler-interface => ../scheduler-interface \n")

def update_dep_ref_core(local_repo_path):
    print("updating dependency for core")
    mod_file = os.path.join(local_repo_path, "go.mod")
    with open(mod_file, "a") as file_object:
        file_object.write("\n")
        file_object.write("replace github.com/apache/incubator-yunikorn-scheduler-interface => ../scheduler-interface \n")

def update_dep_ref(repo_name, local_repo_path):
    switcher={
        "yunikorn-k8shim":update_dep_ref_k8shim,
        "yunikorn-core":update_dep_ref_core,
    }
    if switcher.get(repo_name) is not None:
        switcher.get(repo_name)(local_repo_path)

def main():
    build_release()

if __name__ == "__main__":
    main()
