#!/usr/bin/env bash

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

# Kernel (OS) Name
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
# show that the check has started
echo "checking license headers:"
# run different finds on mac vs linux
if [ "${OS}" = "darwin" ]; then
  find -E . ! -path "./.git*" -regex ".*(Makefile|\.(go|sh|py|md|conf|yaml|yml|tpl))" -exec grep -L "Licensed to the Apache Software Foundation" {} \; > LICRES
else
  find . ! -path "./.git*" -regex ".*\(Makefile\|\.\(go\|py\|sh\|md\|yaml\|yml\|tpl\)\)" -exec grep -L "Licensed to the Apache Software Foundation" {} \; > LICRES
fi
# any file mentioned in the output is missing the license
if [ -s LICRES ]; then
  echo "following files are missing license header:"
	cat LICRES
	rm -f LICRES
	exit 1
fi
rm -f LICRES
echo "  all OK"
exit 0
