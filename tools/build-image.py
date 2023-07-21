#!/usr/bin/env python3
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

import getopt
import getpass
import json
import os
import shutil
import subprocess

import sys

# Supported host architectures for executables and docker images
# Mapped to the correct settings in the Makefile
architecture = {"x86_64": "amd64",
                "aarch64": "arm64"}
# Make targets for the shim repo to generate the images
targets = {"adm_image": "admission",
           "sched_image": "scheduler",
           "plugin_image": "scheduler-plugin"}
# registry setting passed to Makefile to allow testing of the script
repository = "apache"
# authentication info for docker hub
docker_user = ""
docker_pass = ""
docker_token = ""


# fail the execution
def fail(message):
    print(message)
    sys.exit(1)


# get the command from the path
def get_cmd(name):
    cmd = shutil.which(name)
    if not cmd:
        fail("command not found on the path: '%s'" % name)
    return cmd


# load the config, based on the build-release.py code.
def load_config():
    tools_dir = os.path.dirname(os.path.realpath(__file__))
    # load configs
    config_file = os.path.join(tools_dir, "release-configs.json")
    with open(config_file) as configs:
        try:
            data = json.load(configs)
        except json.JSONDecodeError:
            fail("load config: unexpected json decode failure")

    if "release" not in data:
        fail("load config: release data not found")
    release_meta = data["release"]
    if "version" not in release_meta:
        fail("load config: version data not found in release")
    version = release_meta["version"]
    release_package_name = "apache-yunikorn-{0}-src".format(version)
    if "repositories" not in data:
        fail("load config: repository list not found")
    repo_list = data["repositories"]

    staging_dir = os.path.join(os.path.dirname(tools_dir), "staging")
    release_base = os.path.join(staging_dir, release_package_name)

    print("release meta info:")
    print(" - version:        %s" % version)
    print(" - base directory: %s" % release_base)
    print(" - package name:   %s" % release_package_name)

    if not os.path.exists(release_base):
        fail("Staged release dir does not exist:\n\t%s" % release_base)
    return version, repo_list, release_base


# Cleanup image tag
def remove_tag(image_name):
    splits = image_name.split(":")
    if len(splits) != 2:
        fail("Image name is not in the required format")
    cmd = get_cmd("curl")
    curl = [cmd, "-X", "DELETE", "-H", "Authorization: JWT " + docker_token]
    curl.extend(["https://hub.docker.com/v2/repositories/" + splits[0] + "/tags/" + splits[1] + "/"])
    retcode = subprocess.call(curl)
    if retcode:
        fail("docker tag cleanup failed")


# Push an image or manifest
def push_image(cmd, image_name):
    push = [cmd, "push", image_name]
    retcode = subprocess.call(push, stdout=subprocess.DEVNULL)
    if retcode:
        fail("docker push failed")


# get token for rest
# 2FA is not supported by this code (yet) but can be added
def get_token():
    cmd = get_cmd("curl")
    curl = [cmd, "-X", "POST", "-H", "Content-Type: application/json"]
    curl.extend(["-d", '{"username": "' + docker_user + '", "password": "' + docker_pass + '"}'])
    curl.extend(["https://hub.docker.com/v2/users/login/"])
    p = subprocess.run(curl, capture_output=True)
    try:
        data = json.loads(p.stdout)
    except json.JSONDecodeError:
        fail("login failed: unexpected json decode failure: %s" %p.stdout)
    if "detail" in data:
        fail("authentication failed: %s" % data["detail"])
    if "token" not in data:
        fail("login failed: unexpected json content: %s" % data)
    global docker_token
    docker_token = data["token"]


# get user and password on startup
def get_auth():
    global docker_user
    docker_user = input("Enter docker hub username: ")
    global docker_pass
    docker_pass = getpass.getpass(prompt="Docker hub password: ", stream=None)
    if docker_pass == "" or docker_user == "":
        fail("username and password required")


# Login to docker
def login():
    cmd = get_cmd("docker")
    # login to docker
    print("Login to docker hub")
    log_in = [cmd, "login", "--username", docker_user, "--password", docker_pass]
    retcode = subprocess.call(log_in, stdout=subprocess.DEVNULL)
    if retcode:
        fail("docker login failed")
    get_token()


