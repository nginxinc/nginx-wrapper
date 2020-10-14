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
	slog "github.com/go-eden/slf4go"
	"github.com/nginxinc/nginx-wrapper/lib/events"
	"github.com/pkg/errors"
	"github.com/tevino/abool"
	"os/exec"
	"path"
	"syscall"
)

// RunCommand coordinates the run and shutdown of a binary executable at the specified
// path. The expectation is that this function will be executing an nginx instance.
func RunCommand(pm *ProcessMonitor, cmdPath string, args ...string) (*exec.Cmd, error) {
	log := slog.NewLogger("nginx-run")

	preStartErrors := events.GlobalEvents.NginxPreStart.Trigger(map[string]interface{}{})
	if len(preStartErrors) > 0 {
		for _, err := range preStartErrors {
			log.Errorf("%+v", err)
		}
		panic("pre-start event trigger caused error")
	}

	_, file := path.Split(cmdPath)
	cmd := exec.Command(cmdPath, args...)
	cmdErr := Cmd(cmd)

	if cmdErr != nil {
		return nil, errors.Errorf("error constructing command and connecting to pipes: %v",
			cmdErr)
	}

	pm.Increment()
	err := cmd.Start()
	if err != nil {
		pm.Done()
		return nil, errors.Wrapf(err, "error starting (%s)", file)
	}
	exited := abool.NewBool(false)
	go func() {
		defer pm.Done()

		// Mark nginx as started so that we can use this information elsewhere
		pm.started.SetToIf(false, true)
		pm.process = cmd.Process

		err := cmd.Wait()
		exited.SetToIf(false, true)
		if err != nil {
			log.Errorf("%s exited with error: %s", file, err)
		} else {
			log.Debugf("%s exited", file)
		}

		metadata := map[string]interface{}{
			"pid": pm.process.Pid,
		}
		exitErrors := events.GlobalEvents.NginxExit.Trigger(metadata)
		if len(exitErrors) > 0 {
			for _, err := range exitErrors {
				log.Errorf("%+v", err)
			}
			panic("exit event trigger caused error")
		}
	}()
	go func() {
		<-pm.Stop

		if exited.IsSet() {
			return
		}
		log.Infof("killing %s with signal %d", file, syscall.SIGTERM)
		killErr := syscall.Kill(cmd.Process.Pid, syscall.SIGTERM)

		if killErr != nil {
			log.Fatalf("unable to kill process [%v]: %+v",
				cmd.Process.Pid, killErr)
		}
	}()

	return cmd, nil
}
