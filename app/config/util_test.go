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
	"github.com/nginxinc/nginx-wrapper/lib/osenv"
	"reflect"
	"testing"
)

func TestEnvAsMap(t *testing.T) {
	env := []string{
		"PATH=/usr/bin:/usr/local/bin",
		"WEIRD_PATH=/var/doc/ff00=xx3",
		"NO_EQUALS",
		"DUPLICATE=1",
		"",
		"DUPLICATE=2",
	}

	actual := envAsMap(env)
	expected := map[string]string{
		"PATH":       "/usr/bin:/usr/local/bin",
		"WEIRD_PATH": "/var/doc/ff00=xx3",
		"NO_EQUALS":  "",
		"DUPLICATE":  "2",
	}

	if reflect.DeepEqual(actual, expected) {
		t.Errorf(osenv.LineBreak+"actual:   %v"+
			osenv.LineBreak+"expected: %v", actual, expected)
	}
}

func TestPluginDefaultsKeysAddProperly(t *testing.T) {
	pluginDefaults := map[string]map[string]interface{}{
		"plugin_1": {
			"key_1": "value",
			"key_2": false,
		},
		"plugin_2": {
			"key_1": "value",
			"key_2": "value",
			"key_3": 4,
		},
	}
	expected := 5 // there are five subkeys
	actual := totalPluginKeys(pluginDefaults)

	if actual != expected {
		t.Errorf("expected (%d) keys but found (%d)", expected, actual)
	}
}
