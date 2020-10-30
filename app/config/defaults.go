/*
 *  Copyright 2020 F5 Networks
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package config

import (
	"github.com/nginxinc/nginx-wrapper/lib/api"
	"github.com/nginxinc/nginx-wrapper/lib/fs"
	"os"
	"path/filepath"
	"runtime"
)

/*
Defaults serve three purposes:
 1. They let us know the default value of a config setting.
 2. They work around a shortcoming in viper where config sub settings can't have
    defaults reliably set.
 3. They let us know what keys are used, so that we can iterate all values in an
    ordered fashion.
*/

// DefaultConfigPath is a best guess of a relative path that may contain the config file.
const DefaultConfigPath = "nginx-wrapper.toml"

// CoreDefaults for the normal operation of nginx-wrapper.
var CoreDefaults = map[string]interface{}{
	// contains all environment variables as loaded on start up
	"env": envAsMap(os.Environ()),
	// unique id for the host OS
	"host_id": hostID(defaultHostIDGenerators),
	// contains a timestamp of the last time nginx was reloaded
	"last_reload_time": "not reloaded",
	// path to the nginx modules directory
	"modules_path": "/usr/lib/nginx/modules",
	// path to the nginx binary executable
	"nginx_binary": "nginx",
	// version of nginx binary
	"nginx_version": "unknown",
	// if true, indicates nginx is NGINX+
	"nginx_is_plus": false,
	// path to nginx-wrapper plugins directory
	"plugin_path": "./plugins",
	// list of enabled plugins - none by default
	"enabled_plugins": []string{},
	// path on filesystem to make the nginx run path root
	"run_path": filepath.Clean(fs.TempDirectoryPath("nginx-wrapper")),
	// total number of cores OR if running in a cgroup (container),
	// the total number of effective cores that can be used as returned
	// by sched_getaffinity.
	"vcpu_count": runtime.NumCPU(),
}

// LogDefaults for nginx-wrapper log settings.
var LogDefaults = map[string]interface{}{
	// log level of verbosity
	"level": "INFO",
	// destination to write log output to: STDOUT, STDERR, or file path
	"destination": "STDOUT",
	// logger formatter: TextFormatter, JSONFormatter
	"formatter_name": "TextFormatter",
	// additional options passed to formatter
	"formatter_options": TextFormatterOptionsDefaults,
}

// TextFormatterOptionsDefaults for text formatter log settings.
var TextFormatterOptionsDefaults = map[string]interface{}{
	// if true, TextFormatter displays full timestamp
	"full_timestamp": true,
	// if true, TextFormatter will pad log level indent
	"pad_level_text": true,
}

// PluginDefaults dynamically set defaults by plugins for plugins.
// This is a map of maps with each plugin getting its own map.
var PluginDefaults = map[string]map[string]interface{}{}

// DynamicCoreDefaults are defaults that depend on runtime settings.
func DynamicCoreDefaults(settings api.Settings) map[string]interface{} {
	return map[string]interface{}{
		// path to subdirectory of run path to write conf files to
		"conf_path": settings.GetString("run_path") + fs.PathSeparator + "conf",
	}
}
