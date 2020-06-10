YuniKorn Release Procedure
----

This project provides the instructions and tools needed to generate Apache YuniKorn (Incubating) release artifacts. This obeys ASF [release policy](http://www.apache.org/legal/release-policy.html), and [Podling Release Policy](https://incubator.apache.org/policy/incubation.html#releases).

* [Create a Release](#Create-a-Release)
* [Sign a Release](#Sign-a-release)
* [Generate a Key](#Generate-a-Key)
* [Upload the Key to a Public Key Server](#Upload-the-Key-to-a-Public-Key-Server)
* [Create Signature](#Create-Signature)
* [Verify a Signature](#Verify-a-Signature)
* [Upload Release Artifacts](#Upload-Release-Tarball)
* [Start Voting Thread](#Start-Voting-Thread)
* [Publish the Release](#Publish-the-Release)
* [Release Helm Charts](#Release-Helm-Charts)

# Create a Release

1. Create a release branch for the target release in all git repos, such as `branch-0.8`
2. Stabilize the release branch
3. Create a tag and prepare to generate release candidate 1, e.g `release-0.8-rc1`
3. Configure `tools/release-configs.json`
4. Run script `tools/build-release.json` to generate source code tarball

# Sign a release

If you haven't signed any releases before, please read the doc:
- [generate signing key](https://infra.apache.org/openpgp.html#generate-key)

# Generate a Key
Generate a new PGP key (skip this step if you already have a key):

```shell script
gpg --gen-key

# Real name: your full name
# Email address: your apache email address
```

# Upload the Key to a Public Key Server

```shell script
gpg --export --armor
```

then upload to https://pgp.mit.edu/.

You will also need to upload public key files to apache server, https://people.apache.org/keys/.

```shell script
gpg --fingerprint
```

then copy the fingerprint to https://id.apache.org/.

# Create Signature

```shell script
// gpg signature
gpg --armor --output apache-yunikorn-0.8.0-incubating-src.tar.gz.asc --detach-sig apache-yunikorn-0.8.0-incubating-src.tar.gz

# checksum
shasum -a 512 apache-yunikorn-0.8.0-incubating-src.tar.gz > apache-yunikorn-0.8.0-incubating-src.tar.gz.sha512
```

this will create the signature in file `apache-yunikorn-incubating-0.8.0-rc1.asc`

# Verify a Signature

```shell script
gpg --verify apache-yunikorn-incubating-0.8.0-rc1.asc apache-yunikorn-incubating-0.8.0-rc1.tar.gz
```

# Upload Release Artifacts

Make sure your public SSH key is uploaded to https://id.apache.org/. The keys need to be appended to
https://dist.apache.org/repos/dist/dev/incubator/yunikorn/KEYS. Other artifacts need to be uploaded to
https://dist.apache.org/repos/dist/dev/incubator/yunikorn/. Note, you will need to install subversion to
access this repo (use your apache ID). You can use a SVN client, e.g svnX, for convenience.

# Start Voting Thread

According to [podling release doc](https://incubator.apache.org/policy/incubation.html#releases) and [release approval doc](http://www.apache.org/legal/release-policy.html#release-approval). Steps are:
- start a voting thread on `dev@yunikorn.apache.org`. (72 hours)
- send a summary of that vote to the Incubator’s general list and request IPMC to vote. (72 hours)
Both voting need to acquire at least three +1 votes are required and more +1 votes than -1 votes.

# Publish the Release

Once the voting is passed, move the release artifacts to https://dist.apache.org/repos/dist/release/incubator/yunikorn/. Once moved to this space, the content will be automatically sync'd to https://downloads.apache.org/incubator/yunikorn/ which can be used as the final location for release files. Read more for
[location of files on main server](https://infra.apache.org/mirrors#location).

Publish an announcement blog to https://blogs.apache.org/yunikorn/, update the web-site with corresponding
release notes, download links.

# Release Helm Charts

After the voting passed and the RC is accepted, release the helm chart
- Create a release branch for the target release in this release repo
- Package the charts: 
```shell script
helm package --sign --key ${your_key_name} --keyring ${path/to/keyring.secret} helm-charts/yunikorn --destination .
```
Fore more information please check [Helm documentation](https://helm.sh/docs/topics/provenance/)
- upload the packaged chart to the release in this repository
- update the [index.yaml](https://github.com/apache/incubator-yunikorn-release/blob/gh-pages/index.yaml) file in the gh-pages branch with the new release

