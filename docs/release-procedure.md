YuniKorn Release Procedure
----

This project provides the instructions and tools needed to generate Apache YuniKorn (Incubating) release artefacts. This obeys ASF [release policy](http://www.apache.org/legal/release-policy.html), and [Podling Release Policy](https://incubator.apache.org/policy/incubation.html#releases).

* [Create a Release](#Create-a-Release)
* [Sign a Release](#Sign-a-release)
* [Create Signature and Checksum](#Create-Signature-and-Checksum)
* [Verify a Signature](#Verify-a-Signature)
* [Upload Release Artefacts](#Upload-Release-Artefacts)
* [Start Voting Thread](#Start-Voting-Thread)
* [Publish the Release](#Publish-the-Release)
* [Release Helm Charts](#Release-Helm-Charts)

# Create a Release
1. Create a release branch for the target release in all git repos, such as `branch-0.8`
2. Stabilize the release branch
3. Create a tag and prepare to generate release candidate, e.g `v0.8.0`
4. Configure `tools/release-configs.json`
5. Run script `tools/build-release.py` to generate source code tarball

## Tag and update release for version
A release needs to be tagged in git before proceeding. This triggers required updates in the go mod files for the branch.
As an example check [YUNIKORN-358](https://issues.apache.org/jira/browse/YUNIKORN-358).

The tagging is multi step process, all actions are done on the branch that will be released, like `branch-0.8`:
1. Tag the web and scheduler interface with the release tag.
2. Update the `go.mod` file in the core using `go get github.com/apache/incubator-yunikorn-scheduler-interface`  
Add the tag and commit the changes.
3. Update the `go.mod` file in the shim using `go get github.com/apache/incubator-yunikorn-scheduler-interface` and  
`go get github.com/apache/incubator-yunikorn-core`. Add the tag and commit the changes.
## Sign a release
If you haven't signed any releases before, read the documentation to [generate signing key](https://infra.apache.org/openpgp.html#generate-key)
Follow the steps below to add the key.

### Generate a Key
Generate a new PGP key (skip this step if you already have a key):
```shell script
gpg --gen-key
```
And fill out the requested information using your full name and Apache email address.

Upload the exported key to a public key server like `https://pgp.mit.edu/`.
```shell script
gpg --export --armor <email address>
```

Upload the fingerprint to apache server: `https://id.apache.org/`.
```shell script
gpg --fingerprint <email address>
```

### Add the signature to the project KEYS file
Only needed if this is the first release signed with the specific key.
More detail can be found in the document: [Signing a Release](https://infra.apache.org/release-signing.html#keys-policy)
```shell script
(gpg --list-sigs <email address> && gpg --armor --export <email address>) >> MY_KEY
```
Add the content of the generated file to the existing KEYS list at `https://dist.apache.org/repos/dist/release/incubator/yunikorn/KEYS`
Never remove a key from this list!

NOTE: you will need to install subversion to access this repo (use your apache ID). You can use any SVN client, e.g svnX, for convenience.

## Create Signature and Checksum
```shell script
# gpg signature
gpg --local-user <email address> --armor --output apache-yunikorn-0.8.0-incubating-src.tar.gz.asc --detach-sig apache-yunikorn-0.8.0-incubating-src.tar.gz

# checksum
shasum -a 512 apache-yunikorn-0.8.0-incubating-src.tar.gz > apache-yunikorn-0.8.0-incubating-src.tar.gz.sha512
```

This will create the signature in the file: `apache-yunikorn-0.8.0-incubating-src.tar.gz.asc`
And the checksum in the file: `apache-yunikorn-0.8.0-incubating-src.tar.gz.sha512`

## Verify a Signature
```shell script
gpg --verify apache-yunikorn-0.8.0-incubating-src.tar.gz.asc apache-yunikorn-0.8.0-incubating-src.tar.gz
```

## Upload Release Artefacts
The release artefacts consist of three parts:
* source tarball
* signature file
* checksum file
The three artefacts need to be uploaded to: `https://dist.apache.org/repos/dist/dev/incubator/yunikorn/` 

Create a release directory based on the version, i.e. `0.8.0`, add the three files to directory.
Commit the changes.

If you have not done so already make sure to [add your signature](#add-the-signature-to-the-project-keys-file) to the KEYS file.

NOTE: you will need to install subversion to access this repo (use your apache ID). You can use any SVN client, e.g svnX, for convenience.


## Start Voting Thread

According to [podling release doc](https://incubator.apache.org/policy/incubation.html#releases) and [release approval doc](http://www.apache.org/legal/release-policy.html#release-approval). Steps are:
- start a voting thread on `dev@yunikorn.apache.org`. (72 hours)
- send a summary of that vote to the Incubator’s general list and request IPMC to vote. (72 hours)
Both voting need to acquire at least three +1 votes are required and more +1 votes than -1 votes.

## Publish the Release

Once the voting is passed, move the release artefacts to https://dist.apache.org/repos/dist/release/incubator/yunikorn/. 
Once moved to this space, the content will be automatically synced to https://downloads.apache.org/incubator/yunikorn/ which can be used as the final location for release files.
Read more for [location of files on main server](https://infra.apache.org/mirrors#location).

## Release Docker images

Push the latest docker images to docker hub.

## Release Helm Charts

- Create a release branch for the target release in this release repo
- Package the charts: 
```shell script
helm package --sign --key ${your_key_name} --keyring ${path/to/keyring.secret} helm-charts/yunikorn --destination .
```
Fore more information please check [Helm documentation](https://helm.sh/docs/topics/provenance/)
- upload the packaged chart to the release in this repository
- update the [index.yaml](https://github.com/apache/incubator-yunikorn-release/blob/gh-pages/index.yaml) file in the gh-pages branch with the new release

## Update the website

- Create a new version of the [YuniKorn website](https://github.com/apache/incubator-yunikorn-site/tree/master).
- Update the [download page](http://yunikorn.apache.org/community/download) of the website.
- Publish release announcement in the [apache Blog](https://blogs.apache.org/yunikorn/) and the the [YuniKorn website](http://yunikorn.apache.org).
Update the website's release notes and download links.

## Verify the release

After the whole procedure verify the documents and the released artifacts.
