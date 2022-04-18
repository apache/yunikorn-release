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

import distutils.dir_util
import getopt
import hashlib
import json
import os
import re
import shutil
import subprocess
import tarfile
from tempfile import mkstemp

import git
import sys


# Main build routine
def build_release(email_address):
    tools_dir = os.path.dirname(os.path.realpath(__file__))
    # load configs
    config_file = os.path.join(tools_dir, "release-configs.json")
    with open(config_file) as configs:
        data = json.load(configs)

    release_meta = data["release"]
    version = release_meta["version"]
    release_package_name = "apache-yunikorn-{0}-src".format(version)
    repo_list = data["repositories"]

    print("release meta info:")
    print(" - main version: %s" % version)
    print(" - release package name: %s" % release_package_name)

    staging_dir = os.path.join(os.path.dirname(tools_dir), "staging")
    release_base = os.path.join(staging_dir, release_package_name)
    release_top_path = os.path.join(os.path.dirname(tools_dir), "release-top-level-artifacts")
    helm_chart_path = os.path.join(os.path.dirname(tools_dir), "helm-charts")

    # setup artifacts in the release base dir
    setup_base_dir(release_top_path, helm_chart_path, release_base, version)

    # download source code from github repo
    sha = dict()
    for repo_meta in repo_list:
        sha[repo_meta["name"]] = download_sourcecode(release_base, repo_meta)
        update_make_version(repo_meta["name"], os.path.join(release_base, repo_meta["alias"]), version)

    # update the sha for all repos in the build scripts
    # must be run after all repos have been checked out
    update_sha(release_base, repo_list, sha)

    # build the helm package
    call_helm(staging_dir, release_base, version, email_address)

    # generate source code tarball
    tarball_name = release_package_name + ".tar.gz"
    tarball_path = os.path.join(staging_dir, tarball_name)
    print("creating tarball %s" % tarball_path)
    with tarfile.open(tarball_path, "w:gz") as tar:
        tar.add(os.path.join(release_base, "LICENSE"), arcname=release_package_name + "/LICENSE")
        tar.add(release_base, arcname=release_package_name, filter=exclude_files)
    write_checksum(tarball_path, tarball_name)
    if email_address:
        call_gpg(tarball_path, email_address)


# Function passed in as a filter to the tar command to keep tar as clean as possible
def exclude_files(tarinfo):
    file_name = os.path.basename(tarinfo.name)
    exclude = [".git", ".github", ".gitignore", ".travis.yml", ".asf.yaml", ".golangci.yml", ".helmignore", "LICENSE",
               "landmark"]
    if file_name in exclude:
        print("exclude file from tarball %s" % tarinfo.name)
        return None
    return tarinfo


# Setup base for the source code tar ball with release repo files
def setup_base_dir(release_top_path, helm_path, base_path, version):
    print("setting up base dir for release artifacts, path: %s" % base_path)
    if os.path.exists(base_path):
        print("\nstaging dir already exist:\n%s\nplease remove it and retry\n" % base_path)
        sys.exit(1)

    # setup base dir
    os.makedirs(base_path)
    # copy top level artifacts
    for file in os.listdir(release_top_path):
        org = os.path.join(release_top_path, file)
        dest = os.path.join(base_path, file)
        print("copying files: %s ===> %s" % (org, dest))
        shutil.copy2(org, dest)
    # set the base Makefile version info
    replace(os.path.join(base_path, "Makefile"), 'latest', version)
    # set the base validate_cluster version info
    replace(os.path.join(base_path, "validate_cluster.sh"), 'latest', version)
    # update the version tags in the README
    replace(os.path.join(base_path, "README.md"), '-latest', '-' + version)
    # copy the helm charts
    copy_helm_charts(helm_path, base_path, version)


