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
	slog "github.com/go-eden/slf4go"
	"github.com/nginxinc/nginx-wrapper/lib/events"
)

func logEvents() {
	log := slog.NewLogger("events")

	if log.IsDebugEnabled() {
		for _, eventName := range events.GlobalEvents.EventNames() {
			event := events.GlobalEvents.FindEventByName(eventName)
			log.Debugf("Event registered: %v", event)
		}
	}
}
