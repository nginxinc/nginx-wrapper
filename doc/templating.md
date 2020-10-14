# Templating Guide

## Configuration Parameters

| Name                                     | Default                                             | Description                                                                                                                                                                                                                                                                                         |
|------------------------------------------|-----------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `template.conf_output_path`              | `{run_path}/conf`                                   | Path to copy the templated contents of `conf_template_path`                                                                                                                                     |
| `template.conf_template_path`            | `./nginx.conf.tmpl`                                 | Path to a file or a directory containing the nginx.conf file and/or all supporting configuration files. This path will have templates applied to it if files have the matching template suffix. |
| `template.delete_run_path_on_exit`       | false                                               | Flag indicating if we want to delete the nginx run directory when the wrapper exits                                                                                                             |
| `template.delete_templated_conf_on_exit` | true                                                | Flag indicating if we want to delete the templated configuration output when the wrapper exits                                                                                                  |
| `template.run_path_subdirs`              | `client_body`, `proxy`, `fastcgi`, `uswsgi`, `scgi` | Subdirectories to create under the run_path                                                                                                                                                     |
| `template.template_suffix`               | `.tmpl`                                             | Suffix for files that will have templating applied                                                                                                                                              |
| `template.template_var_left_delim`       | `[[`                                                | Left substitution characters used in templating. By default NGINX Wrapper uses two square brackets surrounding the templating directive. This differs from the default for go templates.        |
| `template.template_var_right_delim`      | `]]`                                                | Right substitution characters used in templating. By default NGINX Wrapper uses two square brackets surrounding the templating directive. This differs from the default for go templates.       |

## Templating Modes

NGINX Wrapper will apply templating in one of three ways based on the setting
of the `template.conf_template_path` configuration parameter.

1. If `template.conf_template_path` specifies a *file without* a matching `template.template_suffix`,
   it will copy the file as is with no templating applied to 
   `template.conf_output_path`.
2. If `conf_template_path` specifies a *file with* a matching `template.template_suffix`,
   it will process that file as a template and write the contents of it to 
   `conf_output_path/nginx.conf`.
3. If `conf_template_path` specifies a *directory*, it will recursively process
   the contents of that directory looking for files with a matching 
   `template.template_suffix`. If the suffix matches, templating will be applied. If the
   suffix does not match, the file will be copied as is. The same operation will
   be applied to all subdirectories.
   
## Templating

NGINX Wrapper uses [GO Template](https://golang.org/pkg/text/template/) as the 
template engine with the substitution characters reconfigured to `[[` and `]]`
instead of `{{` and `}}`. This is done to allow for better compatibility with
nginx.conf syntax highlighting. The settings can be modified with the parameters
`template_var_left_delim` and `template_var_right_delim`.

To see runtime value of the settings that will be applied to the templating 
engine, run: 
```
./nginx-wrapper debug --config <config path>`
DEBUG  [2020-05-26T15:28:47-07:00] load-plugin: plugin [coprocess] was detected but not enabled - not loading 
INFO   [2020-05-26T15:28:47-07:00] load-plugin: loaded plugin: [example]
                             conf_path: /tmp/nginx-wrapper/conf
                       enabled_plugins: [template example]
                                   env: map[PATH:/usr/bin:HOME:/home/username]
                 example.example_key_1: five
                 example.example_key_2: 2
                               host_id: 1ac762a9-2ee6-48e2-aeb9-3ab4ce05fe85
                      last_reload_time: not reloaded
                       log.destination: STDOUT
                    log.formatter_name: TextFormatter
                 log.formatter_options: map[full_timestamp:true pad_level_text:true]
                             log.level: debug
                          modules_path: /usr/lib/nginx/modules
                          nginx_binary: /usr/sbin/nginx
                         nginx_is_plus: true
                         nginx_version: 1.19.0
                           plugin_path: ./plugins
                              run_path: /tmp/nginx-wrapper
             template.conf_output_path: /tmp/nginx-wrapper/conf
           template.conf_template_path: ./nginx.conf.tmpl
      template.delete_run_path_on_exit: false
template.delete_templated_conf_on_exit: true
             template.run_path_subdirs: [client_body conf proxy fastcgi uswsgi scgi]
              template.template_suffix: .tmpl
      template.template_var_left_delim: [[
     template.template_var_right_delim: ]]
                            vcpu_count: 3
```

The above parameters can be applied to your templates.

In order to access sub-elements of the configuration such as `log.*`, you will
need to use the underscore (`_`) character instead of the dot (`.`) character
because GO template has is very picky about how the dot character is used. For
example, the parameter `log.level` will become `log_level`.

See the [sample_configs directory](../sample_configs) for a number of sample
template files.

## Common Patterns

Below are some examples of how templating can be used to apply some common
configuration patterns.

### Accessing a single value

The parameter is enclosed in `[[` and `]]` brackets and has a leading dot. The
below example shows how `run_path` and `vcpu_count` is templated:
```
daemon                off;
master_process        on;
pid                   [[.run_path]]/nginx.pid;
error_log             /dev/stdout info;
worker_processes      [[.vcpu_count]];
```

### Conditionally applying templating

In the following examples, the string `plugin` can be replaced with the name of
the plugin that you are using.

```
[[if .plugin_uses_njs_module]]
load_module [[.modules_path]]/ngx_http_js_module.so;
[[end]]
```

```
[[if (eq .plugin_log_mode "with_req_id") ]]
log_format main '$remote_addr "$request" '
                '$status "$http_user_agent" '
                '"$http_x_forwarded_for" $req_id';
[[else]]
log_format main '$remote_addr "$request" '
                '$status "$http_user_agent" '
                '"$http_x_forwarded_for"';
[[end]]
```

### Loops

```
[[range .plugin_upstreams]]
[[if ne (len .Nodes) 0]]
    # backends for the [[.Service]] service
    upstream [[.InternalId]]_backend {
[[range .Nodes]]
        server [[.Host]]:[[.Port]];
[[end]]
        keepalive 32;
    }
[[end]]
```