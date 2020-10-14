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

package nginx

import (
	"fmt"
	slog "github.com/go-eden/slf4go"
	"github.com/nginxinc/nginx-wrapper/lib/api"
	"github.com/nginxinc/nginx-wrapper/lib/events"
	"github.com/tevino/abool"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var log = slog.NewLogger("pm")

// ProcessMonitor represents a WaitGroup extended for monitoring the lifecycle
// of a NGINX process.
type ProcessMonitor struct {
	sync.WaitGroup
	Stop    chan struct{}
	started *abool.AtomicBool
	process *os.Process
}

const secsToWaitForReload = 2

// NewProcessMonitor creates a new fully configured instance.
func NewProcessMonitor() *ProcessMonitor {
	pm := &ProcessMonitor{
		Stop:    make(chan struct{}),
		started: abool.NewBool(false),
	}

	signal.Notify(api.Signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGUSR2)
	go func() {
		// Stop processing new signals because we are shutting down
		defer close(api.Signals)

		sigLog := slog.NewLogger("signals")

		i := 0
		for sig := range api.Signals {
			sigLog.Debugf("received signal %s", sig)

			switch sig {
			case syscall.SIGINT:
				fallthrough
			case syscall.SIGTERM:
				sigLog.Info("shutdown signal received - shutting down")
				pm.Shutdown(sig.String())

				// Exit immediately if we get multiple INT or TERM signals
				if i > 0 {
					os.Exit(1)
				}
				i++
			case syscall.SIGHUP:
				fallthrough
			case syscall.SIGUSR2:
				pm.Reload(fmt.Sprintf("main process received HUP (%v)", sig))
			}
		}
	}()

	return pm
}

// Increment adds 1 to the WaitGroup counter.
func (pm *ProcessMonitor) Increment() {
	pm.Add(1)
}

// Reload starts the process of reconfiguring the wrapped nginx process.
func (pm *ProcessMonitor) Reload(cause string) {
	if !pm.started.IsSet() {
		log.Info("nginx process not yet started - can't HUP")
		return
	}

	log.Infof("SIGHUP sent to nginx process due to: %s", cause)
	if pm.process == nil {
		log.Error("can't issue HUP because there is no process assigned")
		return
	}

	// Wait for last reload to finish
	for events.GlobalEvents.ReloadStarted.IsSet() {
		time.Sleep(secsToWaitForReload * time.Second)
	}

	// Trigger the pre-reload event so that templating and other tasks can happen
	metadata := map[string]interface{}{"reload_cause": cause}
	reloadErrors := events.GlobalEvents.NginxPreReload.Trigger(metadata)
	if len(reloadErrors) > 0 {
		for _, err := range reloadErrors {
			log.Errorf("%+v", err)
		}
		panic("pre-reload event trigger caused error")
	}

	// Mark that we are now in the nginx reload process
	events.GlobalEvents.ReloadStarted.SetToIf(false, true)

	sighupErr := pm.process.Signal(syscall.SIGHUP)
	if sighupErr != nil {
		// Since HUP failed, we have effectively stopped reloading
		// so we mark the global flag as such
		events.GlobalEvents.ReloadStarted.UnSet()
		log.Errorf("SIGHUP of nginx process (%v) failed: %v",
			pm.process.Pid, sighupErr)
	}
}

// Shutdown gracefully stops the wrapped nginx process.
func (pm *ProcessMonitor) Shutdown(cause string) {
	// Mark nginx as shutdown and exit if already shutdown
	if pm.started.SetToIf(true, false) {
		log.Infof("process stopped due to: %s", cause)
		close(pm.Stop)
	}
}
