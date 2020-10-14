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

package events

import (
	"fmt"
	slog "github.com/go-eden/slf4go"
	"github.com/pkg/errors"
	"github.com/tevino/abool"
	"strings"
)

// Maximum size in characters of a process id.
const maxPidWidth = 19

var log = slog.NewLogger("events")

// Event emitted when before NGINX starts for the first time.
var preStartEvent = Event{
	Name:   "pre-start",
	Origin: "nginx",
}

// Event emitted when NGINX starts for the first time.
var startEvent = Event{
	Name:   "start",
	Origin: "nginx",
}

// Event emitted every time a NGINX worker process is started.
var startWorkerEvent = Event{
	Name:   "start-worker",
	Origin: "nginx",
}

// Event emitted when the main NGINX process exits.
var exitEvent = Event{
	Name:   "exit",
	Origin: "nginx",
}

// Event emitted every time a NGINX worker process exits.
var exitWorkerEvent = Event{
	Name:   "exit-worker",
	Origin: "nginx",
}

// Event emitted before NGINX reloads.
var preReloadEvent = Event{
	Name:   "pre-reload",
	Origin: "nginx",
}

// Event emitted when NGINX reloads.
var reloadEvent = Event{
	Name:   "reload",
	Origin: "nginx",
}

// GlobalEvents contains hardcoded NGINX events.
var GlobalEvents = RegisteredEvents{
	NginxPreStart:    &preStartEvent,
	NginxStart:       &startEvent,
	NginxWorkerStart: &startWorkerEvent,
	NginxExit:        &exitEvent,
	NginxWorkerExit:  &exitWorkerEvent,
	NginxPreReload:   &preReloadEvent,
	NginxReload:      &reloadEvent,
	ReloadStarted:    abool.NewBool(false),
}

// Trigger function to execute when an event has OnTrigger()
// invoked.
type Trigger struct {
	Name     string
	Function func(message Message) error
}

func (t *Trigger) String() string {
	return fmt.Sprintf("{%s}", t.Name)
}

// Event represents a unique wrapper event.
type Event struct {
	// Event name
	Name string
	// System of origin for event
	Origin string
	// Array of triggers to be invoked when event is broadcast
	OnTrigger []*Trigger
	// Final trigger that will be invoked after all of the other triggers
	alwaysLastTrigger *Trigger
}

func (e *Event) String() string {
	b := strings.Builder{}
	for i, t := range e.OnTrigger {
		b.WriteString(t.Name)

		if i < len(e.OnTrigger)-1 {
			b.WriteString(",")
		}
	}

	return fmt.Sprintf("{ %s triggers=[%s] }",
		e.ID(), b.String())
}

// AddTrigger adds a trigger to an Event.
func (e *Event) AddTrigger(trigger *Trigger) {
	if e.OnTrigger == nil {
		e.OnTrigger = []*Trigger{trigger}
	} else {
		e.OnTrigger = append(e.OnTrigger, trigger)
	}
}

// AddFinalTrigger adds the penultimate (final) trigger to an event.
func (e *Event) AddFinalTrigger(trigger *Trigger) {
	e.alwaysLastTrigger = trigger
}

