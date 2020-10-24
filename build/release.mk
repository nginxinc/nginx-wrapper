.PHONY: changelog
changelog: LAST_VERSION      := $(shell git tag -l | egrep '^v[0-9]+\.[0-9]+\.[0-9]+$$' | sort --version-sort --field-separator=. --reverse | head -n1)
changelog: LAST_VERSION_HASH := $(shell git show --format=%H $(LAST_VERSION) | head -n1)
changelog: ## Outputs the changes since the last version committed
	$Q echo 'Changes since $(LAST_VERSION):'
	$Q git log --format="%s	(%h)" "$(LAST_VERSION_HASH)..HEAD" | \
		egrep -v '^(ci|chore|docs|build): .*' | \
		sed 's/: /:\t/g1' | \
		column -s "	" -t | \
		sed -e 's/^/ * /'

$(DISTPKGDIR)/nginx-wrapper-$(PLATFORM)_$(ARCH)-v$(VERSION).gz: app
	$(info $(M) building compressed binary of nginx-wrapper app for $(PLATFORM)_$(ARCH))
	$Q mkdir -p $(DISTPKGDIR)
	$Q gzip --stdout --name --best $(OUTPUT_DIR)/nginx-wrapper > $(DISTPKGDIR)/nginx-wrapper-$(PLATFORM)_$(ARCH)-v$(VERSION).gz

$(DISTPKGDIR)/nginx-wrapper-$(PLATFORM)_$(ARCH)-v$(VERSION).gz.sha256sum: $(DISTPKGDIR)/nginx-wrapper-$(PLATFORM)_$(ARCH)-v$(VERSION).gz
	$(info $(M) writing SHA256 checksum of nginx-wrapper app)
	$Q cd $(DISTPKGDIR); sha256sum nginx-wrapper-$(PLATFORM)_$(ARCH)-v$(VERSION).gz > nginx-wrapper-$(PLATFORM)_$(ARCH)-v$(VERSION).gz.sha256sum

package: $(DISTPKGDIR)/nginx-wrapper-$(PLATFORM)_$(ARCH)-v$(VERSION).gz.sha256sum ## Builds packaged artifact of app