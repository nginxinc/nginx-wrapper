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

package initapp

import (
	"fmt"
	"github.com/nginxinc/nginx-wrapper/app/config"
	"github.com/nginxinc/nginx-wrapper/app/nginx/version"
	"github.com/nginxinc/nginx-wrapper/lib/fs"
	"github.com/nginxinc/nginx-wrapper/lib/osenv"
	"github.com/spf13/viper"
	"os"
	"strings"
)

// EnvPrefix is the prefix for configuration parameters passed by environment
// variables. This prefix is followed by an underscore. For example nginx_binary
// becomes NW_NGINX_BINARY.
const EnvPrefix = "NW"

func initConfig(configPath string) *viper.Viper {
	// Load configuration
	vconfig := viper.New()
	vconfig.SetTypeByDefaultValue(true)
	vconfig.SetConfigFile(configPath)
	vconfig.AutomaticEnv()
	vconfig.SetEnvPrefix(EnvPrefix)

	for k, v := range config.CoreDefaults {
		vconfig.SetDefault(k, v)
	}

	readErr := vconfig.ReadInConfig()

	for k, v := range config.DynamicCoreDefaults(vconfig) {
		vconfig.SetDefault(k, v)
	}

	applyLogEnvVars(vconfig)

	if readErr != nil {
		var errMsg string

		if os.IsNotExist(readErr) && configPath == config.DefaultConfigPath {
			errMsg = fmt.Sprintf("The default config file path (%s) was not found in the current directory%s",
				configPath, osenv.LineBreak)
		} else if os.IsNotExist(readErr) {
			errMsg = fmt.Sprintf("The specified config file path (%s) does not exist%s",
				configPath, osenv.LineBreak)
		} else if os.IsPermission(readErr) {
			errMsg = fmt.Sprintf("The specified config file path (%s) does not sufficent privledges: %v%s",
				configPath, readErr, osenv.LineBreak)
		} else {
			errMsg = fmt.Sprintf("%v%s", readErr, osenv.LineBreak)
		}

		_, _ = fmt.Fprintf(os.Stderr, "%v%s", errMsg, osenv.LineBreak)
		os.Exit(1)
	}

	initNginxBinPath(vconfig)
	initVersion(vconfig)

	return vconfig
}

func initNginxBinPath(vconfig *viper.Viper) {
	// Find NGINX binary
	nginxBinPath, findInPathErr := fs.FindInPath(vconfig.GetString("nginx_binary"))
	if findInPathErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v%s", findInPathErr, osenv.LineBreak)
		os.Exit(1)
	}

	vconfig.Set("nginx_binary", nginxBinPath)
}

func initVersion(vconfig *viper.Viper) {
	// Read NGINX version and build information
	nginxBinPath := vconfig.GetString("nginx_binary")
	nginxVersion, versionErr := version.ReadNginxVersion(nginxBinPath)
	if versionErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v%s", versionErr, osenv.LineBreak)
		os.Exit(1)
	}

	vconfig.Set("nginx_version", nginxVersion.Version)
	vconfig.Set("nginx_is_plus", nginxVersion.IsPlus)

	// Set modules default path if we can read it from the NGINX config
	detectedModulesPath, found := nginxVersion.ConfigureArgs.Get("modules-path")
	if found {
		// Update the default values based on the path discovered
		config.CoreDefaults["modules_path"] = detectedModulesPath
		vconfig.SetDefault("modules_path", detectedModulesPath)
	}
}

func applyLogEnvVars(vconfig *viper.Viper) {
	for configKey := range config.LogDefaults {
		envKey := fmt.Sprintf("%s_LOG.%s",
			EnvPrefix, strings.ToUpper(configKey))
		applyEnvToConfig(vconfig, "log."+configKey, envKey)
	}
}

func applyEnvToConfig(vconfig *viper.Viper, configKey string, envKey string) {
	envValue, ok := os.LookupEnv(envKey)
	if ok {
		vconfig.Set(configKey, envValue)
	}
}
