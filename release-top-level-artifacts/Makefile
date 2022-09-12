#
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
#

# standard directories
CORE_DIR := core
SHIM_DIR := k8shim
WEB_DIR := web

# default repository and version
ifeq ($(REGISTRY),)
REGISTRY := apache
endif
export REGISTRY

ifeq ($(VERSION),)
VERSION := latest
endif
export VERSION

# Make sure we are in the same directory as the Makefile
BASE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

all:
	$(MAKE) -C $(dir $(BASE_DIR)) image

.PHONY: image
image:
	$(MAKE) -C $(SHIM_DIR) image
	$(MAKE) -C $(WEB_DIR) image

.PHONY: test
test:
	$(MAKE) -C $(CORE_DIR) test
	$(MAKE) -C $(SHIM_DIR) test
	$(MAKE) -C $(WEB_DIR) test

.PHONY: arch
arch:
	@eval "$(MAKE) -C $(SHIM_DIR) arch"
