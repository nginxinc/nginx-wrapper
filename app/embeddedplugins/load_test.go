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
	"github.com/nginxinc/nginx-wrapper/lib/api"
	"github.com/spf13/viper"
	"testing"
)

func TestIsPluginEnabledTrue(t *testing.T) {
	vconfig := viper.New()
	vconfig.Set("enabled_plugins", []string{"penguin", "bobcat"})

	enabled := isPluginEnabled("bobcat", vconfig)
	if !enabled {
		t.Error("plugin should be marked as enabled")
	}
}

func TestIsPluginEnabledFalse(t *testing.T) {
	vconfig := viper.New()
	vconfig.Set("enabled_plugins", []string{"penguin", "bobcat"})

	enabled := isPluginEnabled("wolf", vconfig)
	if enabled {
		t.Error("plugin should not be marked as enabled")
	}
}

func TestLoadPlugin(t *testing.T) {
	vconfig := viper.New()
	vconfig.Set("config_value", "was set")
	started := false
	metadata := map[string]interface{}{"this": "that"}

	metadataFunc := func(settings api.Settings) map[string]interface{} {
		metadata["config_value"] = settings.Get("config_value")
		return metadata
	}

	startFunc := func(context api.PluginStartupContext) error {
		started = true
		return nil
	}

	plugin := embeddedPlugin{
		name:      "test-plugin",
		metadata:  metadataFunc,
		startFunc: startFunc,
	}

	err := LoadPlugin(false, plugin, vconfig)
	if err != nil {
		t.Error(err)
	}

	if metadata["config_value"] != "was set" {
		t.Error("plugin was not loaded")
	}

	if started {
		t.Error("plugin start function was invoked and not expected")
	}
}

func TestLoadAndStartPlugin(t *testing.T) {
	vconfig := viper.New()
	vconfig.Set("config_value", "was set")
	started := false
	metadata := map[string]interface{}{"this": "that"}

	metadataFunc := func(settings api.Settings) map[string]interface{} {
		metadata["config_value"] = settings.Get("config_value")
		return metadata
	}

	startFunc := func(context api.PluginStartupContext) error {
		started = true
		return nil
	}

	plugin := embeddedPlugin{
		name:      "test-plugin",
		metadata:  metadataFunc,
		startFunc: startFunc,
	}

	err := LoadPlugin(true, plugin, vconfig)
	if err != nil {
		t.Error(err)
	}

	if metadata["config_value"] != "was set" {
		t.Error("plugin was not loaded")
	}

	if !started {
		t.Error("plugin start function was not invoked")
	}
}