# copy the helm charts into the base path and replace the version to the one defined in config
def copy_helm_charts(helm_path, base_path, version):
    print("helm patch: %s, base path: %s" % (helm_path, base_path))
    release_helm_path = os.path.join(base_path, "helm-charts")
    distutils.dir_util.copy_tree(helm_path, release_helm_path)
    # rename the version in the helm charts to the actual version
    yunikorn_chart_path = os.path.join(release_helm_path, "yunikorn")
    replace(os.path.join(yunikorn_chart_path, "values.yaml"), 'tag: scheduler-.*', 'tag: scheduler-' + version)
    replace(os.path.join(yunikorn_chart_path, "values.yaml"), 'tag: admission-.*', 'tag: admission-' + version)
    replace(os.path.join(yunikorn_chart_path, "values.yaml"), 'tag: web-.*', 'tag: web-' + version)
    replace(os.path.join(yunikorn_chart_path, "Chart.yaml"), 'version: .*', 'version: ' + version)
    replace(os.path.join(yunikorn_chart_path, "Chart.yaml"), 'appVersion: .*', 'appVersion: \"' + version + '\"')


# replaces the string that match to pattern to subst from the file_path
def replace(file_path, pattern, subst):
    # Create temp file
    fh, abs_path = mkstemp()
    with os.fdopen(fh, 'w') as new_file:
        with open(file_path) as old_file:
            for line in old_file:
                new_line = re.sub(pattern, subst, line)
                new_file.write(new_line)
    # Copy the file permissions from the old file to the new file
    shutil.copymode(file_path, abs_path)
    # Remove original file
    os.remove(file_path)
    # Move new file
    shutil.move(abs_path, file_path)


def download_sourcecode(base_path, repo_meta):
    print("downloading source code")
    print("repository info:")
    print(" - repository: %s " % repo_meta["repository"])
    print(" - description: %s " % repo_meta["description"])
    print(" - tag: %s " % repo_meta["tag"])
    repo = git.Repo.clone_from(
        url=repo_meta["repository"],
        to_path=os.path.join(base_path, repo_meta["alias"]))
    repo.git.checkout(repo_meta["tag"])
    tag = repo.tag("refs/tags/" + repo_meta["tag"])
    sha = tag.commit.hexsha
    print(" - tag sha: %s " % sha)

    # avoid pulling dependencies from github,
    # add replace to go mod files to make sure it builds locally
    update_dep_ref(repo_meta["name"], os.path.join(base_path, repo_meta["alias"]))
    return sha


# K8shim depends on yunikorn-core and scheduler-interface
def update_dep_ref_k8shim(local_repo_path):
    print("updating dependency for k8shim")
    mod_file = os.path.join(local_repo_path, "go.mod")
    with open(mod_file, "a") as file_object:
        file_object.write("\n")
        file_object.write("replace github.com/apache/yunikorn-core => ../core \n")
        file_object.write(
            "replace github.com/apache/yunikorn-scheduler-interface => ../scheduler-interface \n")


# core depends on scheduler-interface
def update_dep_ref_core(local_repo_path):
    print("updating dependency for core")
    mod_file = os.path.join(local_repo_path, "go.mod")
    with open(mod_file, "a") as file_object:
        file_object.write("\n")
        file_object.write(
            "replace github.com/apache/yunikorn-scheduler-interface => ../scheduler-interface \n")


# update go mod in the repos
def update_dep_ref(repo_name, local_repo_path):
    switcher = {
        "yunikorn-k8shim": update_dep_ref_k8shim,
        "yunikorn-core": update_dep_ref_core,
    }
    if switcher.get(repo_name) is not None:
        switcher.get(repo_name)(local_repo_path)


# replace the default version to release version in the Makefile(s)
def update_make_version(repo_name, local_repo_path, version):
    switcher = {
        "yunikorn-k8shim": "update",
        "yunikorn-web": "update",
    }
    if switcher.get(repo_name) is not None:
        replace(os.path.join(local_repo_path, "Makefile"), 'latest', version)


