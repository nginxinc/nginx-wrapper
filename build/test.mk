TIMESTAMP 	     := $(shell date -u +"%Y-%m-%dT%H-%M-%SZ")
TESTPKGFILES     := '{{ if or .TestGoFiles .XTestGoFiles }}{{ .ImportPath }}{{ end }}'
COVERAGE_MODE    := atomic
COVERAGE_PROFILE := profile.out
COVERAGE_XML     := coverage.xml
COVERAGE_HTML    := index.html

# The running test message is separated so that the unit test results can all
# be concatenated and viewed in one long list
.PHONY: test-message
test-message:
	$(info $(M) running $(NAME:%=% )tests...) @

.PHONY: run-tests/%
run-tests/%:
	$Q cd $(or $(SRC_DIR_$(subst run-tests/,,$@)),$(subst run-tests/,,$@)) && \
		$(GO) test -timeout $(TIMEOUT)s $(ARGS) ./...

TEST_TARGETS := test-default test-bench test-short test-verbose test-race
.PHONY: $(TEST_TARGETS) test-xml check test test-coverate-tools test-coverage
test-bench:   ARGS=-run=__absolutelynothing__ -bench=. ## Run benchmarks
test-short:   ARGS=-short        ## Run only short tests
test-verbose: ARGS=-v            ## Run tests in verbose mode with coverage reporting
test-race:    ARGS=-race         ## Run tests with race detector
$(TEST_TARGETS): NAME=$(MAKECMDGOALS:test-%=%)
$(TEST_TARGETS): test
.PHONY: test
check test: test-message run-tests/nginx-wrapper-lib run-tests/nginx-wrapper $(addprefix run-tests/,$(PLUGIN_ROOTS)) ## Run tests

test/coverage/${TIMESTAMP}:
	$Q mkdir -p $@

test/coverage/${TIMESTAMP}/%: test/coverage/${TIMESTAMP}
	$Q cd $(subst test/coverage/${TIMESTAMP}/,,$@) && \
		PKGS="$$($(GO) list ./... | xargs)"; \
		TESTPKGS="$$($(GO) list -f ${TESTPKGFILES} | xargs) $${PKGS}"; \
		for pkg in $${TESTPKGS}; do \
		  	LIST_DEPS="$$($(GO) list -f '{{ join .Deps "\n" }}' $$pkg)"; \
		  	COVER_PKG="$$(echo $${LIST_DEPS} | grep '^$(PACKAGE)/' | tr '\n' ',')"; \
		  	COVER_PROFILE="$(CURDIR)/test/coverage/${TIMESTAMP}/$$(echo $$pkg | tr "/" "-").cover"; \
			$(GO) test \
				-coverpkg="$${COVER_PKG}$${pkg}" \
				-covermode=$(COVERAGE_MODE) \
				-coverprofile="$${COVER_PROFILE}" $$pkg ;\
		done

test/coverage/${TIMESTAMP}/$(COVERAGE_PROFILE):
	$Q $(GOCOVMERGE) $(dir $@)/*.cover > $(dir $@)/$(COVERAGE_PROFILE)

# We have to be a source code directory for go tool cover to correctly ingest source files
test/coverage/${TIMESTAMP}/$(COVERAGE_HTML): test/coverage/${TIMESTAMP}/profile.out
	$Q cd app && $(GO) tool cover -html=$(CURDIR)/$(dir $@)/$(COVERAGE_PROFILE) -o $(CURDIR)/$(dir $@)$(COVERAGE_HTML)

# We have to be a source code directory for go tool cover to correctly ingest source files
test/coverage/${TIMESTAMP}/$(COVERAGE_XML): test/coverage/${TIMESTAMP}/profile.out
	$Q cd app && $(GOCOV) convert $(CURDIR)/$(dir $@)/$(COVERAGE_PROFILE) | $(GOCOVXML) > $(CURDIR)/$(dir $@)$(COVERAGE_XML)

.PHONY: test-coverage-lib
test-coverage-lib: test/coverage/${TIMESTAMP}/lib

.PHONY: test-coverage-app
test-coverage-app: test/coverage/${TIMESTAMP}/app

.PHONY: test-coverage-plugins
test-coverage-plugins: $(addprefix test/coverage/${TIMESTAMP}/,$(PLUGIN_ROOTS))

.PHONY: test-coverage
test-coverage: test-message test-coverage-lib test-coverage-app test-coverage-plugins
test-coverage: test/coverage/${TIMESTAMP}/$(COVERAGE_HTML) test/coverage/${TIMESTAMP}/$(COVERAGE_XML)