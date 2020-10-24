LDFLAGS       = '-X main.githash=$(GITHASH) -X main.buildstamp=$(DATE) -X main.appversion=$(VERSION)'
DEBUGFLAGS    = -gcflags "all=-N -l"
RELEASEFLAGS  = -tags release
FLAGS         = -trimpath
DEBUG        ?= 0
PLUGIN_TARGETS = $(foreach plugin,$(PLUGIN_ROOTS),${OUTPUT_DIR}/$(plugin)/$(notdir $(plugin)).${SHAREDLIBEXT})

# Define shared library extension based on platform
ifeq ($(PLATFORM),darwin)
	SHAREDLIBEXT = dylib
else
	SHAREDLIBEXT = so
endif

# Define binary output directory based on DEBUG flag
ifeq ($(DEBUG), 1)
	FLAGS += $(DEBUGFLAGS)
	OUTPUT_DIR := target/debug
else
	FLAGS += $(RELEASEFLAGS)
	OUTPUT_DIR := target/release
endif

${OUTPUT_DIR}: target
	@mkdir -p $(CURDIR)/$@

.PHONY: deps/%
deps/%:
	$(info $(M) downloading dependencies for $(notdir $@)...) @
	$Q cd $(or $(SRC_DIR_$(subst deps/,,$@)),$(dir $(subst deps/,,$@))) && $(GO) mod download

.PHONEY: deps
deps: deps/nginx-wrapper-lib deps/nginx-wrapper $(addprefix deps/,$(PLUGIN_ROOTS)) ## Download dependencies

# Compile source directory into binary
${OUTPUT_DIR}/%: ${OUTPUT_DIR}
	$(info $(M) building $(notdir $@)...) @
	$Q cd $(or $(SRC_DIR_$(notdir $@)),$(dir $(subst ${OUTPUT_DIR}/,,$@))) && $(GO) build \
		$(FLAGS) \
		-ldflags $(LDFLAGS) \
		-o $(CURDIR)/$@

.PHONEY: lib
lib: FLAGS += -buildmode=plugin
lib: ${OUTPUT_DIR}/nginx-wrapper-lib

.PHONEY: plugins
plugins: FLAGS += -buildmode=plugin
plugins: $(PLUGIN_TARGETS)

.PHONEY: build
build: lib ${OUTPUT_DIR}/nginx-wrapper plugins