DISTPKGDIR        := target/package
PLUGIN_PKGS        = $(foreach plugin,$(PLUGIN_ROOTS),${DISTPKGDIR}/$(plugin)/$(notdir $(plugin)).${SHAREDLIBEXT}-$(PLATFORM)_$(ARCH)-v$(VERSION).gz)
LAST_VERSION       = $(shell git tag -l | egrep '^v[0-9]+\.[0-9]+\.[0-9]+$$' | sort --version-sort --field-separator=. --reverse | head -n1)
LAST_VERSION_HASH  = $(shell git show --format=%H $(LAST_VERSION) | head -n1)
CHANGES            = $(shell git log --format="%s	(%h)" "$(LAST_VERSION_HASH)..HEAD" | \
					 	egrep -v '^(ci|chore|docs|build): .*' | \
                        sed 's/: /:\t/g1' | \
                        column -s "	" -t | \
                        sed -e 's/^/ * /' | \
                        tr '\n' '\1')

.PHONY: changelog
.ONESHELL: changelog
changelog: ## Outputs the changes since the last version committed
	$Q echo 'Changes since $(LAST_VERSION):'
	$Q echo "$(CHANGES)" | tr '\1' '\n'

$(DISTPKGDIR):
	$Q mkdir -p $(DISTPKGDIR)

.PRECIOUS: $(DISTPKGDIR)/%-$(PLATFORM)_$(ARCH)-v$(VERSION).gz.sha256sum
.ONESHELL: $(DISTPKGDIR)/%-$(PLATFORM)_$(ARCH)-v$(VERSION).gz.sha256sum
$(DISTPKGDIR)/%-$(PLATFORM)_$(ARCH)-v$(VERSION).gz.sha256sum: $(DISTPKGDIR)/%-$(PLATFORM)_$(ARCH)-v$(VERSION).gz
	$(info $(M) writing SHA256 checksum of $* to ${@F}) @
	$Q cd $(DISTPKGDIR)
	$Q sha256sum $*-$(PLATFORM)_$(ARCH)-v$(VERSION).gz > $*-$(PLATFORM)_$(ARCH)-v$(VERSION).gz.sha256sum

.PRECIOUS: $(DISTPKGDIR)/%-$(PLATFORM)_$(ARCH)-v$(VERSION).gz
$(DISTPKGDIR)/%-$(PLATFORM)_$(ARCH)-v$(VERSION).gz: ${OUTPUT_DIR}/% $(DISTPKGDIR)
	$(info $(M) building compressed binary of $* for $(PLATFORM)_$(ARCH)) @
	$Q mkdir -p ${@D}
	$Q gzip --stdout --name --best $(OUTPUT_DIR)/$* > $@

.PHONY: package-app
package-app: build-app $(DISTPKGDIR)/nginx-wrapper-$(PLATFORM)_$(ARCH)-v$(VERSION).gz.sha256sum
.PHONY: package-lib
package-lib: build-lib $(DISTPKGDIR)/nginx-wrapper-lib-$(PLATFORM)_$(ARCH)-v$(VERSION).gz.sha256sum
.PHONY: package-plugins
package-plugins: build-plugins $(addsuffix .sha256sum,$(PLUGIN_PKGS))

.PHONY: package
package: package-lib package-app package-plugins ## Builds packaged artifacts for all source trees (app, lib, plugins)

.PHONY: version
version: ## Outputs the current version
	$Q echo "Version: $(VERSION)"
	$Q echo "Commit : $(GITHASH)"

.PHONY: version-update
.ONESHELL: version-update
version-update: ## Prompts for a new version
	$(info $(M) updating repository to new version) @
	$Q echo "  last committed version: $(LAST_VERSION)"
	$Q echo "  .version file version : v$(VERSION)"
	read -p "  Enter new version in the format (MAJOR.MINOR.PATCH): " version
	$Q echo "$$version" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$$' || \
		(echo "invalid version identifier: $$version" && exit 1) && \
	echo -n $$version > $(CURDIR)/.version

