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
import re
import shutil
import subprocess
import sys
from typing import AnyStr

# Supported host architectures for executables and docker images
# Mapped to the correct settings in the Makefile
ARCHITECTURE: dict[str,str] = {"x86_64": "amd64",
                               "aarch64": "arm64"}
# Make targets for the shim repo to generate the images
TARGETS: dict[str,str] = {"adm_image": "admission",
                          "sched_image": "scheduler"}
# registry setting passed to Makefile to allow testing of the script
_repository: str = "apache"
# Docker Hub host - images are always tagged/pushed fully-qualified under this host so
# podman doesn't rewrite an unqualified name to "localhost/..." in its local store.
DOCKER_HUB: str = "docker.io"
# authentication info for docker hub
docker_user: str  = ""
docker_pass: str  = ""
docker_token: str = ""
# container engine to use (docker or podman)
_engine: str = ""


def get_engine() -> str:
    global _engine
    if _engine:
        return _engine
    if "DOCKER" in os.environ and os.environ["DOCKER"]:
        _engine = os.environ["DOCKER"]
        return _engine
    if shutil.which("docker"):
        _engine = "docker"
    elif shutil.which("podman"):
        _engine = "podman"
    else:
        fail("neither docker nor podman found on the path")
    return _engine


# fail the execution
def fail(message: str):
    print(message)
    sys.exit(1)


def ensure_str(val: AnyStr, encoding: str = "utf-8") -> str:
    if isinstance(val, bytes):
        return val.decode(encoding)
    return val


# get the command from the path
def get_cmd(name: str) -> str:
    cmd = shutil.which(name)
    if not cmd:
        fail("command not found on the path: '%s'" % name)
    return str(cmd)


# Determine the specific go compiler installed (for logging to compare with repro version)
def get_go_version() -> str:
    command = ['go', 'env', 'GOVERSION']
    result = subprocess.run(command, capture_output=True)
    if result.returncode:
        fail("failed to get go version")
    output = re.sub('^go', '', str(result.stdout.strip(), 'utf-8'))
    return output


# Determine the go repro compiler version
def get_repro_version(base: str) -> str:
    repro = os.path.join(base, '.go_repro_version')
    if not os.path.isfile(repro):
        fail("go_repro_version file is missing")
    with open(repro, 'r') as file:
        content = file.readline().strip()
    return content


# load the config, based on the build-release.py code.
def load_config() -> tuple[str, dict, str]:
    tools_dir = ensure_str(os.path.dirname(os.path.realpath(__file__)))
    # load configs
    config_file = os.path.join(tools_dir, "release-configs.json")
    if not os.path.isfile(config_file):
        fail("release-configs.json file is missing")
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

    staging_dir = os.path.join((os.path.dirname(tools_dir)), "build", "staging")
    release_base = os.path.join(staging_dir, release_package_name)

    print("release meta info:")
    print(" - version:        %s" % version)
    print(" - base directory: %s" % release_base)
    print(" - package name:   %s" % release_package_name)
    print(" - go version:     %s" % get_go_version())

    if not os.path.exists(release_base):
        fail("Staged release dir does not exist:\n\t%s" % release_base)
    return version, repo_list, release_base


# Cleanup image tag
def remove_tag(image_name: str):
    # the Docker Hub API path is namespace/repo, not a host-qualified reference
    name = image_name
    if name.startswith(DOCKER_HUB + "/"):
        name = name[len(DOCKER_HUB) + 1:]
    splits = name.split(":")
    if len(splits) != 2:
        fail("Image name is not in the required format")
    cmd = get_cmd("curl")
    curl = [cmd, "-X", "DELETE", "-H", "Authorization: JWT " + docker_token]
    curl.extend(["https://hub.docker.com/v2/repositories/" + splits[0] + "/tags/" + splits[1] + "/"])
    result = subprocess.run(curl, capture_output=True)
    # Access the standard output and standard error
    if result.returncode:
        print("Output:", result.stdout)
        print("Errors:", result.stderr)
        fail("docker tag cleanup failed")


# Push an image or manifest
def push_image(cmd: str, image_name: str):
    push = [cmd, "push", image_name]
    result = subprocess.run(push, capture_output=True)
    # Access the standard output and standard error
    if result.returncode:
        print("Output:", result.stdout)
        print("Errors:", result.stderr)
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


# Login to docker registry
def login():
    engine = get_engine()
    cmd = get_cmd(engine)
    # login to docker hub
    print("Login to docker hub")
    log_in = [cmd, "login", "docker.io","--username", docker_user, "--password", docker_pass]
    result = subprocess.run(log_in, capture_output=True)
    # Access the standard output and standard error
    if result.returncode:
        print("Output:", result.stdout)
        print("Errors:", result.stderr)
        fail("%s login failed" % engine)
    get_token()


