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
	"github.com/nginxinc/nginx-wrapper/lib/events"
	"github.com/spf13/viper"
)

// Init runs all of the needed functions to load the application to its
// pre-running state.
//
// startPlugins is a flag that indicates if we should run the Start() function
// on the plugins found. By default it is off so that plugin metadata can be
// loaded without starting the plugin. If the run command is executed, this
// flag should be set to true.
// configPath is the path to the configuration file or directory.
func Init(startPlugins bool, configPath string) *viper.Viper {
	vconfig := initConfig(configPath)
	initLog(vconfig)
	events.Init()
	initEventListenerLog()
	initPlugins(startPlugins, vconfig)
	logEvents()
	return vconfig
}
