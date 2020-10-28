# NGINX Wrapper Development Guide

## Directory Structure

The project has the following organizational structure: 

```
/
├── .github                 Github CI workflow and action configuration
├── .gopath                 Default GOPATH when not specified in environment
├── app                     Source tree for the nginx-wrapper app
├── build                   Supporting build files and scripts 
├── doc                     Supplemental documentation
├── lib                     Source tree for the nginx-wrapper library
├── plugins                 Contains the source tree for additional plugins
├── recipes                 Contains directories showing off example uses
├── sample_configs          Sample configuration files
├── target                  Build artifacts - binaries, packages, etc 
└── test                    Test run files - code coverage reports, etc
```

## Building

The `nginx-wrapper` app, the `nginx-wrapper-lib` library and all the 
plugins contained in the `plugins/*` directory can be built using the provided
[Makefile](../Makefile) using [GNU Make](https://www.gnu.org/software/make/)
4.1+. Alternative, the source roots can be built directly using `go` from 
within the source directory, but that method will not produce the expected 
artifacts for the release process. Nonetheless, that method may be useful 
within an IDE environment.

### Environments

The following build environments have been tested:
 * Debian Linux amd64
 * Ubuntu Linux amd64
 * Darwin (macos) amd64

The following build environments *should* work but have not been tested:
 * Redhat/CentOS Linux amd64
 * SUSE Linux amd64
 * Linux with other architectures
 * Windows in WSL
 
The following build environments *might* work but have not been tested:
 * Alpine Linux with musl
 * FreeBSD
 * Solaris/illumos

The following build environments *won't* work and have not been tested:
 * Windows

### Build Dependencies

The build process depends upon a number of utilities being available in the
shell `PATH`. To check to see if you have the needed dependencies installed
run the script [check_build_dependencies.sh](../build/bin/check_build_dependencies.sh).

```
$ ./check_build_dependencies.sh
▶ Checking to see if bash is version 4+... yes - [version 4]
▶ Checking to see if uname is installed... yes
▶ Checking to see if awk is installed... yes
▶ Checking to see if cat is installed... yes
▶ Checking to see if column is installed... yes
...
``` 

There are helper scripts available to install the required dependencies on
[MacOS](../build/bin/install_macos_homebrew_dependencies.sh),
[Ubuntu](../build/bin/install_ubuntu_dependencies.sh) or 
[Debian](../build/bin/install_debian_dependencies.sh).

If you are running on MacOS and using [homebrew](https://brew.sh/), you will
need to give preference to GNU CLI applications in the `PATH` by doing:
```
export PATH="/usr/local/opt/make/libexec/gnubin:$PATH"
```

### Using the Build System

The tasks available in the build system can be displayed by issuing `make` or `make help`
from within the project's root:
```
make help

all build                Compiles all source trees (app, lib, plugins)
changelog                Outputs the changes since the last version committed
check test               Run all tests (app, lib, plugins)
clean                    Cleanup everything
commitsar                Run git commit linter
deps                     Download dependencies
docker-all-linters       Runs all lint checks within a Docker container
docker-build-image       Builds a Docker image containing all of the build tools for the project
docker-build-volume      Builds Docker volume that caches the gopath between operations
docker-delete-image      Removes the Docker image containing all of the build tools for the project
docker-delete-volume     Removes the Docker volume proving gopath caching
docker-run-all-checks    Runs all build checks within a Docker container
docker-run-shell         Opens a shell on the Docker container
docker-run-tests         Runs all tests within a Docker container
fmt                      Run source code formatter on all files
get-tools                Retrieves and builds all the required tools
golangci-lint            Run golangci-lint check
linters                  Run all linters
package                  Builds packaged artifacts for all source trees (app, lib, plugins)
release                  Releases a new version to GitHub
test-bench               Run benchmarks
test-coverage            Run tests with code coverage
test-race                Run tests with race detector
test-short               Run only short tests
test-verbose             Run tests in verbose mode with coverage reporting
version                  Outputs the current version
version-apply            Applies the version to resources in the repository
version-commit           Prompts to commit the current version to git
version-update           Prompts for a new version
```

### Compiling

`make clean` will remove all binaries, packages and test coverage artifacts.

`make deps` will download the go dependencies for all source trees.

`make build` will build all artifacts (nginx-wrapper app, nginx-wrapper
library and plugins).

`make build-app` will build only the nginx-wrapper app.

`make build-lib` will build only the nginx-wrapper library.

`make build-plugins` will build all of the nginx-wrapper plugins in the
`plugins/` directory.

### Testing

`make package` will build packaged artifacts for all source trees 
(nginx-wrapper app, nginx-wrapper library and plugins).


`make test` will run the automated testing suite for all source trees 
(nginx-wrapper app, nginx-wrapper library and plugins).

`make run-tests/nginx-wrapper` will run the automated testing suite only for
the nginx-wrapper app.

`make run-tests/nginx-wrapper-lib` will run the automated testing suite only
for the nginx-wrapper library.

`make run-tests-plugins` will run the automated testing suite only for all
plugins.

`make run-tests/plugins/<plugin name>` will run the automated testing suite
only for the specified plugin.

`make test-coverage` will run the automated testing suite and output a code
coverage report for all source trees (nginx-wrapper app, nginx-wrapper 
library and plugins).

`make test-coverage-app` will run the automated testing suite and output a
code coverage report for the nginx-wrapper app.

`make test-coverage-lib` will run the automated testing suite and output a 
code coverage report for the nginx-wrapper library.

`make test-coverage-plugins` will run the automated testing suite and output
a code coverage report for all plugins.

### Linting

`make linters` will run all lint checks on all source trees.

`make golangci-lint` will run the [golangci-lint](https://github.com/golangci/golangci-lint) 
static code analysis tool on all source trees.

`make commitsar` will run the git check in linter.

### Versioning

`make changelog` will output the changes since the last tagged version.

`make version` will output the current version as stored in the `.version` 
file.

`make version-apply` will modify update version dependent files in the 
repository to the version stored in the `.version` file.

`make version-commit` will commit all files that had version updates to git
and create new tags for the version.

`make version-update` will prompt you for a new version, validate it and write
it to the  `.version` file

### Running Builds in Docker

`make docker-all-linters` will runs all lint checks within a Docker container.

`make docker-build-image` will build a Docker image containing the build tools
for the project.

`make docker-build-volume` will build a Docker volume that caches the gopath 
between docker operations.

`make docker-delete-image` will remove the Docker image containing the build
tools for the project.

`make docker-delete-volume` will remove the Docker volume proving gopath caching.

`make docker-run-all-checks` will run all build checks within a Docker 
container (docker-all-linters, docker-run-tests).

`make docker-run-shell` will open a shell on the Docker container.

`make docker-run-tests` will run all tests within a Docker container.

If you are developing on a non-Linux system, you may find the
`docker-*` commands to be quite useful. The `docker-run-shell`
command is useful for providing a known working Linux configuration in which
the build tools will function and compile Linux compatible artifacts.

### Releasing

`make release` will release a new version of the nginx-wrapper app, nginx-wrapper library
and plugins to GitHub.