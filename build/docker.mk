.PHONY: docker-build-image
docker-build-image: ## Builds a Docker image containing all of the build tools for the project
	$(info $(M) building Docker build image) @
	$Q docker build -t nginx-wrapper-build $(CURDIR)/build/

.PHONY: docker-delete-image
docker-delete-image: ## Removes the Docker image containing all of the build tools for the project
	$(info $(M) removing Docker build image) @
	$Q docker rmi nginx-wrapper-build

.PHONY: docker-build-volume
docker-build-volume: docker-build-image ## Builds Docker volume that caches the gopath between operations
	$(info $(M) building Docker build volume to cache GOPATH) @
	$Q docker volume create nginx-wrapper-build-container
	$Q docker run --tty --rm --name nginx-wrapper-build-container --volume nginx-wrapper-build-container:/build/gopath nginx-wrapper-build /build/extract_gopath.sh

.PHONY: docker-delete-volume
docker-delete-volume: ## Removes the Docker volume proving gopath caching
	$(info $(M) removing Docker build volume) @
	$Q docker volume rm nginx-wrapper-build-container

.PHONY: docker-all-linters
docker-all-linters: docker-build-volume ## Runs all lint checks within a Docker container
	$(info $(M) running all linters) @
	$Q docker run --tty --rm --name nginx-wrapper-build-container --volume nginx-wrapper-build-container:/build/gopath --volume $(CURDIR):/build/src --workdir /build/src nginx-wrapper-build make all-linters

.PHONY: docker-run-tests
docker-run-tests: ## Runs all tests within a Docker container
	$(info $(M) running tests in Docker) @
	$Q docker run --tty --rm --name nginx-wrapper-build-container --volume nginx-wrapper-build-container:/build/gopath --volume $(CURDIR):/build/src --workdir /build/src nginx-wrapper-build make test-race

.PHONY: docker-run-all-checks
docker-run-all-checks: docker-all-linters docker-run-tests ## Runs all build checks within a Docker container

.PHONY: docker-run-shell
docker-run-shell: docker-build-volume ## Opens a shell on the Docker container
	$(info $(M) opening Docker shell) @
	$Q docker run --tty --interactive --rm --name nginx-wrapper-build-container --volume nginx-wrapper-build-container:/build/gopath --volume $(CURDIR):/build/src --workdir /build/src nginx-wrapper-build bash || true