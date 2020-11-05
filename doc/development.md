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
[CentOS7](../build/bin/install_centos7_dependencies.sh),
[Ubuntu](../build/bin/install_ubuntu_dependencies.sh), or 
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

## Developing Plugins

### Plugin Types

There are two types of plugins used with the wrapper: embedded plugins and
external plugins (often referred to as just plugins). Both conform to the same 
API. However, they differ in how they are loaded.

External plugins are distributed as a [go plugin shared library](https://golang.org/pkg/plugin/)
built with the `-buildmode=plugin` flag. This allows for external plugins to
be dynamically loaded and distributed as separate files.

Embedded plugins use the same interface as external plugins, but are compiled
as part of the wrapper application. Each embedded plugin's source code is in 
a directory within the [app/embeddedplugins](../app/embeddedplugins) directory.
Ideally, embedded plugins will take on no dependencies on wrapper 
application code and only depend on the API provided by the nginx wrapper 
library. By doing so, embedded plugins can be freely removed and changed to
become an external plugins.

Additionally, when developing external plugins it may be useful to develop
them initially as embedded plugins in order to use a debugger.

### Adding a New External Plugin

#### Creating the Source Tree

Create a new directory within the directory [./plugins](../plugins). In that 
directory, create a file called `go.mod` with the following contents and
substituting values marked as `<value>` appropriately.
```
module github.com/nginxinc/nginx-wrapper/plugins/<plugin name>

go 1.15

replace github.com/nginxinc/nginx-wrapper/lib => ../../lib

require (
	github.com/nginxinc/nginx-wrapper/lib v<current wrapper version>
    // Add this line if you want to support logging
	github.com/go-eden/slf4go v<same version of slf4go as used in lib>
)
```

Create a new file named `main.go` within your plugin directory. Add the following
content to the file.

```go
package main

import (
    slog "github.com/go-eden/slf4go"
    "github.com/nginxinc/nginx-wrapper/lib/api"
)

// PluginName contains the name/id of the plugin.
const PluginName string = "<my plugin name without spaces>"
var log = slog.NewLogger(PluginName)

func Metadata(_ api.Settings) map[string]interface{} {
		return map[string]interface{}{
    		"name": PluginName,
    		"config_defaults": map[string]interface{}{},
    	}
}

func Start(context api.PluginStartupContext) error {
    log.Infof("%s plugin started", PluginName)
	return nil
}
```

We have now created the minimum configuration for a plugin.

#### Compiling the Plugin

To compile all plugins, you can issue the following make command:
```
make build-plugins
```
Alternatively, if you want to only compile a single plugin:
```
PLUGIN_ROOTS=plugins/<plugin dir> make clean build-plugins
```

#### Running the Plugin

Once your plugin is compiled, copy the binary created from the path
`target/<os_arch>/<release|debug>/plugins/<plugin name>/<plugin name>.<so|dynlib>`
to the plugin path that you have specified in your NGINX Wrapper configuration. 
The configuration variable is called `plugin_path` and by default it is set 
to a subdirectory named `plugins`. For example, if you are running the 
wrapper in the directory `/opt/nginx-wrapper`, you would copy your newly
compiled plugin into `/opt/nginx-wrapper/plugins`.

```
$ cd /opt/nginx-wrapper

$ ls
bin  nginx-wrapper.toml  plugins  run  template

$ grep plugin_path nginx-wrapper.toml 
  plugin_path = "./plugins"

$ ls plugins/
myplugin.so
```

Once the plugin is copied into the `plugins` directory, you will need to alter
your configuration to enable it. This is done by adding the name (the name 
you specified in `const PluginName`) of the plugin to the configuration value 
`enabled_plugins`.

```
$ grep enabled_plugins nginx-wrapper.toml 
  enabled_plugins = [ "coprocess", "template" ]

$ vim nginx-wrapper.toml # edit the file to include myplugin

$ grep enabled_plugins nginx-wrapper.toml 
  enabled_plugins = [ "coprocess", "template", "myplugin" ] 
```

Now, let's run `nginx-wrapper`.

```
./bin/nginx-wrapper --config nginx-wrapper.toml run
INFO   [2020-10-29T18:42:49Z] load-plugin: started plugin: [template]      
INFO   [2020-10-29T18:42:49Z] myplugin: myplugin plugin started                    
INFO   [2020-10-29T18:42:49Z] load-plugin: started plugin: [myplugin]          
INFO   [2020-10-29T18:42:49Z] nginx: 455#455: using the "epoll" event method 
INFO   [2020-10-29T18:42:49Z] nginx: 455#455: nginx/1.19.3                 
INFO   [2020-10-29T18:42:49Z] nginx: 455#455: built by gcc 8.3.0 (Debian 8.3.0-6) 
INFO   [2020-10-29T18:42:49Z] nginx: 455#455: OS: Linux 5.4.0-48-generic   
INFO   [2020-10-29T18:42:49Z] nginx: 455#455: getrlimit(RLIMIT_NOFILE): 1048576:1048576 
INFO   [2020-10-29T18:42:49Z] nginx: 455#455: start worker processes       
INFO   [2020-10-29T18:42:49Z] nginx: 455#455: start worker process 456     
```

Here we were able to verify that the plugin's `Start` function was called 
because of the log line `myplugin: myplugin plugin started`. We can do
additional troubleshooting by running `nginx-wrapper debug`.

```
$ ./bin/nginx-wrapper --config nginx-wrapper.toml debug
time="2020-10-29T18:46:05Z" level=info msg="load-plugin: loaded plugin: [template]"
time="2020-10-29T18:46:05Z" level=info msg="load-plugin: loaded plugin: [myplugin]"
                             conf_path: /opt/nginx-wrapper/run/conf
                       enabled_plugins: [template myplugin]
...
```

From the output above, you can see that the `myplugin` plugin was loaded and 
listed as an enabled plugin.