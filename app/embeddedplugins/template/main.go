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

package template

import (
	slog "github.com/go-eden/slf4go"
	"github.com/nginxinc/nginx-wrapper/lib/api"
	"github.com/nginxinc/nginx-wrapper/lib/events"
	"github.com/nginxinc/nginx-wrapper/lib/fs"
	"github.com/pkg/errors"
	"os"
	"strings"
	"time"
)

// PluginName contains the name/id of the plugin.
const PluginName string = "template"

var log = slog.NewLogger("template")

var templateDefaults = map[string]interface{}{
	// if true, deletes all files within the run path upon exit
	"delete_run_path_on_exit": false,
	// if true, deletes only the template files upon exit
	"delete_templated_conf_on_exit": true,
	// file extension for all nginx conf files that will be templated
	"template_suffix": ".tmpl",
	// template substitution left delimiter
	"template_var_left_delim": "[[",
	// template substitution right delimiter
	"template_var_right_delim": "]]",
	// subdirectories within the run path to create on startup
	"run_path_subdirs": []string{"client_body", "conf", "proxy", "fastcgi", "uswsgi", "scgi"},
}

// Metadata passed to nginx-wrapper initialization.
func Metadata(settings api.Settings) map[string]interface{} {
	var templateSuffix string
	// Override the default with the provided value for template suffix because it is
	// needed to create the default template.conf_template_path value.
	if settings.GetString(PluginName+".template_suffix") != "" {
		templateSuffix = settings.GetString(PluginName + ".template_suffix")
	} else {
		templateSuffix = templateDefaults["template_suffix"].(string)
	}

	// path to the template nginx config file(s)
	templateDefaults["conf_output_path"] = settings.Get("conf_path")
	// path to the nginx config template directory or file
	templateDefaults["conf_template_path"] = "./nginx.conf" + templateSuffix

	return map[string]interface{}{
		"name":            PluginName,
		"config_defaults": templateDefaults,
	}
}

// Start is invoked by nginx-wrapper to start the plugin.
func Start(context api.PluginStartupContext) error {
	settings := context.Settings
	log.Tracef("plugin [%s] starting", PluginName)

	initTemplatingEvents(settings)

	return nil
}

func initTemplatingEvents(settings api.Settings) {
	log := slog.NewLogger("template-event")

	confTemplate := NewTemplate()

	// Creates the directory structure for nginx at the run_path specified in the config
	initRunPathTrigger := events.Trigger{
		Name: PluginName + ".init-runpath",
		Function: func(message events.Message) error {
			runPath := settings.GetString("run_path")
			log.Tracef("creating directory/verifying: %s", runPath)
			mkdirRunPathErr := os.MkdirAll(runPath, os.ModePerm)
			if mkdirRunPathErr != nil {
				return errors.Wrapf(mkdirRunPathErr,
					"error creating run_path directory (%s): %v",
					runPath, mkdirRunPathErr)
			}

			mksubDirErr := makeRunPathSubDirs(settings)
			if mksubDirErr != nil {
				return errors.Wrapf(mksubDirErr,
					"error creating subdirectory for run_path (%s): %v",
					runPath, mksubDirErr)
			}

			return nil
		},
	}

	// Runs the templating engine and copies the output such that nginx can load it
	applyTemplatesTrigger := events.Trigger{
		Name: PluginName + ".apply-templates",
		Function: func(message events.Message) error {
			settings.Set("last_reload_time", time.Now().UTC().String())

			discoverErr := confTemplate.DiscoverTemplateFiles(settings)
			if discoverErr != nil {
				return errors.Wrap(discoverErr, "error traversing nginx conf template files")
			}

			templatingErr := confTemplate.ApplyTemplating(settings)
			if templatingErr != nil {
				// Exit early without a stacktrace for templating related problems so
				// that users can better troubleshoot problems with templates
				if templatingErr.IsATemplatingProblem {
					log.Fatal(templatingErr)
				}

				return templatingErr
			}

			return nil
		},
	}

	// Cleans up the template files or the nginx directory
	cleanUpTrigger := events.Trigger{
		Name: PluginName + ".clean-up-templates",
		Function: func(message events.Message) error {
			if settings.GetBool("delete_templated_conf_on_exit") {
				confTemplate.CleanOutputConfiguration()
			}

			if settings.GetBool(PluginName + ".delete_run_path_on_exit") {
				runPath := settings.GetString("run_path")
				if runPath != "" {
					log.Tracef("removing nginx working directory (%s)", runPath)
					removeErr := os.RemoveAll(runPath)
					if removeErr != nil {
						log.Errorf("unable to remove run_path (%s): %v",
							runPath, removeErr)
					}
				}
			}

			return nil
		},
	}

	events.GlobalEvents.NginxPreStart.AddTrigger(&initRunPathTrigger)
	events.GlobalEvents.NginxPreStart.AddFinalTrigger(&applyTemplatesTrigger)
	events.GlobalEvents.NginxPreReload.AddFinalTrigger(&applyTemplatesTrigger)
	events.GlobalEvents.NginxExit.AddFinalTrigger(&cleanUpTrigger)
}

func makeRunPathSubDirs(settings api.Settings) error {
	paths := settings.GetStringSlice(PluginName + ".run_path_subdirs")

	builder := strings.Builder{}

	for i, path := range paths {
		subdir := settings.GetString("run_path") + fs.PathSeparator +
			path
		mkdirSubdirErr := os.MkdirAll(subdir, os.ModePerm)
		if mkdirSubdirErr != nil {
			return errors.Wrapf(mkdirSubdirErr, "error creating subdirectory (%s)",
				subdir)
		}

		if log.IsTraceEnabled() {
			builder.WriteString(subdir)
			if i < len(paths)-1 {
				builder.WriteString(" ")
			}
		}
	}

	if log.IsTraceEnabled() {
		log.Tracef("creating/verifying directories: %s", builder.String())
	}

	return nil
}
