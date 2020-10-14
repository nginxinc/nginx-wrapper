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

package embeddedplugins

import (
	"fmt"
	slog "github.com/go-eden/slf4go"
	"github.com/nginxinc/nginx-wrapper/app/config"
	"github.com/nginxinc/nginx-wrapper/app/embeddedplugins/coprocess"
	"github.com/nginxinc/nginx-wrapper/app/embeddedplugins/template"
	"github.com/nginxinc/nginx-wrapper/lib/api"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"reflect"
)

type embeddedPlugin struct {
	name      string
	metadata  func(settings api.Settings) map[string]interface{}
	startFunc func(api.PluginStartupContext) error
}

var embeddedPlugins = []embeddedPlugin{
	{name: coprocess.PluginName, metadata: coprocess.Metadata, startFunc: coprocess.Start},
	{name: template.PluginName, metadata: template.Metadata, startFunc: template.Start},
}

// pluginMetadataMalformedType indicates that the plugin metadata was passed the wrong data type.
type pluginMetadataMalformedType struct {
	KeyName      string
	ExpectedType reflect.Kind
	Err          error
}

func (p *pluginMetadataMalformedType) Error() string {
	return fmt.Sprintf("plugin metadata with wrong data type for value with key (%s)", p.KeyName)
}

// LoadAll loads all embedded plugins. The list of plugins to be loaded is hardcoded
// in the variable embeddedPlugins. The parameter startPlugins is a flag indicating
// if the embedded plugins should be started.
func LoadAll(startPlugins bool, vconfig *viper.Viper) {
	log := slog.NewLogger("init-embedded-plugins")

	for _, embeddedPlugin := range embeddedPlugins {
		if isPluginEnabled(embeddedPlugin.name, vconfig) {
			err := LoadPlugin(startPlugins, embeddedPlugin, vconfig)
			if err != nil {
				log.Errorf("error loading embedded plugin (%s): %v",
					embeddedPlugin.name, err)
			}
		} else {
			log.Debugf("plugin [%s] was detected but not enabled - not loading", embeddedPlugin.name)
		}
	}
}

func isPluginEnabled(pluginName string, vconfig *viper.Viper) bool {
	enabledPlugins := vconfig.GetStringSlice("enabled_plugins")

	found := false
	for _, p := range enabledPlugins {
		if p == pluginName {
			found = true
			break
		}
	}

	return found
}

// LoadPlugin will load a plugins settings into memory and start it if the
// startPlugin flag is true.
func LoadPlugin(startPlugin bool, embeddedPlugin embeddedPlugin, settings api.Settings) error {
	log := slog.NewLogger("load-plugin")

	m := embeddedPlugin.metadata(settings)
	if m["config_defaults"] == nil {
		m["config_defaults"] = map[string]interface{}{}
	}

	if reflect.ValueOf(m["config_defaults"]).Kind() != reflect.Map {
		return errors.WithStack(&pluginMetadataMalformedType{
			KeyName:      "config_defaults",
			ExpectedType: reflect.Map,
		})
	}

	configDefaults := m["config_defaults"].(map[string]interface{})

	config.PluginDefaults[embeddedPlugin.name] = configDefaults

	for k, v := range configDefaults {
		configKey := embeddedPlugin.name + "." + k
		settings.SetDefault(configKey, v)
	}

	startFunc := embeddedPlugin.startFunc
	startupContext := api.PluginStartupContext{Settings: settings}

	if startPlugin {
		startErr := startFunc(startupContext)
		if startErr != nil {
			return startErr
		}
		log.Infof("started plugin: [%s]", embeddedPlugin.name)
	} else {
		log.Infof("loaded plugin: [%s]", embeddedPlugin.name)
	}

	return nil
}
