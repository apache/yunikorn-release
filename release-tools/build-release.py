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


# fail the execution
def fail(message):
    print(message)
    sys.exit(1)


# Main build routine
def build_release(email_address):
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
        if "name" not in repo_meta:
            fail("repository name missing in repo list")
        name = repo_meta["name"]
        if "alias" not in repo_meta:
            fail("repository alias missing in repo list")
        alias = repo_meta["alias"]

        sha[name] = download_sourcecode(release_base, repo_meta)
        update_make_version(name, os.path.join(release_base, alias), version)

    # update required Golang, NodeJS and Angular versions
    update_required_versions(release_base, repo_list)

    # update the sha for all repos in the build scripts
    # must be run after all repos have been checked out
    update_sha(release_base, repo_list, sha)

    # merge licenses for anything that was added not part of Apache
    merge_licenses(release_base, repo_list)

    # build the helm package
    call_helm(staging_dir, release_base, version, email_address)

    # Ensure release dirs are clean (and generate build.date files)
    clean_release(release_base)

    # generate staging source code tarball
    tarball_name = release_package_name + "-staging.tar.gz"
    tarball_path = os.path.join(staging_dir, tarball_name)
    print("creating tarball %s" % tarball_path)
    with tarfile.open(tarball_path, "w:gz") as tar:
        tar.add(release_base, arcname=release_package_name, filter=exclude_files)

    # generate yunikorn-web reproducible binaries
    web_hashes_amd64 = build_web_and_generate_hashes(staging_dir, release_package_name, "x86_64")
    web_hashes_arm64 = build_web_and_generate_hashes(staging_dir, release_package_name, "aarch64")

    # generate yunikorn-k8shim reproducible binaries
    shim_hashes_amd64 = build_shim_and_generate_hashes(staging_dir, release_package_name, "x86_64")
    shim_hashes_arm64 = build_shim_and_generate_hashes(staging_dir, release_package_name, "aarch64")

    # merge hashes
    hashes_amd64 = "\n".join([shim_hashes_amd64, web_hashes_amd64])
    hashes_arm64 = "\n".join([shim_hashes_arm64, web_hashes_arm64])

    # remove staging tarball
    os.remove(tarball_path)

    # update reproducible build information in README
    go_version = get_go_version()
    update_reproducible_build_info(release_base, go_version, hashes_amd64, hashes_arm64)

    # generate final source code tarball
    tarball_name = release_package_name + ".tar.gz"
    tarball_path = os.path.join(staging_dir, tarball_name)
    print("creating tarball %s" % tarball_path)
    with tarfile.open(tarball_path, "w:gz") as tar:
        tar.add(release_base, arcname=release_package_name, filter=exclude_files)
    write_checksum(tarball_path, tarball_name)
    if email_address:
        call_gpg(tarball_path, email_address)


# Function passed in as a filter to the tar command to keep tar as clean as possible
def exclude_files(tarinfo):
    file_name = os.path.basename(tarinfo.name)
    exclude = [".DS_Store", ".git", ".github", ".gitignore", ".asf.yaml"]
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
    replace(os.path.join(yunikorn_chart_path, "values.yaml"), '(tag: .*-)(latest)', '\\g<1>' + version)
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
    # these two have been checked before we get here
    alias = repo_meta["alias"]
    name = repo_meta["name"]
    # make sure the rest is OK
    if "tag" not in repo_meta:
        fail("repository tag missing in repo list")
    tag = repo_meta["tag"]
    if "repository" not in repo_meta:
        fail("repository url missing in repo list")
    url = repo_meta["repository"]
    description = ""
    if "description" in repo_meta:
        description = repo_meta["description"]
    print("downloading source code")
    print("repository info:")
    print(" - repository:  %s" % url)
    print(" - description: %s" % description)
    print(" - tag:         %s" % tag)
    repo = git.Repo.clone_from(url=url, to_path=os.path.join(base_path, alias))
    repo.git.checkout(tag)
    tags = repo.tag("refs/tags/" + tag)
    sha = tags.commit.hexsha
    print(" - tag sha:     %s" % sha)

    # avoid pulling dependencies from github,
    # add replace to go mod files to make sure it builds locally
    update_dep_ref(name, os.path.join(base_path, alias))
    return sha

# Run distclean on the source code path
def clean_release(local_repo_path):
    print("ensuring local source repo is clean")
    path = os.getcwd()
    os.chdir(local_repo_path)
    retcode = subprocess.call(['make', 'distclean'])
    if retcode:
        fail("failed to clean staging repo")
    os.chdir(path)

# Unpack tarball into tmp dir 
def unpack_staging_tarball(staging_dir, dest_dir, release_name):
    path = os.getcwd()
    os.chdir(staging_dir)
    retcode = subprocess.call(['rm', '-rf', dest_dir])
    if retcode:
        fail("failed to clean dest dir")
    retcode = subprocess.call(['mkdir', dest_dir])
    if retcode:
        fail("failed to create dest dir")
    os.chdir(dest_dir)
    retcode = subprocess.call(['tar', 'xf', os.path.join(staging_dir, "%s-staging.tar.gz" % release_name)])
    if retcode:
        fail("failed to unpack tarball")
    os.chdir(path)

