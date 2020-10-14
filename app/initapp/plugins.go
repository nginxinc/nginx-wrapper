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
	"github.com/davecgh/go-spew/spew"
	slog "github.com/go-eden/slf4go"
	"github.com/nginxinc/nginx-wrapper/app/config"
	"github.com/nginxinc/nginx-wrapper/app/embeddedplugins"
	"github.com/nginxinc/nginx-wrapper/lib/api"
	"github.com/nginxinc/nginx-wrapper/lib/osenv"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"
)

func initPlugins(startPlugins bool, vconfig *viper.Viper) {
	initEmbeddedPlugins(startPlugins, vconfig)
	initExternalPlugins(startPlugins, vconfig)
}

func initEmbeddedPlugins(startPlugins bool, vconfig *viper.Viper) {
	embeddedplugins.LoadAll(startPlugins, vconfig)
}

func initExternalPlugins(startPlugins bool, vconfig *viper.Viper) {
	log := slog.NewLogger("init-plugins")
	pluginPath := vconfig.GetString("plugin_path")
	if pluginPath == "" {
		log.Errorf("plugin_path is blank - can't load plugins")
		return
	}

	pathInfo, statErr := os.Stat(pluginPath)
	if statErr != nil {
		log.Errorf("error reading plugin_path (%s) - can't load plugins: %v",
			pluginPath, statErr)
		return
	}

	if !pathInfo.IsDir() {
		log.Error("plugin_path is not a directory - can't load plugins")
		return
	}

	var plugins []string
	walkErr := filepath.Walk(pluginPath, func(path string, info os.FileInfo, err error) error {
		// skip directories
		// skip files without the shared object suffix
		if !info.IsDir() && strings.HasSuffix(path, osenv.SharedObjectSuffix) {
			plugins = append(plugins, path)
		}
		return nil
	})
	if walkErr != nil {
		log.Errorf("unable to traverse plugin_path - can't load plugins: %v",
			walkErr)
		return
	}

	for _, file := range plugins {
		loadPluginErr := loadPlugin(file, startPlugins, vconfig)
		if loadPluginErr != nil {
			if strings.Contains(errors.Cause(loadPluginErr).Error(),
				"plugin was built with a different version of package") {
				log.Warn(loadPluginErr)
			} else {
				log.Errorf("unable to load plugin (%s): %+v", file, loadPluginErr)
			}
		}
	}
}

func loadPlugin(file string, startPlugins bool, vconfig *viper.Viper) error {
	log := slog.NewLogger("load-plugin")

	pluginRef, pluginOpenErr := plugin.Open(file)
	if pluginOpenErr != nil {
		return errors.Wrap(pluginOpenErr, "unable to open plugin file")
	}

	metadata, loadMetadataErr := loadPluginMetadata(pluginRef, vconfig)
	if loadMetadataErr != nil {
		return errors.Wrap(loadMetadataErr, "unable to load plugin metadata")
	}

	start, loadStartFuncErr := loadPluginStartFunc(pluginRef)
	if loadStartFuncErr != nil {
		return errors.Wrap(loadStartFuncErr, "unable to load plugin Start func")
	}

	pluginName := metadata["name"].(string)

	if !isPluginEnabled(pluginName, vconfig) {
		log.Debugf("plugin [%s] was detected but not enabled - not loading", pluginName)
		return nil
	}

	configDefaults := metadata["config_defaults"].(map[string]interface{})

	// Add the defaults read from plugin so that we know what are the keys
	// used in a plugin
	config.PluginDefaults[pluginName] = configDefaults

	// Set the defaults as read from the plugin
	for k, v := range configDefaults {
		configKey := pluginName + "." + k
		vconfig.SetDefault(configKey, v)
	}

	if startPlugins {
		startErr := startPlugin(start, pluginName, vconfig)
		if startErr != nil {
			return startErr
		}
		log.Infof("started plugin: [%s]", pluginName)
	} else {
		log.Infof("loaded plugin: [%s]", pluginName)
	}

	return nil
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

func startPlugin(startFunc func(api.PluginStartupContext) error, pluginName string,
	vconfig api.Settings) error {

	context := api.PluginStartupContext{
		Settings: vconfig,
	}

	// It is the plugin's decision if they want to kick off
	// another go worker. However, the plugin must return
	// from the Start() function.
	startErr := startFunc(context)
	if startErr != nil {
		return errors.Wrapf(startErr, "error starting plugin (%s)", pluginName)
	}

	return nil
}

func loadPluginMetadata(pluginRef *plugin.Plugin, settings api.Settings) (map[string]interface{}, error) {
	pluginSymbol, lookupErr := pluginRef.Lookup("Metadata")
	if lookupErr != nil {
		return nil, errors.Wrapf(lookupErr, "unable to load plugin metadata (%v)", pluginRef)
	}

	metadata, ok := pluginSymbol.(func(settings api.Settings) map[string]interface{})
	if !ok {
		return nil, errors.Errorf("unable to load plugin metadata from symbol (%v).Metadata()", pluginRef)
	}

	m := metadata(settings)

	if m["name"] == nil {
		return nil, errors.WithStack(&PluginMetadataKeyMissing{KeyMissing: "name"})
	}
	if reflect.ValueOf(m["name"]).Kind() != reflect.String {
		return nil, errors.WithStack(&PluginMetadataMalformedType{
			KeyName:      "name",
			ExpectedType: reflect.String,
		})
	}
	if m["config_defaults"] == nil {
		m["config_defaults"] = map[string]interface{}{}
	}
	if reflect.ValueOf(m["config_defaults"]).Kind() != reflect.Map {
		return nil, errors.WithStack(&PluginMetadataMalformedType{
			KeyName:      "config_defaults",
			ExpectedType: reflect.Map,
		})
	}

	return m, nil
}

func loadPluginStartFunc(pluginRef *plugin.Plugin) (func(api.PluginStartupContext) error, error) {
	pluginSymbol, lookupErr := pluginRef.Lookup("Start")
	if lookupErr != nil {
		return nil, errors.Wrapf(lookupErr, "unable to load plugin start function (%v)",
			pluginRef)
	}

	start, ok := pluginSymbol.(func(api.PluginStartupContext) error)
	if !ok {
		return nil, errors.Errorf("unexpected type (%v) from start module symbol",
			spew.Sdump(start))
	}

	return start, nil
}