// Trigger runs a closure associated with event and broadcasts
// message to event channel.
func (e *Event) Trigger(metadata map[string]interface{}) []error {
	var errs []error

	message := Message{
		Event:    *e,
		Metadata: metadata,
	}

	if e.OnTrigger != nil {
		for _, t := range e.OnTrigger {
			log.Tracef("invoking trigger (%s) for event (%s)", t.Name, e.ID())
			err := t.Function(message)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}
	if e.alwaysLastTrigger != nil {
		err := e.alwaysLastTrigger.Function(message)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// ID returns a unique ID for the event based on the Origin and Name.
func (e *Event) ID() string {
	return e.Origin + "." + e.Name
}

// Message is used when notifying that an event occurred.
type Message struct {
	Event    Event
	Metadata map[string]interface{}
}

// RegisteredEvents provides global access to events and flags.
type RegisteredEvents struct {
	NginxPreStart    *Event
	NginxStart       *Event
	NginxWorkerStart *Event
	NginxExit        *Event
	NginxWorkerExit  *Event
	NginxPreReload   *Event
	NginxReload      *Event

	ReloadStarted *abool.AtomicBool

	firstStartWorkerProcessesFound bool
	firstStartWorkerProcessFound   bool
	workerCount                    int
}

// EventNames returns the names of the available nginx events.
func (re *RegisteredEvents) EventNames() []string {
	names := []string{
		re.NginxExit.Name,
		re.NginxWorkerExit.Name,
		re.NginxPreReload.Name,
		re.NginxPreStart.Name,
		re.NginxReload.Name,
		re.NginxStart.Name,
		re.NginxWorkerStart.Name,
	}

	return names
}

// AddTriggerByEventName allows for adding a reference to a trigger that is to be associated with
// an event (as indicated by name).
func (re *RegisteredEvents) AddTriggerByEventName(eventName string, trigger *Trigger) error {
	switch eventName {
	case preStartEvent.Name:
		preStartEvent.AddTrigger(trigger)
	case startEvent.Name:
		startEvent.AddTrigger(trigger)
	case startWorkerEvent.Name:
		startWorkerEvent.AddTrigger(trigger)
	case exitEvent.Name:
		exitEvent.AddTrigger(trigger)
	case exitWorkerEvent.Name:
		exitWorkerEvent.AddTrigger(trigger)
	case preReloadEvent.Name:
		preReloadEvent.AddTrigger(trigger)
	case reloadEvent.Name:
		reloadEvent.AddTrigger(trigger)
	default:
		return errors.Errorf("unknown event name (%s)", eventName)
	}

	return nil
}

// FindEventByName allows for looking up an event object by its name.
func (re *RegisteredEvents) FindEventByName(eventName string) *Event {
	switch eventName {
	case preStartEvent.Name:
		return &preStartEvent
	case startEvent.Name:
		return &startEvent
	case startWorkerEvent.Name:
		return &startWorkerEvent
	case exitEvent.Name:
		return &exitEvent
	case exitWorkerEvent.Name:
		return &exitWorkerEvent
	case preReloadEvent.Name:
		return &preReloadEvent
	case reloadEvent.Name:
		return &reloadEvent
	default:
		return nil
	}
}

// AddTriggerToAllEvents adds a trigger to all events listed in RegisteredEvents.
func (re *RegisteredEvents) AddTriggerToAllEvents(trigger Trigger) {
	re.NginxPreStart.AddTrigger(&trigger)
	re.NginxPreReload.AddTrigger(&trigger)
	re.NginxReload.AddTrigger(&trigger)
	re.NginxStart.AddTrigger(&trigger)
	re.NginxWorkerStart.AddTrigger(&trigger)
	re.NginxExit.AddTrigger(&trigger)
	re.NginxWorkerExit.AddTrigger(&trigger)
}

// Init creates the event channel and populates GlobalEvents instance.
func Init() {
}

// ParseForTriggerableEvent parses a NGINX log line for matching text
// that is associated with an  event type. If an associated event type
// is found, then the Trigger() method of that event will be invoked.
//
// Note: This method is stateful in that it behaves differently depending
// on what previous lines were received. Unfortunately, this is a needed
// complexity due to the way NGINX logs.
func (re *RegisteredEvents) ParseForTriggerableEvent(line string) {
	/*
		Rough implementation:

		[start]         line ending with "start worker processes" and the first
		                "start worker process #####" line.
		[start-worker]  line ending with "start worker process #####"
		[exit]          not emitted from log parsing - emitted from process monitor
		[exit-worker]   line ending with " exit" when the worker count is > 0
		[reload]        the ReloadStarted flag has been set to true externally and
		                we get a "start worker process #####" log line

	*/

	// If no worker processes have been started yet, we are looking
	// for a sign that nginx is getting ready to kick off the first
	// worker process because that would correspond to an initial
	// start event. That "sign" is the message "start worker processes".
	if !re.firstStartWorkerProcessesFound && strings.HasSuffix(line, "start worker processes") {
		re.firstStartWorkerProcessesFound = true

		return
	}

	// Here we establish if this is the first time that workers have been
	// started. If so, we know that a start event will need to be emitted
	// when the first worker starts up.
	if re.firstStartWorkerProcessesFound {
		// We determine if the message text ends with a pid for a worker process
		// starting. If so, we know that another worker process is being kicked
		// off.
		workerPid := extractPidFromStartWorkerProcess(line)
		if workerPid >= 0 {
			re.parseStartEvents(line, workerPid)
			return
		}
	}

	// Emit events for when each worker process exits - if we see an exit when
	// all of the processes have already exited it is for the main process
	// and not a worker process. In that case, we want the "nginx-exit" event
	// which is implemented elsewhere to be emitted.
	if strings.HasSuffix(line, " exit") && re.workerCount > 0 {
		workerPid, workerTid, err := parsePidAndTid(line)
		if err != nil {
			log.Warn(err)
			return
		}

		re.workerCount--
		metadata := map[string]interface{}{
			"worker_pid":   workerPid,
			"worker_tid":   workerTid,
			"worker_count": re.workerCount,
		}
		exitWorkerErrors := exitWorkerEvent.Trigger(metadata)
		if len(exitWorkerErrors) > 0 {
			for _, err := range exitWorkerErrors {
				log.Error(err)
			}
			panic("exit event trigger caused error")
		}
	}
}

func (re *RegisteredEvents) parseStartEvents(line string, workerPid int64) {
	// If we have the first worker process started after nginx start up
	// or the first nginx worker process started after reload
	if !re.firstStartWorkerProcessFound || re.ReloadStarted.IsSet() {
		pid, tid, err := parsePidAndTid(line)
		if err != nil {
			log.Warn(err)
		}
		metadata := map[string]interface{}{
			"pid": pid,
			"tid": tid,
		}

		if re.ReloadStarted.IsSet() {
			reloadErrors := reloadEvent.Trigger(metadata)
			if len(reloadErrors) > 0 {
				for _, err := range reloadErrors {
					log.Error(err)
				}
				panic("reload event trigger caused error")
			}
			// Mark that we are no longer reloading because the reload event
			// triggers have finished running
			re.ReloadStarted.SetToIf(true, false)
		} else {
			startErrors := startEvent.Trigger(metadata)
			if len(startErrors) > 0 {
				for _, err := range startErrors {
					log.Error(err)
				}
				panic("start event trigger caused error")
			}

			re.firstStartWorkerProcessFound = true
		}
	}

	_, workerTid, err := parsePidAndTid(line)
	if err != nil {
		log.Warn(err)
	}

	// All workers will emit events indicating that they started
	re.workerCount++
	metadata := map[string]interface{}{
		"worker_pid":   workerPid,
		"worker_tid":   workerTid,
		"worker_count": re.workerCount,
	}
	startWorkerErrors := startWorkerEvent.Trigger(metadata)
	if len(startWorkerErrors) > 0 {
		for _, err := range startWorkerErrors {
			log.Error(err)
		}
		log.Panic("start worker event trigger caused error")
	}
}
