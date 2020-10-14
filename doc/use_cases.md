# Use Cases

NGINX Wrapper can be thought of as a multipurpose tool that can be modified
for your specific needs. The plugin interface and templating engine
combine to allow for many simplifications of common application architectures.

## 12 Factor App

One of the prevailing forms of application composition for cloud-native and
container-native environments is [The Twelve-Factor App](https://12factor.net/).
NGINX Wrapper allows NGINX to follow this pattern more closely by:

* Allowing for templated configuration files, so that configuration can be
  the same regardless of environment (eg production, staging). This also
  reduces barriers to managing configuration files by source control
  systems.
* NGINX deployment dependencies such as agents (for APM, logging, statistics, 
  etc)  can be easily declared and packaged together in a single unit. This 
  reduces the scope of externalized dependencies that need to be coordinating 
  in sync with NGINX.
* All configuration parameters can be set via environment variables or 
  configuration files (`JSON`, `toml`, `yaml`) because NGINX Wrapper 
  uses [viper](https://github.com/spf13/viper) for configuration. All 
  configuration settings are available in the templating engine. This model
  allows for flexible integration with different deployment and configuration
  management systems because the runtime settings of NGINX can be injected
  via environment variables or programmatically applied to common machine-
  readable file formats.
* NGINX Wrapper simplifies the running of NGINX in an unprivileged (not 
  root) mode when needed by creating the needed directory structure before
  NGINX starts.
* NGINX Wrapper supports SIGHUP for dynamic reconfiguration and SIGTERM for
  graceful shutdown.
* Logs can be output to `STDOUT`, `STDERR` or to a file with the default 
  operation for all logs to be written to `STDOUT`.

## Templating Based on External Configuration

When the wrapper starts with the template plugin enabled, it can 
template in NGINX runtime configuration from system environment variables. 
In containerized environments, it is common to inject runtime configuration 
via environment variables. Configuring NGINX is entirely a file based 
operation and a tool is needed to bridge the paradigms. Rather than rely 
on `sed` or `awk` in  wrapper shell scripts, a more systematic approach 
can be taken with the NGINX Wrapper.

Additionally, go provides a number of container friendly libraries
that allow for us to introspect the running container and determine the
number of CPUs in the container's `cpuset` and hand off that information
to the templating engine in order to configure the number of NGINX workers
in a more container friendly manner (using [`runtime.NumCPU()`]
(https://www.golangprograms.com/find-out-how-many-logical-processors-used-by-current-process.html))
by calling [`sched_getaffinity`](https://github.com/golang/go/issues/3921) 
on Linux.

## Plugins

NGINX Wrapper has a plugin interface that allows for a plugin to execute
events upon NGINX pre-start, start, start-worker, pre-reload, reload,
exit, and exit-worker. Additionally, plugins can run background processes 
while NGINX is running and issue reload commands to NGINX. This allows
for dynamic reconfiguration based on an external data source.

Plugins can also access all the configuration data in the wrapper and 
add their own configuration entries. Plugin configuration can be used
within the templating system to customize the operation of NGINX.

## Agent Coprocess

There are many products that offer complimentary functionality to NGINX
by using agents (processes) that run alongside NGINX. Agent lifecycle 
(startup and shutdown) is often tied to the lifecycle of NGINX. When
running in a Docker container, the default behavior is to run one process
per container. This model makes running coprocess agents difficult. 
[Solutions](https://docs.docker.com/config/containers/multi-service_container/)
to this problem include Kubernetes sidecar containers, running
process managers like `systemd`, or using wrapper scripts and the `--init`
flag in Docker. However, all the solutions increase the complexity
of the deployment artifact and management of the agents.

Agents running as [coprocesses](https://en.wikipedia.org/wiki/Coprocess) 
allow for a simple configuration of NGINX and a coordinated startup and
shutdown. They also allow for a single deployment artifact to be created
where all the relevant binaries needed to run NGINX in production are 
located in a single place.

## Service Discovery Integration

Service discovery tools such as [Consul](https://www.consul.io/), 
[etcd](https://etcd.io/), and [ZooKeeper](https://zookeeper.apache.org/) 
can be integrated with NGINX Wrapper as plugins and used as data sources
for configuring NGINX. Service discovery backends can be actively monitored
for changes. Upon change, plugins can update configuration and restart NGINX.

## Secret Store Integration

Support for secret stores such as [Vault](https://www.vaultproject.io/) and 
[Keywhiz](https://square.github.io/keywhiz/) can be added to NGINX Wrapper
as plugins. Like service discovery, secret stores can be a data source that
drives configuration of NGINX dynamically.

## Custom Authentication Services

Plugins can implement their own authentication interfaces that are called
over [Unix domain sockets](https://en.wikipedia.org/wiki/Unix_domain_socket) 
by NGINX using the [`auth_request` module](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html).
This implementation can allow for easier customized authentication schemes
written in go. Additionally, the authentication engine can be deployed as
a single artifact along with NGINX and communication with the service can
bypass TCP by using sockets.

## Service Mesh Proxy Integration

NGINX Wrapper simplifies the configuration of NGINX for use as a proxy server
with a Service Mesh by allowing for plugins that use [service discover](#service-discovery-integration)
and a [custom authentication service](#custom-authentication-services) that are
implemented in go which call out to the service mesh authentication service.