# Create an image name based on passed in details
def create_image_name(image: str, version: str, arch: str) -> str:
    image_name = DOCKER_HUB + "/" + _repository + "/yunikorn:" + image
    if arch != "":
        image_name += "-" + arch
    image_name += "-" + version
    return image_name


# Create the manifest
def build_manifest(manifest: str, version: str):
    print("Building manifest")
    print(" - manifest: %s" % manifest)
    print(" - version:  %s" % version)
    multi_image = create_image_name(manifest, version, "")
    engine = get_engine()
    cmd = get_cmd(engine)
    command = [cmd, "manifest", "create", multi_image]
    for arch in ARCHITECTURE:
        image_name = create_image_name(manifest, version, ARCHITECTURE[arch])
        print(" - image:    %s" % image_name)
        # image_manifest = create_image_name(manifest, version, manifestmap[arch])
        # print(" - image manifest:    %s" % image_manifest)
        # temporary push to create tag to allow manifest build
        # https://github.com/docker/cli/issues/3350
        push_image(cmd, image_name)
        command.extend(["--amend", image_name])
    result = subprocess.run(command, capture_output=True)
    # Access the standard output and standard error
    if result.returncode:
        print("Output:", result.stdout)
        print("Errors:", result.stderr)
        fail("%s manifest creation failed" % engine)
    # push the manifest
    # purge option is needed for docker: https://github.com/docker/cli/issues/954
    # podman uses --rm to remove local manifest list after push
    purge_flag = "--rm" if engine == "podman" else "--purge"
    command = [cmd, "manifest", "push", purge_flag, multi_image]
    result = subprocess.run(command, capture_output=True)
    # Access the standard output and standard error
    if result.returncode:
        print("Output:", result.stdout)
        print("Errors:", result.stderr)
        fail("%s manifest push failed" % engine)
    # remove temporary tags that allowed manifest build
    for arch in ARCHITECTURE:
        image_name = create_image_name(manifest, version, ARCHITECTURE[arch])
        remove_tag(image_name)


# Build a scheduler image
def build_image(base_dir: str, image: str, arch: str, version: str):
    # move .gitignore in the staging dirs so repro version builds work
    git_ignore = os.path.join(base_dir, ".gitignore")
    shutil.move(git_ignore, git_ignore+".tmp")
    cmd = get_cmd("make")
    my_env = os.environ.copy()
    my_env["QUIET"] = "--quiet"          # stop image build from being chatty
    my_env["VERSION"] = version          # force version, just be safe
    my_env["HOST_ARCH"] = arch           # the architecture override
    my_env["REPRODUCIBLE_BUILDS"] = "1"  # always use reproducible builds
    my_env["REGISTRY"] = DOCKER_HUB + "/" + _repository  # fully-qualified, avoids podman's localhost/ rewrite
    my_env["DOCKER"] = get_engine()      # pass container engine to make
    command = [cmd, "clean", image]
    # build the image using make
    result = subprocess.run(command, cwd=base_dir, env=my_env, capture_output=True)

    # move .gitignore back
    shutil.move(git_ignore+".tmp", git_ignore)
    # Access the standard output and standard error
    if result.returncode:
        print("Output:", result.stdout)
        print("Errors:", result.stderr)
        fail("make image failed")


# Build the web image
def web_image(base_dir: str, version: str):
    # build the images
    for arch in ARCHITECTURE:
        print("Building image for 'web', using 'image', architecture: '%s'" % arch)
        build_image(base_dir, "image", arch, version)
    # build the manifest
    build_manifest("web", version)


# Build the scheduler images
def scheduler_images(base_dir: str, version: str):
    # build the images for each target
    for target in TARGETS:
        image = TARGETS[target]
        # build all architectures
        for arch in ARCHITECTURE:
            print("Building image '%s' using: '%s', architecture: '%s'" % (image, target, arch))
            print(" - go repro version: %s" % get_repro_version(base_dir))
            build_image(base_dir, target, arch, version)
        # build the manifest
        build_manifest(image, version)


# Build the combined architecture images
def build_images():
    print("Using container engine: %s" % get_engine())
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
def usage(script: str):
    print("%s [--repository <name>] [--engine <name>]" % script)
    print("repository override should only be used for testing")
    print("engine: container engine to use (default: auto-detected, docker preferred)")
    sys.exit(2)


def main(argv: list[str]):
    script = argv[0]
    try:
        opts, args = getopt.getopt(argv[1:], "", ["repository=", "engine="])
    except getopt.GetoptError:
        usage(script)
    if args:
        usage(script)
    global _repository, _engine
    for opt, arg in opts:
        if opt == "--repository":
            if not arg:
                usage(script)
            _repository = arg
        if opt == "--engine":
            if not arg:
                usage(script)
            _engine = arg
    build_images()


if __name__ == "__main__":
    main(sys.argv)
