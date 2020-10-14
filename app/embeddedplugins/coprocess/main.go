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

package coprocess

import (
	"fmt"
	slog "github.com/go-eden/slf4go"
	"github.com/nginxinc/nginx-wrapper/lib/api"
	"github.com/nginxinc/nginx-wrapper/lib/events"
	"github.com/pkg/errors"
)

// PluginName contains the name/id of the plugin.
const PluginName string = "coprocess"

var log = slog.NewLogger("coprocess")

// Metadata passed to nginx-wrapper initialization.
func Metadata(_ api.Settings) map[string]interface{} {
	return map[string]interface{}{
		"name":            PluginName,
		"config_defaults": map[string]interface{}{},
	}
}

// Start is invoked by nginx-wrapper to start the plugin.
func Start(context api.PluginStartupContext) error {
	settings := context.Settings
	log.Tracef("plugin [%s] starting", PluginName)

	coprocesses, err := readCoprocesses(settings)
	if err != nil {
		return err
	}

	for _, cp := range coprocesses {
		err := registerCoprocessTriggers(cp)
		if err != nil {
			log.Error(err)
			continue
		}
	}

	return nil
}

func registerCoprocessTriggers(cp Coprocess) error {
	execTrigger := events.Trigger{
		Name: fmt.Sprintf("%s.start-coprocess-%s", PluginName, cp.Name),
		Function: func(message events.Message) error {
			if cp.Background {
				go func() {
					defer cp.Done.SetToIf(false, true)

					err := cp.ExecuteCoprocess()
					if err != nil {
						log.Errorf("%+v", err)
					}
				}()

			} else {
				err := cp.ExecuteCoprocess()
				cp.Done.SetToIf(false, true)
				if err != nil {
					return errors.WithStack(err)
				}
			}

			return nil
		},
	}

	preStopTrigger := events.Trigger{
		Name: fmt.Sprintf("%s.exec-coprocess-stop-cmd-%s", PluginName, cp.Name),
		Function: func(message events.Message) error {
			err := cp.ExecuteStopCmd()
			return errors.Wrapf(err, "error issuing stop command for coprocess (%s)",
				cp.Name)
		},
	}

	stopTrigger := events.Trigger{
		Name: fmt.Sprintf("%s.terminate-coprocess-%s", PluginName, cp.Name),
		Function: func(message events.Message) error {
			err := cp.TerminateCoprocess()
			return errors.WithStack(err)
		},
	}

	log.Tracef("adding exec event (%s) trigger for coprocess (%s)", cp.ExecEvent, cp.Name)
	err := events.GlobalEvents.AddTriggerByEventName(cp.ExecEvent, &execTrigger)
	if err != nil {
		return errors.Errorf("Unable to add exec trigger to event (%s), coprocess (%s) will be disabled",
			cp.ExecEvent, cp.Name)
	}

	if len(cp.StopExec) > 1 {
		log.Tracef("adding prestop command event (%s) trigger for coprocess (%s)", cp.StopEvent, cp.Name)
		err = events.GlobalEvents.AddTriggerByEventName(cp.StopEvent, &preStopTrigger)
		if err != nil {
			return errors.Errorf("Unable to add prestop trigger to event (%s), coprocess (%s) will be disabled",
				cp.ExecEvent, cp.Name)
		}
	}

	log.Tracef("adding stop command event (%s) trigger for coprocess (%s)", cp.StopEvent, cp.Name)
	err = events.GlobalEvents.AddTriggerByEventName(cp.StopEvent, &stopTrigger)
	if err != nil {
		return errors.Errorf("Unable to add stop trigger to event (%s), coprocess (%s) will be disabled",
			cp.ExecEvent, cp.Name)
	}

	return nil
}

func readCoprocesses(settings api.Settings) ([]Coprocess, error) {
	if !settings.IsSet("coprocess") {
		return nil, errors.New("coprocess section in configuration not found")
	}

	var coprocesses []Coprocess

	for section := range settings.GetStringMap("coprocess") {
		cp, err := initCoprocess(section, settings)
		if err != nil {
			log.Warnf("error parsing coprocess configuration: %v", err)
		} else {
			log.Tracef("parsed coprocess definition (%s) for (%s) coprocess", section, cp.Name)
			coprocesses = append(coprocesses, cp)
		}
	}

	return coprocesses, nil
}

func initCoprocess(section string, settings api.Settings) (Coprocess, error) {
	name := settings.GetString(fmt.Sprintf("coprocess.%s.name", section))
	cpExec := settings.GetStringSlice(fmt.Sprintf("coprocess.%s.exec", section))
	stopExec := settings.GetStringSlice(fmt.Sprintf("coprocess.%s.stop_exec", section))
	user := settings.GetString(fmt.Sprintf("coprocess.%s.user", section))
	restarts := settings.GetString(fmt.Sprintf("coprocess.%s.restarts", section))
	background := settings.GetBool(fmt.Sprintf("coprocess.%s.background", section))
	execEvent := settings.GetString(fmt.Sprintf("coprocess.%s.exec_event", section))
	stopEvent := settings.GetString(fmt.Sprintf("coprocess.%s.stop_event", section))

	cp, err := New(name, cpExec, stopExec, user, restarts, background, execEvent, stopEvent,
		settings)

	if err != nil {
		err.ConfigSectionName = section
		return Coprocess{}, err
	}

	return cp, nil
}
