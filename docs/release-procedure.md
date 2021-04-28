<!--
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to you under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

YuniKorn Release Procedure
----

This project provides the instructions and tools needed to generate Apache YuniKorn (Incubating) release artefacts. This obeys ASF [release policy](http://www.apache.org/legal/release-policy.html), and [Podling Release Policy](https://incubator.apache.org/policy/incubation.html#releases).

* [Create a Release](#Create-a-Release)
    * [Tag and update release for version](#Tag-and-update-release-for-version)
    * [Update the CHANGELOG](#Update-the-CHANGELOG)
    * [Run the release tool](#Run-the-release-tool)
        * [Create Signature](#Create-Signature)
        * [Create Checksum](#Create-Checksum)
    * [Upload Release Candidate Artefacts](#Upload-Release-Candidate-Artefacts)
    * [Start Voting Thread](#Start-Voting-Thread)
    * [Publish the Release](#Publish-the-Release)
        * [Release Docker images](#Release-Docker-images)
        * [Release Helm Charts](#Release-Helm-Charts)
        * [Update the website](#Update-the-website)
        * [Create the GIT releases](#Create-the-GIT-releases)
    * [Verify the release](#Verify-the-release)
* [Signing your first release](#Signing-your-first-release)
    * [Generate a Key](#Generate-a-Key)
    * [Add the signature to the project KEYS file](#Add-the-signature-to-the-project-KEYS-file)

# Create a Release
Simplified release procedure: 
1. Create a release branch for the target release in all git repos, such as `branch-0.8`
2. Stabilize the release by fixing test failures and bugs only
3. Tag update release for a new version to prepare a release candidate, e.g `v0.8.0`
4. Update the CHANGELOG
5. Configure [release-configs.json](../tools/release-configs.json)
6. Run script [build-release.py](../tools/build-release.py) to generate source code tarball, checksum and signature.
7. Voting and releasing the candidate

## Tag and update release for version
Branching and tagging can, and in most cases will, require changes in the go mod files.
Branching is part of the release preparation and often has happened some time before the release process starts. 
A release needs to be tagged in git before starting the release process.
As an example check [YUNIKORN-358](https://issues.apache.org/jira/browse/YUNIKORN-358).
Release candidates and final release use the same tag which gets moved if a new release candidate is generated. 

The tagging is multi step process, all actions are done on the branch that will be released, like `branch-0.8`:
1. Tag the web and scheduler interface with the release tag.
2. Update the `go.mod` file in the core using `go get github.com/apache/incubator-yunikorn-scheduler-interface`  
Add the tag and commit the changes.
3. Update the `go.mod` file in the shim using `go get github.com/apache/incubator-yunikorn-scheduler-interface` and  
`go get github.com/apache/incubator-yunikorn-core`. Add the tag and commit the changes.
4. Create a new branch in the yunikorn-release repo, set the correct chart version in [Chart.yaml](../helm-charts/yunikorn/Chart.yaml), and then create the tag.

## Update the CHANGELOG
In the release artifacts a [CHANGELOG](../release-top-level-artifacts/CHANGELOG) is added for each release.
The CHANGELOG should contain the list of jiras fixed in the release.
Follow these steps to generate the list:
- Go to the [releases page in jira](https://issues.apache.org/jira/projects/YUNIKORN?selectedItem=com.atlassian.jira.jira-projects-plugin%3Arelease-page&status=released-unreleased)
- Click on the version that is about to be released, i.e. `0.8`
- Click on the `Release Notes` link on the top of the page
- Click the button `Configure Release Notes`
- Select the style `Text` and click `create`
- Scroll to the bottom of the page and copy the content of the text area and update the [CHANGELOG](../release-top-level-artifacts/CHANGELOG) file.

## Run the release tool
A tool has been written to handle most of the release tasks.
The tool requires a simple [json](../tools/release-configs.json) input file to be updated before running.
This configuration points to the current release tag. Only update the tag for each repository.

The tool has one requirement outside of standard Python 3: [GitPython](https://gitpython.readthedocs.io/en/stable/intro.html)
Make sure you have installed it by running `pip install gitpython`.

Run the tool:
```shell script
python3 build-release.py
```
If you want to automatically sign the release using your GPG key run the tool using:
```shell script
python3 build-release.py --sign <email-address>
```

### Create Signature
If you have GPG with a _pinentry_ program setup you can automatically sign the release using the release tool.
On MacOSX this will be setup automatically if you use the keychain for the keys.
For more details check the [GnuPG tools wiki](https://wiki.archlinux.org/index.php/GnuPG) and specifically the [pinentry](https://wiki.archlinux.org/index.php/GnuPG#pinentry) chapter.  

Run the release tool using the option `--sign <email-address>` to auto sign the release.
 
Manually creating the signature for the file generated by the tool:
```shell script
gpg --local-user <email-address> --armor --output apache-yunikorn-0.8.0-incubating-src.tar.gz.asc --detach-sig apache-yunikorn-0.8.0-incubating-src.tar.gz
```
This will create the signature in the file: `apache-yunikorn-0.8.0-incubating-src.tar.gz.asc`
Verify that the signature is correct using:
```shell script
gpg --verify apache-yunikorn-0.8.0-incubating-src.tar.gz.asc apache-yunikorn-0.8.0-incubating-src.tar.gz
```

### Create Checksum
This step is included in the release after generation of the source tar ball, if the release tool is used this step can be skipped. 
```shell script
shasum -a 512 apache-yunikorn-0.8.0-incubating-src.tar.gz > apache-yunikorn-0.8.0-incubating-src.tar.gz.sha512
```
This will create the checksum in the file: `apache-yunikorn-0.8.0-incubating-src.tar.gz.sha512`
Verify that the checksum is correct using:
```shell script
shasum -a 512 -c apache-yunikorn-0.8.0-incubating-src.tar.gz.sha512 
```

## Upload Release Candidate Artefacts
The release artefacts consist of three parts:
- source tarball
- signature file
- checksum file
The three artefacts need to be uploaded to: `https://dist.apache.org/repos/dist/dev/incubator/yunikorn/` 

Create a release directory based on the version, i.e. `0.8.0`, add the three files to directory.
Commit the changes.

If you have not done so already make sure to [add your signature](#add-the-signature-to-the-project-keys-file) to the KEYS file.
Do not remove any keys from the file they are kept here to enable older releases to be verified.

NOTE: you will need to install subversion to access this repo (use your apache ID). You can use any SVN client, e.g svnX, for convenience.

## Start Voting Thread
According to [podling release doc](https://incubator.apache.org/policy/incubation.html#releases) and [release approval doc](http://www.apache.org/legal/release-policy.html#release-approval). Steps are:
- start a voting thread on `dev@yunikorn.apache.org`. (72 hours)
- send a summary of that vote to the Incubator’s general list and request IPMC to vote. (72 hours)
Both voting need to acquire at least three +1 votes are required and more +1 votes than -1 votes.

## Publish the Release
Once the voting is passed, move the release artefacts from the staging area to the release location `https://dist.apache.org/repos/dist/release/incubator/yunikorn/`. 
Once moved to this space, the content will be automatically synced to `https://downloads.apache.org/incubator/yunikorn/` which must be used as the final location for release files.
Read more for [location of files on main server](https://infra.apache.org/mirrors#location).

This will temporarily provide us with two releases in the release area.
This is needed to allow the start the mirror sync process and allow for the download page to be updated.
Cleanup of the older release is handled after the website has been updated in the [cleanup](#Cleanup). 

### Release Docker images
The standard build process should be used to build the image.
Run a `make image` in the `web`, and `k8shim` repositories to generate the three images required (web, scheduler and admission-controller):
```shell script
VERSION=0.8.0; make image
```

Make can also be used to build and push the image if you have access to the Apache docker hub YuniKorn container.
Push the latest docker images to the apache docker hub using the release as tag.
Make sure the docker image is built on the specific SHA.
```shell script
VERSION=0.8.0; DOCKER_USERNAME=<name>; DOCKER_PASSWORD=<password>; make push 
```
Publish an announcement email to the `dev@yunikorn.apache.org` email list. 

### Release Helm Charts
This step is part of the release tool if the release tool is used the packaging can be skipped.

If the release tool is **not** used the `Chart.yaml` and the `values.yaml` must be updated manually.
The other option is to run the helm script against the generated source directory as the tool does: 
```shell script
helm package --sign --key ${your_key_name} --keyring ${path/to/keyring.secret} staging/<release-dir>/helm-charts/yunikorn --destination staging/ 
```
Signing the helm package requires a legacy PGP keyring. The PGP v2 keyring must be converted to the legacy format.
For more information please check [Helm documentation](https://helm.sh/docs/topics/provenance/).
Helm charts should be signed on release.
Contrary to the source code tar ball signing, signing the helm charts requires manual entry of the key password. 

The helm package will generate two files:
- helm package: example `yunikorn-0.8.0.tgz`
- provenance or signature file: example `yunikorn-0.8.0.tgz.prov`
Both files should be attached to the [release in GIT](#Create-the GIT-releases) for the release repository.

Last step is to update the [index.yaml](https://github.com/apache/incubator-yunikorn-release/blob/gh-pages/index.yaml) file in the `gh-pages` branch with the new release.
The `digest` mentioned in the index.yaml file is the digest that gets printed by the tool (unsigned package) or stored in the provenance file.
It can be generated manually using:
```shell script
shasum -a 256 yunikorn-0.8.0.tgz
```

Note: do not use the `helm repo index` command to update the `index.yaml` file. The command does not handle the enhanced information stored in the `index.yaml` file nicely.
Update the file manually.

### Update the website
- Create a new documentation version on YuniKorn website based on the latest content in [docs](https://github.com/apache/incubator-yunikorn-site/tree/master/docs) directory. Refer to [this](https://github.com/apache/incubator-yunikorn-site/tree/master#release-a-new-version) guide to create the new documentation. 
- Create the release announcement to be referenced from download page on the website. The release announcement is a markdown file based on the version: `0.8.0.md`. The file is stored as part of the [static pages](https://github.com/apache/incubator-yunikorn-site/tree/master/src/pages/release-announce) on the website. 
- Update the [download page](https://github.com/apache/incubator-yunikorn-site/tree/master/src/pages/community/download.md) of the website.

The release announcement are linked to the release details on the download page.

Links for the releases have to follow these rules:
* The first download link on the page **must** use the mirror resolution link for the source tar ball only.
* The signature and checksum links **must** point to the release location.
* The non-current releases **must** use the archive links: `https://archive.apache.org/dist/incubator/yunikorn/` for the tar ball, the signature and the checksum.

A limited set of three (3) or four (4) releases should be maintained in the table for direct access.
Older releases not mentioned in the table can still be accessed via the archive link on the bottom of the page and do not need to be referenced.

### Cleanup
NOTE: this step should be performed after the website updates have been made as the download links change.

There should only be one release, the latest, in the release area.
Any release that has been in the release area will be automatically copied to the archive. 
Older releases should be downloaded from the archive directly, not from the release area.

The releases need to clean up in two locations:
* Remove the newly released version from the _dev_ area by removing the old release candidate directory.
  For the location see [release candidate location](#Upload-Release-Candidate-Artefacts)
* Remove the non-current release from the _release_ area by removing the old release directory.
  For the location see [release location](#Publish-the-Release)

### Create the GIT releases
In the GIT repositories finish the release process by creating a release based on the git tag that was added.
Repeat these steps for all five repositories (core, k8shim, web, scheduler-interface and release):
- Go to the `tags` page
- click the `...` at the right-hand side of the page that you want to release, select `Create Release` from the drop down
- update the name and note
- add the packaged helm chart files (incubator-yunikorn-release repository only)
- click `Publish Release` to finish the steps

## Verify the release
After the whole procedure verify the documentation on the website and that the released artifacts can be downloaded.
Mirror links might take up to 24 hours to be updated.

# Signing your first release
If you haven't signed any releases before, read the documentation to [generate signing key](https://infra.apache.org/openpgp.html#generate-key)
Follow the steps below to add the key you can use to sign. 

## Generate a Key
Generate a new PGP key (skip this step if you already have an Apache linked key):
```shell script
gpg --gen-key
```
Fill out the requested information using your full name and Apache email address.

Upload the exported key to a public key server like `https://pgp.mit.edu/`.
```shell script
gpg --export --armor <email-address>
```

Upload the fingerprint to apache server: `https://id.apache.org/`.
```shell script
gpg --fingerprint <email-address>
```

## Add the signature to the project KEYS file
Only needed if this is the first release signed with the specific key.
More detail can be found in the document: [Signing a Release](https://infra.apache.org/release-signing.html#keys-policy)
```shell script
(gpg --list-sigs <email-address> && gpg --armor --export <email-address>) >> MY_KEY
```
Add the content of the generated file to the existing KEYS list at `https://dist.apache.org/repos/dist/release/incubator/yunikorn/KEYS`
Never remove a key from this list!

NOTE: you will need to install subversion to access this repo (use your apache ID). You can use any SVN client, e.g svnX, for convenience.