# Generate binaries for yunikorn-web and compute checksums
def build_web_and_generate_hashes(staging_dir, release_name, arch):
    print("generating reproducible build artifacts for yunikorn-web (%s)" % arch)
    path = os.getcwd()
    tmp_dir = os.path.join(staging_dir, "tmp")
    release_dir = os.path.join(tmp_dir, release_name)
    unpack_staging_tarball(staging_dir, tmp_dir, release_name)
    web_dir = os.path.join(release_dir, "web")
    os.chdir(web_dir)
    retcode = subprocess.call(['make', 'REPRODUCIBLE_BUILDS=1', 'HOST_ARCH=' + arch, 'build_server_prod'])
    if retcode:
        fail("failed to build yunikorn-web (%s)" % arch)
    hash = get_checksum("build/prod/yunikorn-web", "yunikorn-web")
    os.chdir(staging_dir)
    retcode = subprocess.call(['rm', '-rf', 'tmp'])
    if retcode:
        fail("failed to clean temp dir")
    os.chdir(path)
    return hash

# Generate binaries for yunikorn-k8shim and compute checksums
def build_shim_and_generate_hashes(staging_dir, release_name, arch):
    print("generating reproducible build artifacts for yunikorn-k8shim (%s)" % arch)
    path = os.getcwd()
    tmp_dir = os.path.join(staging_dir, "tmp")
    release_dir = os.path.join(tmp_dir, release_name)
    unpack_staging_tarball(staging_dir, tmp_dir, release_name)
    shim_dir = os.path.join(release_dir, "k8shim")
    os.chdir(shim_dir)
    retcode = subprocess.call(['make', 'REPRODUCIBLE_BUILDS=1', 'HOST_ARCH=' + arch, 'scheduler', 'plugin', 'admission'])
    if retcode:
        fail("failed to build yunikorn-k8shim (%s)" % arch)
    adm_hash = get_checksum("build/bin/yunikorn-admission-controller", "yunikorn-admission-controller")
    scheduler_hash = get_checksum("build/bin/yunikorn-scheduler", "yunikorn-scheduler")
    plugin_hash = get_checksum("build/bin/yunikorn-scheduler-plugin", "yunikorn-scheduler-plugin")
    hash = "\n".join([adm_hash, scheduler_hash, plugin_hash])
    os.chdir(staging_dir)
    retcode = subprocess.call(['rm', '-rf', 'tmp'])
    if retcode:
        fail("failed to clean temp dir")
    os.chdir(path)
    return hash

# K8shim depends on yunikorn-core and scheduler-interface
def update_dep_ref_k8shim(local_repo_path):
    print("updating dependency for k8shim")
    mod_file = os.path.join(local_repo_path, "go.mod")
    if not os.path.isfile(mod_file):
        fail("k8shim go.mod does not exist")
    path = os.getcwd()
    os.chdir(local_repo_path)
    command = ['go', 'mod', 'edit']
    command.extend(['-replace', 'github.com/apache/yunikorn-core=../core'])
    command.extend(['-replace', 'github.com/apache/yunikorn-scheduler-interface=../scheduler-interface'])
    retcode = subprocess.call(command)
    if retcode:
        fail("failed to update k8shim go.mod references")
    os.chdir(path)

# core depends on scheduler-interface
def update_dep_ref_core(local_repo_path):
    print("updating dependency for core")
    mod_file = os.path.join(local_repo_path, "go.mod")
    if not os.path.isfile(mod_file):
        fail("core go.mod does not exist")
    path = os.getcwd()
    os.chdir(local_repo_path)
    command = ['go', 'mod', 'edit']
    command.extend(['-replace', 'github.com/apache/yunikorn-scheduler-interface=../scheduler-interface'])
    retcode = subprocess.call(command)
    if retcode:
        fail("failed to update core go.mod references")
    os.chdir(path)


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
    if not os.path.isfile(make_file):
        fail("k8shim repo Makefile does not exist")
    replace(make_file, "(CORE_SHA=)(.*)", "\\g<1>" + sha["yunikorn-core"])
    replace(make_file, "(SI_SHA=)(.*)", "\\g<1>" + sha["yunikorn-scheduler-interface"])
    replace(make_file, "(SHIM_SHA=)(.*)", "\\g<1>" + sha[repo_name])


# web only uses its own revision
def update_sha_web(repo_name, local_repo_path, sha):
    print("updating sha for web")
    make_file = os.path.join(local_repo_path, "Makefile")
    if not os.path.isfile(make_file):
        fail("web repo Makefile does not exist")
    replace(make_file, "(WEB_SHA=)(.*)", "\\g<1>" + sha[repo_name])


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