.PHONY: version-apply
.ONESHELL: version-apply
version-apply: target ## Applies the version to resources in the repository
	$(info $(M) applying version $(VERSION) to repository) @
	$Q echo "  updating CHANGELOG.md with latest changes"
	LAST_VERSION_CHANGELOG_LINE="$$(grep --line-number $(LAST_VERSION) CHANGELOG.md | cut -f1 -d:)"
	CHANGELOG_HEADER_LINE="$$(expr $$LAST_VERSION_CHANGELOG_LINE - 1)"
	$Q head --lines=$$CHANGELOG_HEADER_LINE CHANGELOG.md > target/CHANGELOG.md.new
	$Q echo "## $(VERSION)" >> target/CHANGELOG.md.new
	$Q echo "$(CHANGES)" | tr '\1' '\n' >> target/CHANGELOG.md.new
	$Q tail --lines=+$$LAST_VERSION_CHANGELOG_LINE CHANGELOG.md >> target/CHANGELOG.md.new
	$Q mv target/CHANGELOG.md.new CHANGELOG.md
	$Q echo "  updating app/go.mod to use nginx-wrapper-lib with latest version"
	$Q sed --in-place -e "s|github.com/nginxinc/nginx-wrapper/lib\s\{1,\}v[0-9]\{1,\}.[0-9]\{1,\}.[0-9]\{1,\}|github.com/nginxinc/nginx-wrapper/lib v$(VERSION)|" app/go.mod
	$Q echo "  updating plugins/*/go.mod to use nginxwrapper-lib with latest version"
	$Q find plugins -maxdepth 2 -mindepth 1 -type f -name go.mod -exec \
		sed --in-place -e "s|github.com/nginxinc/nginx-wrapper/lib\s\{1,\}v[0-9]\{1,\}.[0-9]\{1,\}.[0-9]\{1,\}|github.com/nginxinc/nginx-wrapper/lib v$(VERSION)|" '{}' \;
	make package
	$Q echo "  updating nginx-wrapper version in recipes/*/Dockerfile"
	$Q find recipes -maxdepth 2 -mindepth 1 -type f -name Dockerfile -exec \
		sed --in-place -e "s|ENV NGINX_WRAPPER_VERSION\s\{1,\}v[0-9]\{1,\}.[0-9]\{1,\}.[0-9]\{1,\}|ENV NGINX_WRAPPER_VERSION v$(VERSION)|" '{}' \;
	$Q echo "  updating nginx-wrapper checksum in recipes/*/Dockerfile"
	WRAPPER_CHECKSUM="$$(cat $(DISTPKGDIR)/nginx-wrapper-$(PLATFORM)_$(ARCH)-v$(VERSION).gz.sha256sum | cut -f1 -d' ')"
	$Q find recipes -maxdepth 2 -mindepth 1 -type f -name Dockerfile -exec \
		sed --in-place -e "s|ENV NGINX_WRAPPER_CHECKSUM\s\{1,\}.\{1,\}|ENV NGINX_WRAPPER_CHECKSUM $$WRAPPER_CHECKSUM|" '{}' \;

.PHONY: version-commit
.ONESHELL:
version-commit: ## Prompts to commit the current version to git
	$(info $(M) committing version $(VERSION) to repository) @
	$Q git commit --edit -m "chore: incremented version to v$(VERSION)" \
		.version \
		CHANGELOG.md \
		app/go.mod \
		$(addsuffix /go.mod,$(PLUGIN_ROOTS)) \
		$$(find recipes -maxdepth 2 -mindepth 1 -type f -name Dockerfile | xargs)
	$Q git tag v$(VERSION)
	$Q git tag app/v$(VERSION)
	$Q git tag lib/v$(VERSION)

.ONESHELL: $(DISTPKGDIR)/release_notes.md
$(DISTPKGDIR)/release_notes.md: $(DISTPKGDIR)
	$(info $(M) building release notes) @
	$Q echo 'Changes since last release:' > $(DISTPKGDIR)/release_notes.md
	$Q echo '```' >> $(DISTPKGDIR)/release_notes.md
	$Q echo "$(CHANGES)" | tr '\1' '\n' >> $(DISTPKGDIR)/release_notes.md
	$Q echo '```' >> $(DISTPKGDIR)/release_notes.md
	$Q echo 'SHA256 Checksums:' >> $(DISTPKGDIR)/release_notes.md
	$Q echo '```' >> $(DISTPKGDIR)/release_notes.md
	$Q find $(DISTPKGDIR) -type f -name \*.sha256sum -exec cat '{}' \; >> $(DISTPKGDIR)/release_notes.md
	$Q echo '```' >> $(DISTPKGDIR)/release_notes.md

.PHONY: release
.ONESHELL: release
release: clean version-update version-apply $(DISTPKGDIR)/release_notes.md version-commit
	$(info $(M) pushing changes to github) @
	$Q git push --tags origin master
	RELEASE_PKGS="$$(find $(DISTPKGDIR) -type f -name \*.gz | xargs)"
	PRERELEASE="$$(echo $(VERSION) | grep -qE '^0.[0-9]+\.[0-9]+$$' && echo '--prerelease')"
	$Q gh release create v$(VERSION) $$RELEASE_PKGS \
		--notes-file $(DISTPKGDIR)/release_notes.md \
		--title "v$(VERSION)" \
		$$PRERELEASE