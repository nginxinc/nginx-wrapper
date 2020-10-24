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
get-tools: $(GOLANGCILINT) $(GOCOVMERGE) $(GOCOV) $(GOCOVXML) $(GO2XUNIT) $(COMMITSAR) ## Retrieves and builds all the required tools

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
all-linters: golangci-lint commitsar ## Run all linters

.PHONY: fmt
fmt: ## Run source code formatter
	$(info $(M) running gofmt...) ## Run gofmt on all source files
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./... ); do \
		$(GOFMT) -l -w $$d/*.go || ret=$$? ; \
	 done ; exit $$ret