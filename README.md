# NGINX Wrapper

NGINX Wrapper (nginx-wrapper) is a NGINX process wrapper written in go 
that supports templating and plugins. Templating allows for modification
of `nginx` conf files before start up and on reload. Plugins allow for 
integrating templating and events by hooking into `nginx` events such as 
start up, reload and exit.

```
                     !WARNING!
THIS APPLICATION IS ALPHA AND STILL IN ACTIVE DEVELOPMENT
             APIS MAY CHANGE AT ANY TIME
```

## Use Cases

NGINX Wrapper can be thought of as a multipurpose tool that can be modified
for your specific needs. The plugin interface and templating engine
combine to allow for many simplifications of common application architectures.

* [12 Factor App](doc/use_cases.md#12-factor-app)
* [Templating Based on External Configuration](doc/use_cases.md#templating-based-on-external-configuration)
* [Plugins](doc/use_cases.md#plugins)
* [Agent Coprocess](doc/use_cases.md#agent-coprocess)
* [Service Discovery Integration](doc/use_cases.md#service-discovery-integration)
* [Secret Store Integration](doc/use_cases.md#secret-store-integration)
* [Custom Authentication Services](doc/use_cases.md#custom-authentication-services)
* [Service Mesh Proxy Integration](doc/use_cases.md#service-mesh-proxy-integration)

## Prerequisites

You will need to have NGINX installed on the same system as nginx-wrapper.
NGINX 1.17.9 is the earliest version of NGINX that has been tested with
the wrapper. If the `nginx` binary is not in your `PATH`, then you will
need to specify the path to the binary in the configuration with
the `nginx_binary` setting.

Currently, `nginx-wrapper` has only been tested on Linux. However, it
_may_ work on FreeBSD or Darwin (MacOS) because both platforms support
plugins. There is no plan for Windows support. 

## Installation

Copy the `nginx-wrapper` binary to your directory of choice.
Copy any plugins that you wish to use to a subdirectory `./plugins`
relative to the `nginx-wrapper` binary or define a plugin directory 
in your configuration.

Create a configuration file using a [viper](https://github.com/spf13/viper/)
compatible format (JSON, TOML, YAML, HCL, envfile and Java properties 
formats). There are examples available for 
[JSON](sample_configs/nginx-wrapper-example.json),
[YAML](sample_configs/nginx-wrapper-example.yml), and
[TOML](sample_configs/nginx-wrapper-example.toml).

Create an `nginx.conf.tmpl` file in the working directory or create 
a directory containing `nginx.conf.tmpl` and other templates. Then 
specify that directories' path with the configuration parameter
`conf_template_path`. See the [simple example](sample_configs/simple)
for how such a configuration may look.

For more detailed information on configuring NGINX Wrapper, refer to
the [configuration guide](doc/config.md). For help with setting up
template configuration, refer to the [templating guide](doc/templating.md).

## Usage

```
./nginx-wrapper help

NGINX Wrapper is a process wrapper that monitors NGINX for 
(start, reload, and exit) events, provides a templating framework for 
NGINX conf files and allows for plugins that extend its functionality.

Usage:
  nginx-wrapper [flags]
  nginx-wrapper [command]

Available Commands:
  debug       Display runtime configuration
  help        Help about any command
  run         run NGINX in a process wrapper
  version     Prints the nginx-wrapper version

Flags:
      --config string   path to configuration file (default "nginx-wrapper.toml")
  -h, --help            help for nginx-wrapper

Use "nginx-wrapper [command] --help" for more information about a command.
```

### Viewing Runtime Configuration

Run NGINX Wrapper's `debug` command to check your configuration
(replacing the sample config below with your own):
```shell script
./nginx-wrapper debug --config sample_configs/nginx-wrapper-example.toml
```

You should see the runtime configuration of NGINX Wrapper.

### Running NGINX

To start up NGINX via NGINX Wrapper (replacing the sample config below 
with your own):
 ```shell script
./nginx-wrapper run --config sample_configs/nginx-wrapper-example.toml
````

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[Apache 2.0](./LICENSE)