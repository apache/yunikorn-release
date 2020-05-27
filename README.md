<!--
#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
-->

# Apache YuniKorn (Incubating) Release
----
This project provides the instructions and tools needed to generate Apache YuniKorn (Incubating) release artifacts.
Reference:
 - [ASF Release Creation Process](https://infra.apache.org/release-publishing.html)
 - [ASF Release Policy](http://www.apache.org/legal/release-policy.html).

# Release Procedure

- Create a release branch for the target release in all git repos, such as `branch-0.8`
- Stabilize the release branch
- Create a tag and prepare to generate release candidate 1, e.g `release-0.8-rc1`
- Configure `tools/release-configs.json`
- Run script `tools/build-release.json` to generate source code tarball
- Sign the release
- Upload tarball for voting

See full document at this [doc](docs/release-procedure.md)

