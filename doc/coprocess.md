# Coprocess Guide

The coprocess configuration block is specified as one block per coprocess. 
With the first sub-key indicating the name of the coprocess as it is identified
within the configuration system (it doesn't change anything). The order of 
execution of coprocesses is the same as the order in which they appear in the 
configuration file.

## Configuration Parameters

| Name                                     | Notes / Defaults                                    | Description                                                                                                                                                                                                                                                                   |
|------------------------------------------|-----------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `coprocess.<name>.name`                  | <no default - must not be blank> (Required)                                                    | Name of the coprocess used to identify it within the wrapper.                                                                                                                                                                      |                                                 
| `coprocess.<name>.exec`                  | <no default - must not be blank> (Required)                                                    | Command to execute as the coprocess. Parameters are separated by spaces like a CLI invocation. Interpolated settings specified in the format of `${variable}` can be used. See table below for valid settings.                     |
| `coprocess.<name>.stop_exec`             | <no default - must not be blank>                                                               | Command to execute after process finishes. Parameters are separated by spaces like a CLI invocation. Interpolated settings specified in the format of `${variable}` can be used. See table below for valid settings.               |
| `coprocess.<name>.user`                  | <no default - must not be blank>                                                               | User to run command as. The wrapper must be run with sufficient permissions that it can change to the user.                                                                                                                        |
| `coprocess.<name>.restarts`              | `never`, `unlimited`, or integer value (Required)                                              | Number of times to restart the process after it exits. `never` implies that it won't restart and `unlimited` means that it will continually restart until the wrapper exits.                                                       |
| `coprocess.<name>.time_between_restarts` | <`0s`> - duration value of a [number plus a time unit](https://pkg.go.dev/time#ParseDuration)  | Duration to wait between each restart of the process. If restarts are set to `never` this value is ignored.                                                                                                                        |
| `coprocess.<name>.background`            | `false`                                                                                        | Flag indicating if the coprocess will be run in the background. If run in the background, the wrapper can continue execution without the process exiting. If not, the wrapper will wait for the process to exit before continuing. |
| `coprocess.<name>.exec_event`            | <event name> (Required)                                                                        | Name of the wrapper event that will trigger running `coprocess.<name>.exec`. Valid values are listed in the [events documentation](events.md).                                                                                     |
| `coprocess.<name>.stop_event`            | <event name>                                                                                   | Name of the wrapper event that will trigger running `coprocess.<name>.stop_exec`. Valid values are listed in the [events documentation](events.md).                                                                                |

## Interpolated Settings

| Name                                     | Setting                                            |
|------------------------------------------|----------------------------------------------------|
| `host_id`                                | Same value as specified in core configuration.     |
| `modules_path`                           | Same value as specified in core configuration.     |
| `nginx_binary`                           | Same value as specified in core configuration.     |
| `plugin_path`                            | Same value as specified in core configuration.     |
| `run_path`                               | Same value as specified in core configuration.     |
| `vcpu_count`                             | Same value as specified in core configuration.     |
| `last_reload_time`                       | Same value as specified in core configuration.     |
| `wrapper_pid`                            | Process id of the running instance of the wrapper. |

## Examples

In the following example, we have two coprocesses. The first coprocess runs once and
modifies a Consul configuration file using `sed`. This coprocess blocks execution of the
wrapper until it completes. The next coprocess runs consul in the background.

```toml
# coprocess.consul-service-config is the config id, this doesn't have to be the same as name below, but
# by convention it typically is. 
[coprocess.consul-service-config]
    # Name of the coprocess used in logging
    name = "consul-service-config"
    # Executes sed so that it replaces the string __host-id__ with the host id stored in the wrapper
    exec = [ "sed", "-i", "s/__host-id__/${host_id}/g", "/opt/consul/conf.d/nginx.json" ]
    # Executes consul and tells consul to deregister when the wrapper is exiting
    stop_exec = [ "consul", "services", "deregister", "-id=edge-nginx-${host_id}" ]
    # Commands are run as the user 'consul' 
    user = "consul"
    # The exec command is never run again after it exists
    restarts = "never"
    # The exec command is run in the foreground blocking the NGINX Wrapper process for continuing
    background = "false"
    # The exec command will be run when the NGINX Wrapper fires the pre-start event
    exec_event = "pre-start"
    # The stop_exec command will be run when the NGINX Wrapper fires the exit event
    stop_event = "exit"

[coprocess.consul]
    # Name of the coprocess used in logging
    name = "consul"
    # Command to run as coprocess
    exec = [
        "consul",
        "agent",
        "-node-meta=host_id:${host_id}",
        "-node-meta=hostname:${HOSTNAME}",
        "-node-id=${host_id}",
        "-node=edge-${host_id}",
        "-client", "127.0.0.1",
        "-join", "172.21.0.1",
        "-data-dir", "/var/opt/consul",
        "-config-dir", "/opt/consul/conf.d",
        "-log-level=info"
    ]
    # Leaves the consul cluster when NGINX Wrapper exits
    stop_exec = [ "consul", "leave" ]
    # User to setuid to when running coprocess (nginx-wrapper must be run as root)
    user = "consul"
    # Restart policy for when coprocess exits: never, unlimited, or integer value indicate times to restart
    restarts = "unlimited"
    # Flag indicating if the coprocess should be backgrounded, if false nginx-wrapper will wait for coprocess to exit
    background = true
    # nginx-wrapper event that will trigger start up of coprocess
    exec_event = "pre-start"
    # nginx-wrapper event that will trigger termination of coprocess
    stop_event = "exit"
``` 