# Create an image name based on passed in details
def create_image_name(image, version, arch):
    image_name = repository + "/yunikorn:" + image
    if arch != "":
        image_name += "-" + arch
    image_name += "-" + version
    return image_name


# Create the manifest
def build_manifest(manifest, version):
    print("Building manifest")
    print(" - manifest: %s" % manifest)
    print(" - version:  %s" % version)
    multi_image = create_image_name(manifest, version, "")
    cmd = get_cmd("docker")
    command = [cmd, "manifest", "create", multi_image]
    for arch in architecture:
        image_name = create_image_name(manifest, version, architecture[arch])
        print(" - image:    %s" % image_name)
        # image_manifest = create_image_name(manifest, version, manifestmap[arch])
        # print(" - image manifest:    %s" % image_manifest)
        # temporary push to create tag to allow manifest build
        # https://github.com/docker/cli/issues/3350
        push_image(cmd, image_name)
        command.extend(["--amend", image_name])
    retcode = subprocess.call(command, stdout=subprocess.DEVNULL)
    if retcode:
        fail("docker manifest creation failed")
    # push the manifest
    # purge option is needed: https://github.com/docker/cli/issues/954
    command = [cmd, "manifest", "push", "--purge", multi_image]
    retcode = subprocess.call(command, stdout=subprocess.DEVNULL)
    if retcode:
        fail("docker manifest push failed")
    # remove temporary tags that allowed manifest build
    for arch in architecture:
        image_name = create_image_name(manifest, version, architecture[arch])
        remove_tag(image_name)


# Build a scheduler image
def build_image(base_dir, image, arch, version):
    cmd = get_cmd("make")
    my_env = os.environ.copy()
    my_env["QUIET"] = "--quiet"      # stop image build from being chatty
    my_env["VERSION"] = version      # force version, just be safe
    my_env["HOST_ARCH"] = arch       # the architecture override
    my_env["REGISTRY"] = repository  # repository override (test only)
    command = [cmd, "clean", image]
    # build the image using make
    retcode = subprocess.call(command, cwd=base_dir, env=my_env, stdout=subprocess.DEVNULL)
    if retcode:
        fail("make image failed")


# Build the web image
def web_image(base_dir, version):
    # build the images
    for arch in architecture:
        print("Building image for 'web', using 'image', architecture: '%s'" % arch)
        build_image(base_dir, "image", arch, version)
    # build the manifest
    build_manifest("web", version)


# Build the scheduler images
def scheduler_images(base_dir, version):
    # build the images for each target
    for target in targets:
        image = targets[target]
        # build all architectures
        for arch in architecture:
            print("Building image '%s' using: '%s', architecture: '%s'" % (image, target, arch))
            build_image(base_dir, target, arch, version)
        # build the manifest
        build_manifest(image, version)


# Build the combined architecture images
def build_images():
    get_auth()
    login()
    version, repo_list, release_base = load_config()
    for repo_meta in repo_list:
        if "name" not in repo_meta:
            fail("repository name missing in repo list")
        repo_name = repo_meta["name"]
        switcher = {
            "yunikorn-k8shim": scheduler_images,
            "yunikorn-web": web_image,
        }
        if switcher.get(repo_name) is not None:
            if "alias" not in repo_meta:
                fail("repository alias missing in repo list")
            alias = repo_meta["alias"]
            switcher.get(repo_name)(os.path.join(release_base, alias), version)


# Print the usage info
def usage(script):
    print("%s [--repository <name>]" % script)
    print("repository override should only be used for testing")
    sys.exit(2)


def main(argv):
    script = argv[0]
    try:
        opts, args = getopt.getopt(argv[1:], "", ["repository="])
    except getopt.GetoptError:
        usage(script)
    if args:
        usage(script)
    global repository
    for opt, arg in opts:
        if opt == "--repository":
            if not arg:
                usage(script)
        repository = arg
    build_images()


if __name__ == "__main__":
    main(sys.argv)
