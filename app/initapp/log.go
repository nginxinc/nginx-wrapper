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
	slog "github.com/go-eden/slf4go"
	"github.com/nginxinc/nginx-wrapper/app/logging"
	"github.com/nginxinc/nginx-wrapper/lib/api"
	"github.com/nginxinc/nginx-wrapper/lib/events"
	"github.com/nginxinc/nginx-wrapper/lib/osenv"
	"os"
)

func initLog(settings api.Settings) {
	logInitErr := logging.InitLogger(settings)
	if logInitErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v%s", logInitErr, osenv.LineBreak)
		os.Exit(1)
	}
}

func initEventListenerLog() {
	eventLog := slog.NewLogger("event")

	if !eventLog.IsDebugEnabled() {
		return
	}

	logTrigger := events.Trigger{
		Name: "core.log-event-fired",
		Function: func(m events.Message) error {
			if eventLog.IsDebugEnabled() {
				eventLog.Debugf("[%s] triggered: %v", m.Event.ID(), m.Metadata)
			}

			return nil
		},
	}

	events.GlobalEvents.AddTriggerToAllEvents(logTrigger)
}
