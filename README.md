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

# Apache YuniKorn Release
----
This project provides the instructions and tools needed to generate Apache YuniKorn release artifacts.
Reference:
 - [ASF Release Creation Process](https://infra.apache.org/release-publishing.html)
 - [ASF Release Policy](http://www.apache.org/legal/release-policy.html).

# Release Procedure
A simplified procedure: 
- Create a release branch in all git repos, such as `branch-0.8`
- Stabilize the release
- Create a tag and prepare to generate the release, e.g `v0.8.0`
- Run the release tool to generate source code tarball, checksum and signature
- Upload tarball, signature and checksum as a release candidate
- Start a voting thread for the project 
- Publish the release

The full procedure is documented in the [release procedure](https://yunikorn.apache.org/community/release_procedure).
