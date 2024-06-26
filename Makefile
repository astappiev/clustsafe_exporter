# Copyright 2015 The Prometheus Authors
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Override the default common all.
.PHONY: all
all: precheck style unused build test

DOCKER_ARCHS      ?= amd64 arm64
DOCKER_IMAGE_NAME ?= clustsafe-exporter
DOCKER_REPO       ?= astappiev

include Makefile.common

crossbuild: promu
	@echo ">> building crossbuild release"
	@$(PROMU) crossbuild

crossbuild-tarballs: promu
	@echo ">> building crossbuild release tarballs"
	@$(PROMU) crossbuild tarballs

crossbuild-checksum: promu
	@echo ">> calculating checksums for released packages"
	@$(PROMU) checksum .tarballs

crossbuild-release: promu crossbuild crossbuild-tarballs crossbuild-checksum
