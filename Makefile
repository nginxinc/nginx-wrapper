#
#  Copyright 2020 F5 Networks
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#

PACKAGE_ROOT  := github.com/nginxinc
PACKAGE       := $(PACKAGE_ROOT)/nginx-wrapper
DATE          ?= $(shell date -u +%FT%T%z)
VERSION       ?= $(shell cat $(CURDIR)/.version 2> /dev/null || echo 0.0.0)
GITHASH       ?= $(shell git rev-parse HEAD)

GOPATH        ?= $(CURDIR)/.gopath
BIN           := $(GOPATH)/bin
SED           ?= $(shell which gsed 2> /dev/null || which sed 2> /dev/null)
ARCH          := $(shell uname -m | $(SED) -e 's/x86_64/amd64/g' -e 's/i686/i386/g')
PLATFORM      := $(shell uname | tr '[:upper:]' '[:lower:]')
PLUGIN_ROOTS  ?= $(shell find plugins -maxdepth 1 -mindepth 1 -type d | sort)
GOPATH        ?= .gopath
SHELL         := bash

GO      = go
GODOC   = godoc
GOFMT   = gofmt
TIMEOUT = 45
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")

export GOPATH:=$(GOPATH)
$(info $(M) using GOPATH=$(GOPATH))

# Look up variables that allow us to match a target to a directory
SRC_DIR_nginx-wrapper     := app
SRC_DIR_nginx-wrapper-lib := lib

.PHONY: help
help:
	@grep --no-filename -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-24s\033[0m %s\n", $$1, $$2}' | sort

include build/docker.mk
include build/tools.mk
include build/compile.mk
include build/test.mk
include build/release.mk

target:
	$Q mkdir -p $(CURDIR)/$@

.PHONY: clean
clean: ; $(info $(M) cleaning...)	@ ## Cleanup everything
	@rm -rf target || true
	@rm -rf test