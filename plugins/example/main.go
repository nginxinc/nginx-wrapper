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

package main

import (
	slog "github.com/go-eden/slf4go"
	"github.com/nginxinc/nginx-wrapper/lib/api"
	"github.com/nginxinc/nginx-wrapper/lib/events"
	"syscall"
	"time"
)

// PluginName contains the name/id of the plugin.
const PluginName string = "example"
const sleepSeconds = 3

// nolint:unused
var log = slog.NewLogger("example")

// Metadata is a required function that is read by nginx-wrapper
// in order to discover the available and default configuration
// properties used by the plugin.
// nolint:deadcode,unparam,unused,gomnd
// noinspection GoUnusedExportedFunction .
func Metadata(_ api.Settings) map[string]interface{} {
	return map[string]interface{}{
		"name": PluginName,
		"config_defaults": map[string]interface{}{
			"example_key_1": "one",
			"example_key_2": 2,
		},
	}
}

// Start is invoked by nginx-wrapper to start the plugin.
// nolint:deadcode,unparam,unused
// noinspection GoUnusedExportedFunction .
func Start(context api.PluginStartupContext) error {
	log.Tracef("plugin [%s] starting", PluginName)

	settings := context.Settings

	// Setting can be overridden within the plugin and their values will
	// propagate within the main application
	settings.Set("example.example_key_2", 9)

	// When accessing settings for the plugin, you access them by pulling in
	// the settings passed to Start().
	log.Debugf("Example key 1: %v", settings.Get("example.example_key_1"))
	log.Debugf("Example key 2: %v", settings.Get("example.example_key_2"))

	// If you want to hook into events that nginx-wrapper is triggering,
	// then you can do it by adding a trigger.
	events.GlobalEvents.NginxReload.AddTrigger(&events.Trigger{
		Name:     PluginName + ".on-reload",
		Function: onReload,
	})
	events.GlobalEvents.NginxExit.AddTrigger(&events.Trigger{
		Name:     PluginName + ".on-shutdown",
		Function: onShutdown,
	})

	// For any operations that you want to be running continually in
	// the background, you will need to set them up here. If you block
	// within the Start() function, nginx-wrapper will fail to start.

	// Example of reloading nginx from within a plugin
	go func() {
		time.Sleep(sleepSeconds * time.Second)
		// SIGUSR2 will reload ONLY nginx
		api.Signals <- syscall.SIGUSR2
		// SIGHUP will reload nginx-wrapper in addition to nginx
	}()

	return nil
}

// nolint:unused
func onReload(events.Message) error {
	println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	println("!!!!!!!!!!! EXAMPLE PLUGIN !!!!!!!!!!!")
	println("!!!!!!!!!!! NGINX RELOADED !!!!!!!!!!!")
	println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")

	return nil
}

// nolint:unused
func onShutdown(events.Message) error {
	println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	println("!!!!!!!!!!! EXAMPLE PLUGIN !!!!!!!!!!!")
	println("!!!!!!!!!!!  NGINX EXITED  !!!!!!!!!!!")
	println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")

	return nil
}