# update required Golang version in the README.md
def update_required_go_version(base_path, local_repo_path):
    print("updating required go version")
    go_version_file = os.path.join(local_repo_path, ".go_version")
    if not os.path.isfile(go_version_file):
        fail("k8shim repo .go_version does not exist")
    with open(go_version_file) as f:
        go_version = f.readline().strip()
    if not go_version:
        fail("k8shim repo .go_version is empty")
    print(f" - go version:  {go_version}")
    replace(os.path.join(base_path, "README.md"), 'Go 1.16', 'Go ' + go_version)

# update reproducible build information in README
def update_reproducible_build_info(base_path, go_version, hashes_amd64, hashes_arm64):
    print("recording go compiler used for reproducible builds")
    replace(os.path.join(base_path, "README.md"), '@GO_VERSION@', go_version)
    print("recording build artifact hashes (amd64)")
    replace(os.path.join(base_path, "README.md"), '@AMD64_BINARIES@', hashes_amd64)
    print("recording build artifact hashes (arm64)")
    replace(os.path.join(base_path, "README.md"), '@ARM64_BINARIES@', hashes_arm64)

# update required Node.js and angular versions in the README.md
def update_required_node_and_angular_versions(base_path, local_repo_path):
    print("updating required Node.js version")
    nvmrc_file = os.path.join(local_repo_path, ".nvmrc")
    if not os.path.isfile(nvmrc_file):
        fail("web repo .nvmrc does not exist")
    with open(nvmrc_file) as f:
        node_version = f.readline().strip()
    if not node_version:
        fail("web repo .nvmrc is empty")
    print(f" - node version:  {node_version}")
    replace(os.path.join(base_path, "README.md"), 'Node.js 16.14.2', 'Node ' + node_version)

    print("updating required Angular version")
    package_json_file = os.path.join(local_repo_path, "package.json")
    if not os.path.isfile(package_json_file):
        fail("web repo package.json does not exist")
    with open(package_json_file) as f:
        try:
            data = json.load(f)
        except json.JSONDecodeError:
            fail("load web package.json: unexpected json decode failure")
    angular_version_match = re.search("\d+\.\d+\.\d+", data.get("dependencies", {}).get("@angular/core", ""))
    if not angular_version_match:
        fail("web repo package.json: unexpected @angular/core version")
    angular_version = angular_version_match.group()
    print(f" - angular version:  {angular_version}")
    replace(os.path.join(base_path, "README.md"), 'Angular CLI 13.3.0', 'Angular CLI ' + angular_version)


# update required versions in the README.md
def update_required_versions(release_base, repo_list):
    switcher = {
        "yunikorn-k8shim": update_required_go_version,
        "yunikorn-web": update_required_node_and_angular_versions,
    }
    for repo_meta in repo_list:
        repo_name = repo_meta["name"]
        if switcher.get(repo_name) is not None:
            switcher.get(repo_name)(release_base, os.path.join(release_base, repo_meta["alias"]))


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

# Generate a checksum for a file
def get_checksum(file_path, file_name):
    print("generating sha512 checksum for %s" % file_name)
    h = hashlib.sha512()
    # read the file and generate the sha
    with open(file_path, 'rb') as file:
        while True:
            data = file.read(65536)
            if not data:
                break
            h.update(data)
    sha = h.hexdigest()
    return "%s  %s" % (sha, file_name)

# Sign the source archive if an email is provided
def call_gpg(tarball_file, email_address):
    cmd = shutil.which("gpg")
    if not cmd:
        print("gpg not found on the path, not signing package")
        return
    print("Signing source code file using email: %s" % email_address)
    command = [cmd, '--armor', '--detach-sig']
    command.extend(['--local-user', email_address])
    command.extend(['--output', tarball_file + ".asc", tarball_file])
    retcode = subprocess.call(command)
    if retcode:
        fail("failed to create gpg signature")

# Determine the specific go compiler in use
def get_go_version():
    command = ['go', 'env', 'GOVERSION']
    result = subprocess.run(command, capture_output=True, text=True)
    if result.returncode:
        fail("failed to get go version")
    output = re.sub(r'^go', r'', result.stdout.strip())
    return output

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
    retcode = subprocess.call(command)
    if retcode:
        fail("helm chart creation failed")
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


# Merge the added lines from the license files
def merge_licenses(base_dir, repo_list):
    start = 202  # Apache License is 202 lines
    lic = os.path.join(base_dir, "LICENSE")
    if not os.path.isfile(lic):
        fail("license does not exist at top level staging directory")
    with open(lic, 'a') as lp:
        for repo_meta in repo_list:
            alias = repo_meta["alias"]
            lic_repo = os.path.join(os.path.join(base_dir, alias), "LICENSE")
            if not os.path.isfile(lic_repo):
                fail("license does not exist in '%s' repository" % alias)
            with open(lic_repo, 'r') as fp:
                lines = fp.readlines()
                if len(lines) <= start:
                    continue
                print("copying license details from: %s\n" % alias)
                i = start
                while i < len(lines):
                    lp.write(lines[i])
                    i += 1


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
