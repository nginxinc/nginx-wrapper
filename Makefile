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

PACKAGE_ROOT  = github.com/nginxinc
PACKAGE       = $(PACKAGE_ROOT)/nginx-wrapper
DATE         ?= $(shell date -u +%FT%T%z)
VERSION      ?= $(shell cat $(CURDIR)/.version 2> /dev/null || echo unknown)
GITHASH      ?= $(shell git rev-parse HEAD)

GOPATH        ?= $(CURDIR)/.gopath
BIN           = $(GOPATH)/bin
BASE          = $(GOPATH)/src/$(PACKAGE)
LDFLAGS       = '-X main.githash=$(GITHASH) -X main.buildstamp=$(DATE) -X main.appversion=$(VERSION)'
DEBUGFLAGS    = -gcflags "all=-N -l"
RELEASEFLAGS  = -tags release
FLAGS         = -trimpath
PKGS          = $(or $(PKG),$(shell cd $(BASE) && env GOPATH=$(GOPATH) $(GO) list ./app/... ./lib/... ./plugins/...))
TESTPKGS      = $(shell env GOPATH=$(GOPATH) $(GO) list -f '{{ if or .TestGoFiles .XTestGoFiles }}{{ .ImportPath }}{{ end }}' $(PKGS))
ARCH          = $(shell uname -m | sed -e 's/x86_64/amd64/g' -e 's/i686/i386/g')
PLATFORM      = $(shell uname | tr '[:upper:]' '[:lower:]')
PLUGINS       = $(sort $(dir $(wildcard $(BASE)/plugins/*/)))
DEBUG         = 0
DISTPKGDIR    = 'target/package'

GO      = go
GODOC   = godoc
GOFMT   = gofmt
TIMEOUT = 45
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")

ifeq ($(PLATFORM),darwin)
	SHAREDLIBEXT = 'dylib'
else
	SHAREDLIBEXT = 'so'
endif

ifeq ($(DEBUG), 1)
	FLAGS += $(DEBUGFLAGS)
	OUTPUT_DIR = 'target/debug'
else
	FLAGS += $(RELEASEFLAGS)
	OUTPUT_DIR = 'target/release'
endif

export GOPATH:=$(GOPATH)

$(BASE):
	$(info $(M) setting GOPATH...)
	@mkdir -p $(GOPATH)/src/$(PACKAGE)
	@ln -sf $(CURDIR)/app $(GOPATH)/src/$(PACKAGE)/app
	@ln -sf $(CURDIR)/lib $(GOPATH)/src/$(PACKAGE)/lib
	@ln -sf $(CURDIR)/plugins $(GOPATH)/src/$(PACKAGE)/plugins

# Tools

GOLINT = $(BIN)/golint
$(BIN)/golint:
	$(info $(M) building golint...)
	$Q go get -u golang.org/x/lint/golint
GOLANGCILINT = $(BIN)/golangci-lint
$(BIN)/golangci-lint:
	$(info $(M) building golangci-lint...)
	$Q GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint
GOCOVMERGE = $(BIN)/gocovmerge
$(BIN)/gocovmerge:
	$(info $(M) building gocovmerge...)
	$Q go get github.com/wadey/gocovmerge

GOCOV = $(BIN)/gocov
$(BIN)/gocov:
	$(info $(M) building gocov...)
	$Q go get github.com/axw/gocov/...

GOCOVXML = $(BIN)/gocov-xml
$(BIN)/gocov-xml:
	$(info $(M) building gocov-xml...)
	$Q go get github.com/AlekSi/gocov-xml

GO2XUNIT = $(BIN)/go2xunit
$(BIN)/go2xunit:
	$(info $(M) building go2xunit...)
	$Q go get github.com/tebeka/go2xunit

COMMITSAR = $(BIN)/commitsar
$(BIN)/commitsar:
	$(info $(M) building commitsar...)
	$Q go get github.com/aevea/commitsar

.PHONY: get-tools
get-tools: $(GOLINT) $(GOLANGCILINT) $(GOCOVMERGE) $(GOCOV) $(GOCOVXML) $(GO2XUNIT) $(COMMITSAR) ## Retrieves and builds all the required tools

# Tests

TEST_TARGETS := test-default test-bench test-short test-verbose test-race
.PHONY: $(TEST_TARGETS) test-xml check test test-coverate-tools test-coverage
test-bench:   ARGS=-run=__absolutelynothing__ -bench=. ## Run benchmarks
test-short:   ARGS=-short        ## Run only short tests
test-verbose: ARGS=-v            ## Run tests in verbose mode with coverage reporting
test-race:    ARGS=-race         ## Run tests with race detector
$(TEST_TARGETS): NAME=$(MAKECMDGOALS:test-%=%)
$(TEST_TARGETS): test
check test: all ## Run tests
	$(info $(M) running $(NAME:%=% )tests...) @ ## Run tests

	$Q cd $(BASE)/lib && \
		$(GO) test -timeout $(TIMEOUT)s $(ARGS) ./...
	$Q cd $(BASE)/app && \
		$(GO) test -timeout $(TIMEOUT)s $(ARGS) ./...
	$Q find ./plugins -type f -name go.mod \
		-execdir $(GO) test -timeout $(TIMEOUT)s $(ARGS) ./... \;

test-xml: lib app $(GO2XUNIT) ; $(info $(M) running $(NAME:%=% )tests...) @ ## Run tests with xUnit output
	$Q cd $(BASE)/app && 2>&1 $(GO) test -timeout 20s -v $(TESTPKGS) | tee test/tests.output
	$(GO2XUNIT) -fail -input test/tests.output -output test/tests.xml

COVERAGE_MODE = atomic
COVERAGE_PROFILE = $(COVERAGE_DIR)/profile.out
COVERAGE_XML = $(COVERAGE_DIR)/coverage.xml
COVERAGE_HTML = $(COVERAGE_DIR)/index.html
.PHONY: test-coverage test-coverage-tools
test-coverage-tools: | $(GOCOVMERGE) $(GOCOV) $(GOCOVXML)
test-coverage: COVERAGE_DIR := $(CURDIR)/test/coverage.$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
test-coverage: test-coverage-tools | $(BASE) ; $(info $(M) running coverage tests...) @ ## Run coverage tests
	$Q mkdir -p $(COVERAGE_DIR)/coverage
	$Q cd $(BASE)/app && for pkg in $(TESTPKGS); do \
		$(GO) test \
			-coverpkg=$$($(GO) list -f '{{ join .Deps "\n" }}' $$pkg | \
					grep '^$(PACKAGE)/' | \
					tr '\n' ',')$$pkg \
			-covermode=$(COVERAGE_MODE) \
			-coverprofile="$(COVERAGE_DIR)/coverage/`echo $$pkg | tr "/" "-"`.cover" $$pkg ;\
	 done
	$Q $(GOCOVMERGE) $(COVERAGE_DIR)/coverage/*.cover > $(COVERAGE_PROFILE)
	$Q $(GO) tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)
	$Q $(GOCOV) convert $(COVERAGE_PROFILE) | $(GOCOVXML) > $(COVERAGE_XML)

.PHONY: lint
lint: $(GOLINT) ## Run golint check
	$(info $(M) running golint...) @
	@$(GOLINT) $(CURDIR)/app/... $(CURDIR)/lib/...

	@find $(CURDIR)/plugins -mindepth 1 -maxdepth 1 -type d \
		-exec $(GOLINT) "{}/..." \;
.PHONY: golangci-lint
golangci-lint: deps $(GOLANGCILINT) ## Run golangci-lint check
	$(info $(M) running golangci-lint...) @
	@cd app; \
		$(GOLANGCILINT) --config $(CURDIR)/.golangci.toml --color always --issues-exit-code 0 --path-prefix app/ run
	@cd lib; \
		$(GOLANGCILINT) --config $(CURDIR)/.golangci.toml --color always --issues-exit-code 0 --path-prefix lib/ run
	@cd plugins/example; \
		$(GOLANGCILINT) --config $(CURDIR)/.golangci.toml --color always --issues-exit-code 0 --path-prefix plugins/example/ run

.PHONY: commitsar
commitsar: $(COMMITSAR)  ## Run git commit linter
	$(info $(M) running commitsar...) @
	@ $(COMMITSAR)

.PHONY: all-linters
all-linters: lint golangci-lint commitsar ## Run all linters

.PHONY: fmt
fmt: ## Run source code formatter
	$(info $(M) running gofmt...) ## Run gofmt on all source files
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./... ); do \
		$(GOFMT) -l -w $$d/*.go || ret=$$? ; \
	 done ; exit $$ret

.PHONY: deps
deps: $(BASE) ## Download dependencies
	$(info $(M) downloading dependencies...) @
	$Q cd $(BASE)/lib && $(GO) mod download
	$Q cd $(BASE)/lib && GO111MODULE=off $(GO) get
	$Q cd $(BASE)/app && $(GO) mod download
	$Q cd $(BASE)/app && GO111MODULE=off $(GO) get
	@find $(BASE)/plugins -type f -name go.mod -execdir $(GO) mod download \;
	@GO111MODULE=off find $(BASE)/plugins -type f -name go.mod -execdir pwd \;
$(OUTPUT_DIR)/nginx-wrapper-lib: $(BASE) deps
	$(info $(M) building common libraries...) @
	$Q cd $(BASE)/lib && $(GO) build \
		$(FLAGS) \
		-ldflags $(LDFLAGS) \
		-o $(CURDIR)/$(OUTPUT_DIR)/nginx-wrapper-lib \
        $(GOPATH)/src/$(PACKAGE)/lib/*.go

$(OUTPUT_DIR)/nginx-wrapper: lib $(BASE)
	$(info $(M) building nginx-wrapper...) @
	$Q cd $(BASE)/app && $(GO) build \
		$(FLAGS) \
		-ldflags $(LDFLAGS) \
		-o $(CURDIR)/$(OUTPUT_DIR)/nginx-wrapper \
        $(GOPATH)/src/$(PACKAGE)/app/*.go

$(OUTPUT_DIR)/plugins/example.$(SHAREDLIBEXT): lib
	$(info $(M) building plugin [example]...) @
	$Q cd $(BASE)/plugins/example && $(GO) build \
		$(FLAGS) \
		-ldflags $(LDFLAGS) \
		-buildmode=plugin \
		-o $(CURDIR)/$(OUTPUT_DIR)/plugins/example.$(SHAREDLIBEXT) \
		$(GOPATH)/src/$(PACKAGE)/plugins/example/*.go

$(OUTPUT_DIR)/plugins: \
	$(OUTPUT_DIR)/plugins/example.$(SHAREDLIBEXT)

.PHONY: all
all: lib app plugins ## Build everything - including plugins

.PHONY: lib
lib: $(OUTPUT_DIR)/nginx-wrapper-lib ## Build common library

.PHONY: app
app: $(OUTPUT_DIR)/nginx-wrapper ## Build nginx-wrapper application

.PHONY: plugins
plugins: $(OUTPUT_DIR)/plugins ## Build all plugins

# Docker based CI tasks

.PHONY: ci-build-image
ci-build-image: ## Builds a Docker image containing all of the build tools for the project
	$(info $(M) building Docker build image) @
	$Q docker build -t nginx-wrapper-build $(CURDIR)/build/

.PHONY: ci-delete-image
ci-delete-image: ## Removes the Docker image containing all of the build tools for the project
	$(info $(M) removing Docker build image) @
	$Q docker rmi nginx-wrapper-build

.PHONY: ci-build-image-volume
ci-build-image-volume: ci-build-image ## Builds Docker volume that caches the gopath between operations
	$(info $(M) building Docker build volume to cache GOPATH) @
	$Q docker volume create nginx-wrapper-build-container
	$Q docker run --tty --rm --name nginx-wrapper-build-container --volume nginx-wrapper-build-container:/build/gopath nginx-wrapper-build /build/extract_gopath.sh

.PHONY: ci-delete-image-volume
ci-delete-image-volume: ## Removes the Docker volume proving gopath caching
	$(info $(M) removing Docker build volume) @
	$Q docker volume rm nginx-wrapper-build-container

.PHONY: ci-all-linters
ci-all-linters: ci-build-image-volume ## Runs all lint checks within a Docker container
	$(info $(M) running all linters) @
	$Q docker run --tty --rm --name nginx-wrapper-build-container --volume nginx-wrapper-build-container:/build/gopath --volume $(CURDIR):/build/src --workdir /build/src nginx-wrapper-build make all-linters

.PHONY: ci-run-all-checks
ci-run-all-checks: ci-all-linters ## Runs all build checks within a Docker container
	$(info $(M) running unit tests) @
	$Q docker run --tty --rm --name nginx-wrapper-build-container --volume nginx-wrapper-build-container:/build/gopath --volume $(CURDIR):/build/src --workdir /build/src nginx-wrapper-build make test-race

# Misc

.PHONY: changelog
LAST_VERSION      = $(shell git tag -l | egrep '^v[0-9]+\.[0-9]+\.[0-9]+$$' | sort --version-sort --field-separator=. --reverse | head -n1)
LAST_VERSION_HASH = $(shell git show --format=%H $(LAST_VERSION) | head -n1)
changelog: ## Outputs the changes since the last version committed
	$Q echo 'Changes since $(LAST_VERSION):'
	$Q git log --format="%s	(%h)" "$(LAST_VERSION_HASH)..HEAD" | \
		egrep -v '^(ci|chore|docs|build): .*' | \
		sed 's/: /:\t/g1' | \
		column -s "	" -t | \
		sed -e 's/^/ * /'

$(DISTPKGDIR)/nginx-wrapper-$(PLATFORM)_$(ARCH)-$(VERSION).gz: app
	$(info $(M) building compressed binary of nginx-wrapper app for $(PLATFORM)_$(ARCH))
	$Q mkdir -p $(DISTPKGDIR)
	$Q gzip --stdout --name --best $(OUTPUT_DIR)/nginx-wrapper > $(DISTPKGDIR)/nginx-wrapper-$(PLATFORM)_$(ARCH)-$(VERSION).gz

$(DISTPKGDIR)/nginx-wrapper-$(PLATFORM)_$(ARCH)-$(VERSION).gz.sha256sum: $(DISTPKGDIR)/nginx-wrapper-$(PLATFORM)_$(ARCH)-$(VERSION).gz
	$(info $(M) writing SHA256 checksum of nginx-wrapper app)
	$Q cd $(DISTPKGDIR); sha256sum nginx-wrapper-$(PLATFORM)_$(ARCH)-$(VERSION).gz > nginx-wrapper-$(PLATFORM)_$(ARCH)-$(VERSION).gz.sha256sum

package: $(DISTPKGDIR)/nginx-wrapper-$(PLATFORM)_$(ARCH)-$(VERSION).gz.sha256sum ## Builds packaged artifact of app

.PHONY: clean
clean: ; $(info $(M) cleaning...)	@ ## Cleanup everything
	@chmod -R +w $(GOPATH) || true
	@rm -rf target || true
	@rm -rf test/tests.* test/coverage.*

.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-24s\033[0m %s\n", $$1, $$2}' | sort

.PHONY: version
version:
	@echo $(VERSION)

release: all