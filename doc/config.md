# Configuration Guide

Configuration can be specified using environment variables, or a single CLI 
flag `--config` that points to a configuration file. Internally NGINX 
Wrapper uses [viper](https://github.com/spf13/viper) for configuration.

The same configuration subsystem used in application start up and plugins
is the data source when applying templates. This allows for the configuration to be
applied in a tiered fashion, to be dynamically updated by plugins, and for
events to update the runtime configuration state.

## Configuration Parameters

### Core Parameters

| Name                            | Default                                             | Description                                                                                                                                                                                                                                                                                         |
|---------------------------------|-----------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `enabled_plugins`               |                                                     | List of plugins that are enabled to run in the NGINX Wrapper                                                                                                                                                                                                                                        |
| `host_id`                       | contents of /etc/machine-id or hostname             | Unique string representing the host NGINX Wrapper is running on                                                                                                                                                                                                                                     |
| `modules_path`                  | `{nginx -V}`                                        | Path to the nginx modules directory - by default this is parsed from the output of `nginx -V`                                                                                                                                                                                                       |
| `nginx_binary`                  | `nginx`                                             | Path to the nginx binary                                                                                                                                                                                                                                                                            |
| `plugin_path`                   | `./plugins`                                         | Path to read nginx-wrapper plugins from                                                                                                                                                                                                                                                             |
| `run_path`                      | `{os.TempDir()}/nginx-wrapper`                      | Runtime directory of the nginx process. Configuration and other related files will be copied into this directory.                                                                                                                                                                                   |
| `vcpu_count`                    | `{nproc}`                                           | Core count that can be used for templating the number of worker processes that we want NGINX to start. By default in Linux this is set to the total number of cores OR if running in a cgroup (container), the total number of effective cores that can be used as returned by `sched_getaffinity`. |

### Read-Only Parameters

| Name                            | Default                                             | Description                                                                                                                                                                                                                                                                                         |
|---------------------------------|-----------------------------------------------------|---------------------------------------------------------------|
| `last_reload_time`              | `not reloaded`                                      | Contains the timestamp of when the wrapper was last reloaded. |

### Logging Parameters

| Name                    | Default                                      | Description                                                                                                                                                               |
|-------------------------|----------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `log.destination`       | `STDOUT`                                     | Log output destination - valid values are: STDOUT, STDERR, file path                                                                                                      |
| `log.formatter_name`    | `TextFormatter`                              | Log format for output - valid values are: TextFormatter, JSONFormatter                                                                                                    |
| `log.formatter_options` | `full_timestamp=true`, `pad_level_text=true` | Section containing options for the log formatter. Reference [logrus](https://github.com/sirupsen/logrus) for valid values. Both snake case and title case are acceptable. |
| `log.level`             | `INFO`                                       | Log verbosity for output - valid values are: TRACE, DEBUG, INFO, WARN, ERROR, PANIC, FATAL                                                                                |

## Environment Variables

Configuration parameters can be specified as environment variables by 
prefixing the configuration parameter name with `NW_` and converting
the whole string to uppercase. For example, `nginx_binary` becomes
`NW_NGINX_BINARY` and `log.level` becomes `NW_LOG.LEVEL`.

## Configuration File

Configuration parameters can be specfied in any [viper-compatible](https://github.com/spf13/viper)
format. Examples are included for [JSON](../sample_configs/nginx-wrapper-example.json), 
[TOML](../sample_configs/nginx-wrapper-example.toml) and 
[YAML](../sample_configs/nginx-wrapper-example.yml).

## Plugins

### Templating

For configuring templates, see the [templating documentation](templating.md).

### Coprocesses

For configuring coprocesses, see the [coprocess documentation](coprocess.md).