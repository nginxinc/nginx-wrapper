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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"sort"
	"strings"
)

// Keys contain all of the known default value keys.
func Keys() []string {
	dynamicDefaults := DynamicCoreDefaults(viper.New())
	var keys []string

	for k := range CoreDefaults {
		keys = append(keys, k)
	}
	for k := range dynamicDefaults {
		keys = append(keys, k)
	}
	for k := range LogDefaults {
		if !strings.Contains(k, ".") {
			keys = append(keys, "log."+k)
		}
	}
	for pluginName, pluginConfig := range PluginDefaults {
		keyPrefix := pluginName + "."

		for k := range pluginConfig {
			keys = append(keys, keyPrefix+k)
		}
	}

	sort.Strings(keys)

	return keys
}

// AllKnownElements contain all known keys and their values from the passed
// Viper configuration instance formatted using the supplied delimiter to
// indicate sub-keys.
func AllKnownElements(settings api.Settings, delimiter string) map[string]interface{} {
	keys := Keys()
	all := make(map[string]interface{}, len(keys))

	addElements(settings, CoreDefaults, all, "")
	addElements(settings, DynamicCoreDefaults(settings), all, "")
	logConfig := SubViperConfig(settings, "log")

	addElements(logConfig, LogDefaults, all, "log"+delimiter)

	for pluginName, pluginDefaults := range PluginDefaults {
		pluginConfig := SubViperConfig(settings, pluginName)
		if pluginConfig == nil {
			pluginConfig = viper.New()
		}

		addElements(pluginConfig, pluginDefaults, all, pluginName+delimiter)
	}

	return all
}

// SubViperConfig returns a new instance of viper.Viper which contains all
// of the values for the given key prefix. This function works around
// deficiencies in Viper's Sub() method.
func SubViperConfig(settings api.Settings, subName string) *viper.Viper {
	prefix := subName + "."
	subConfig := viper.New()

	for _, k := range settings.AllKeys() {
		if strings.HasPrefix(k, prefix) {
			keyWithoutPrefix := k[len(prefix):]
			subConfig.Set(keyWithoutPrefix, settings.Get(k))
		}
	}

	return subConfig
}

func addElements(settings api.Settings, src map[string]interface{}, all map[string]interface{}, outputPrefix string) {
	for k := range src {
		v := settings.Get(k)

		if v == nil || v == "" {
			v = src[k]
		}
		all[outputPrefix+k] = v
	}
}

// totalPluginKeys sums the total number of subkeys within all of the keys
// of pluginDefaults.
func totalPluginKeys(pluginDefaults map[string]map[string]interface{}) int {
	totalKeys := 0

	for _, pluginConfigValue := range pluginDefaults {
		totalKeys += len(pluginConfigValue)
	}

	return totalKeys
}

// envAsMap parsed os.Environ() and reprocesses it into a map.
func envAsMap(env []string) map[string]string {
	envMap := make(map[string]string, len(env))
	numOfKVSegments := 2

	for _, e := range env {
		segments := strings.SplitN(e, "=", 2)

		if len(segments) == numOfKVSegments {
			envMap[segments[0]] = segments[1]
		} else if len(segments) == 1 {
			envMap[segments[0]] = ""
		} else {
			log.Warnf("invalid environment variable: %s", e)
		}
	}

	return envMap
}