# k8shim uses its own, yunikorn-core and scheduler-interface revisions
def update_sha_shim(repo_name, local_repo_path, sha):
    print("updating sha for k8shim")
    make_file = os.path.join(local_repo_path, "Makefile")
    replace(make_file, 'coreSHA=.*\\)', 'coreSHA=' + sha["yunikorn-core"])
    replace(make_file, 'siSHA=.*\\)', 'siSHA=' + sha["yunikorn-scheduler-interface"])
    replace(make_file, 'shimSHA=.*\\)', 'shimSHA=' + sha[repo_name])


# web only uses its own revision
def update_sha_web(repo_name, local_repo_path, sha):
    print("updating sha for web")
    replace(os.path.join(local_repo_path, "Makefile"), "SHA=.*\\)", 'SHA=' + sha[repo_name])


# update git revision in the makefiles
def update_sha(release_base, repo_list, sha):
    for repo_meta in repo_list:
        repo_name = repo_meta["name"]
        switcher = {
            "yunikorn-k8shim": update_sha_shim,
            "yunikorn-web": update_sha_web,
        }
        if switcher.get(repo_name) is not None:
            switcher.get(repo_name)(repo_name, os.path.join(release_base, repo_meta["alias"]), sha)


# Write the checksum for the source code tarball to file
def write_checksum(tarball_file, tarball_name):
    print("generating sha512 checksum file for tar")
    h = hashlib.sha512()
    # read the file and generate the sha
    with open(tarball_file, 'rb') as file:
        while True:
            data = file.read(65536)
            if not data:
                break
            h.update(data)
    sha = h.hexdigest()
    # write out the checksum
    sha_file = open(tarball_file + ".sha512", "w")
    sha_file.write(sha)
    sha_file.write("  " + tarball_name)
    sha_file.write("\n")
    sha_file.close()
    print("sha512 checksum: %s" % sha)


# Sign the source archive if an email is provided
def call_gpg(tarball_file, email_address):
    cmd = shutil.which("gpg")
    if not cmd:
        print("gpg not found on the path, not signing package")
        return
    print("Signing source code file using email: %s" % email_address)
    subprocess.call([cmd,
                     '--local-user',
                     email_address,
                     '--armor',
                     '--output',
                     tarball_file + ".asc",
                     '--detach-sig',
                     tarball_file])


# Package the helm chart and sign if an email is provided
def call_helm(staging_dir, base_path, version, email_address):
    cmd = shutil.which("helm")
    if not cmd:
        print("helm not found on the path, not creating package")
        return
    release_helm_path = os.path.join(base_path, "helm-charts/yunikorn")
    command = [cmd, 'package']
    if email_address:
        secring = os.path.expanduser('~/.gnupg/secring.gpg')
        if os.path.isfile(secring):
            print("Packaging helm chart, signed with: %s" % email_address)
            command.extend(['--sign', '--key', email_address, '--keyring', secring])
        else:
            print("Packing helm chart (unsigned)\nFile with pre gpg2 keys not found, expecting: %s" % secring)
            email_address = None
    else:
        print("Packaging helm chart (unsigned)")
    command.extend([release_helm_path, '--destination', staging_dir])
    subprocess.call(command)
    if not email_address:
        helm_package = "yunikorn-" + version + ".tgz"
        helm_pack_path = os.path.join(staging_dir, helm_package)
        h = hashlib.sha256()
        # read the file and generate the sha
        with open(helm_pack_path, 'rb') as file:
            while True:
                data = file.read(65536)
                if not data:
                    break
                h.update(data)
        print("Helm package digest: %s  %s\n" % (h.hexdigest(), helm_package))


# Print the usage info
def usage(script):
    print("%s [--sign=<email>]" % script)
    sys.exit(2)


def main(argv):
    script = sys.argv[0]
    email_address = ''
    try:
        opts, args = getopt.getopt(argv[1:], "", ["sign="])
    except getopt.GetoptError:
        usage(script)

    if args:
        usage(script)

    for opt, arg in opts:
        if opt == "--sign":
            if not arg:
                usage(script)
        email_address = arg

    build_release(email_address)


if __name__ == "__main__":
    main(sys.argv)
