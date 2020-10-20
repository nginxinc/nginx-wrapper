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
	"bufio"
	"fmt"
	slog "github.com/go-eden/slf4go"
	"github.com/nginxinc/nginx-wrapper/lib/api"
	"github.com/nginxinc/nginx-wrapper/lib/events"
	"github.com/pkg/errors"
	"github.com/tevino/abool"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// RestartPolicy represents what we do when a process has stopped.
// Valid values are defined in the constants below or alternatively
// an integer value represented as a string.
type RestartPolicy string

const (
	// Unlimited restart policy means that the process will always be restarted if it exits.
	Unlimited RestartPolicy = "unlimited"
	// Never restart policy means that the process will never be restarted if it exits.
	Never RestartPolicy = "never"
	// Terminating restart policy means that the wrapper is attempting to end the process
	Terminating RestartPolicy = "terminating"

	secondsToWaitForProcessExit = 5
)

// Coprocess represents a process that will run alongside nginx and be
// tied to a nginx start event and stop event.
type Coprocess struct {
	Name       string
	Exec       []string
	StopExec   []string
	User       string
	Restarts   RestartPolicy
	Background bool
	ExecEvent  string
	StopEvent  string
	Done       *abool.AtomicBool

	cmd     *exec.Cmd
	cmdLock *sync.RWMutex
	log     *slog.Logger
}

func (c Coprocess) String() string {
	return fmt.Sprintf("{%s %v %v %v %s %s}",
		c.Name, c.Exec, c.Restarts, c.Background, c.ExecEvent, c.StopEvent)
}

// MaxRestarts returns the maximum number of restarts for the associated process with -1 indicating
// an unlimited amount of restarts.
func (c *Coprocess) MaxRestarts() int {
	if c.Restarts == Never || c.Restarts == Terminating {
		return 0
	}
	if c.Restarts == Unlimited {
		return -1
	}

	i, err := strconv.Atoi(string(c.Restarts))
	if err != nil {
		slog.Fatalf("invalid restart policy (%v): %v", c.Restarts, errors.WithStack(err))
	}

	return i
}

// ExecuteCoprocess invokes the associated process and optionally runs it repeatedly according to the
// associated restart policy.
func (c *Coprocess) ExecuteCoprocess() error {
	maxRestarts := c.MaxRestarts()
	restartCount := 0

	for {
		if restartCount < 1 {
			c.log.Tracef("initiating coprocess (%s)", c.Name)
		} else {
			c.log.Tracef("initiating restart (%d) of coprocess (%s)", restartCount, c.Name)
		}

		c.cmdLock.Lock()

		// Don't start process if it is marked as supposed to end
		if c.Restarts == Terminating {
			break
		}

		cmd, initCmdErr := c.initCoprocessCmd(c.Exec, c.log)
		if initCmdErr != nil {
			c.cmdLock.Unlock()
			return errors.Wrapf(initCmdErr, "unable to initiate coprocess (%s)",
				c.Name)
		}
		c.cmd = cmd

		c.log.Tracef("starting coprocess (%s)", c.Name)
		startCmdErr := cmd.Start()
		if startCmdErr != nil {
			if c.User != "" {
				c.cmdLock.Unlock()
				return errors.Wrapf(startCmdErr, "unable to start coprocess (%s) running as user (%s)",
					c.Name, c.User)
			}

			c.cmdLock.Unlock()
			return errors.Wrapf(startCmdErr, "unable to start coprocess (%s)", c.Name)
		}

		// Wait for process to fully start before we pass a reference to the cmd
		for cmd.Process == nil {
			time.Sleep(1 * time.Second)
		}

		c.cmdLock.Unlock()

		c.cmdLock.RLock()
		c.waitOnCmd(c.cmd, c.log)
		c.cmdLock.RUnlock()

		if c.Restarts != Unlimited && restartCount >= maxRestarts {
			break
		} else {
			restartCount = 1 + restartCount
		}
	}

	return nil
}

// ExecuteStopCmd executes the optional command associated with the coprocess to run when the
// stop event is encountered.
func (c *Coprocess) ExecuteStopCmd() error {
	stopLog := slog.NewLogger("stop-" + c.Name)

	if len(c.StopExec) < 1 {
		c.log.Tracef("no stop command specified for coprocess (%s) exiting immediately",
			c.Name)
		return nil
	}
	runLog := slog.NewLogger("stop-" + c.Name)
	cmd, initCmdErr := c.initCoprocessCmd(c.StopExec, runLog)
	if initCmdErr != nil {
		return errors.Wrapf(initCmdErr, "unable to initiate stop command for coprocess (%s)",
			c.Name)
	}

	// Don't allow the process to restart because we are now in the shutdown sequence
	if c.Restarts != Terminating {
		c.Restarts = Terminating
	}

	stopLog.Tracef("issuing stop command for coprocess (%s)", c.Name)
	stopCmdErr := cmd.Start()
	if stopCmdErr != nil {
		if c.User != "" {
			return errors.Wrapf(stopCmdErr, "unable to execute stop command for coprocess "+
				"(%s) running as user (%s)",
				c.Name, c.User)
		}

		return errors.Wrapf(stopCmdErr, "unable to execute stop command for coprocess (%s)", c.Name)
	}

	c.waitOnCmd(cmd, runLog)
	stopLog.Tracef("stop command for coprocess (%s) completed", c.Name)

	return nil
}

func (c *Coprocess) waitOnCmd(cmd *exec.Cmd, runLog *slog.Logger) {
	err := cmd.Wait()
	if err != nil {
		if err.Error() == "signal: terminated" {
			runLog.Debugf("coprocess (%s) process (%s) exited: %s",
				c.Name, cmd.Path, err)
		} else if strings.HasPrefix(err.Error(), "exit status ") {
			runLog.Infof("coprocess (%s) process (%s) exited with non-zero code: %d",
				c.Name, cmd.Path, cmd.ProcessState.ExitCode())
		} else {
			runLog.Errorf("coprocess (%s) process (%s) exited with error: %s",
				c.Name, cmd.Path, err)
		}
	} else {
		runLog.Debugf("coprocess (%s) process (%s) exited", c.Name, cmd.Path)
	}
}

func (c *Coprocess) initCoprocessCmd(execCmd []string, runLog *slog.Logger) (*exec.Cmd, error) {
	cmdPath := execCmd[0]
	var args []string
	if len(execCmd) > 1 {
		args = execCmd[1:]
	} else {
		args = []string{}
	}

	cmd := exec.Command(cmdPath, args...)

	runLog.Tracef("initiated cmd: %v", cmd)

	// Assign runtime user if specified
	if c.User != "" {
		runLog.Debugf("running coprocess (%s) as user (%s)", c.Name, c.User)

		cmdUser, userLookupErr := user.Lookup(c.User)
		if userLookupErr != nil {
			return nil, errors.Wrapf(userLookupErr, "problem looking up user "+
				"(%s) specified for coprocess (%s)", c.User, c.Name)
		}
		uid, parseIntErr := strconv.ParseUint(cmdUser.Uid, 10, 32)
		if parseIntErr != nil {
			return nil, errors.Wrapf(parseIntErr, "unable to parse uid (%s) for "+
				"coprocess (%s)", cmdUser.Uid, c.Name)
		}

		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: uint32(uid),
			},
		}
	}

	err := c.attachCmdLogger(cmd, runLog)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return cmd, nil
}

func (c *Coprocess) attachCmdLogger(cmd *exec.Cmd, runLog *slog.Logger) error {
	if cmd == nil {
		return errors.New("cmd not initialized yet")
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.WithStack(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return errors.WithStack(err)
	}

	reader := io.MultiReader(stdout, stderr)
	scan := bufio.NewScanner(reader)
	go func() {
		for scan.Scan() {
			line := scan.Text()
			if len(line) == 0 {
				continue
			}

			runLog.Info(line)
		}
	}()

	return nil
}

// TerminateCoprocess terminates the associated coprocess by sending (in this order)
// SIGTERM, SIGINT, or SIGKILL until the process exits.
func (c *Coprocess) TerminateCoprocess() error {
	// Mark that we don't want to restart the process after it exits
	if c.Restarts != Terminating {
		c.Restarts = Terminating
	}

	// If the process has exited and isn't set to restart, then we don't need to do anything
	if c.Done.IsSet() {
		c.log.Trace("coprocess already exited")
		return nil
	}

	c.cmdLock.RLock()
	if c.cmd == nil {
		c.log.Trace("coprocess state unavailable")
		c.cmdLock.RUnlock()
		return nil
	}

	process := c.cmd.Process
	sigTermErr := c.sendSignal(process, syscall.SIGTERM, c.Name)
	if sigTermErr != nil {
		c.log.Warn(sigTermErr)
	} else {
		// sleep to wait for process to exit
		time.Sleep(secondsToWaitForProcessExit * time.Second)
	}
	c.cmdLock.RUnlock()

	// We check to see if the process exited after sending a SIGTERM, if so then we can just return
	// Notice that we don't used the process variable as defined above because it is possible that
	// there could be a weird race condition where we made it one more time through the restart
	// loop and the pid changed.
	if c.Done.IsSet() {
		return nil
	}

	// Reacquire process reference because it may have changed because the app may be rapidly restarting
	c.cmdLock.RLock()
	process = c.cmd.Process
	sigIntErr := c.sendSignal(process, syscall.SIGINT, c.Name)
	if sigIntErr != nil {
		c.log.Warn(sigIntErr)
	} else {
		// sleep to wait for process to exit
		time.Sleep(secondsToWaitForProcessExit * time.Second)
	}
	c.cmdLock.RUnlock()

	if c.Done.IsSet() {
		return nil
	}

	// If we are here, the process didn't exit nicely

	// Reacquire process reference because it may have changed because the app may be rapidly restarting
	c.cmdLock.RLock()
	process = c.cmd.Process
	sigKillErr := c.sendSignal(process, syscall.SIGKILL, c.Name)
	if sigIntErr != nil {
		return errors.Wrapf(sigKillErr, "SIGKILL failed for coprocess (%s) for pid (%d)",
			c.Name, process.Pid)
	}
	c.cmdLock.RUnlock()

	if c.Done.IsSet() {
		return nil
	}

	// Reacquire process reference because it may have changed because the app may be rapidly restarting
	c.cmdLock.RLock()
	process = c.cmd.Process

	// We wait for the process to exit after sending a SIGKILL
	if c.cmd != nil && process != nil {
		_, err := process.Wait()
		if err != nil {
			c.log.Errorf("error waiting for coprocess (%s) with pid (%d) to stop after SIGKILL: %v",
				c.Name, c.cmd.Process.Pid, err)
		}
		if c.log.IsDebugEnabled() {
			c.log.Debugf("coprocess (%s) exited with successfully after SIGKILL",
				c.Name)
		}
	} else {
		c.cmdLock.RUnlock()
		return errors.Errorf("no process information returned after coprocess (%s) was killed - this"+
			"is an unexpected state", c.Name)
	}
	c.cmdLock.RUnlock()

	return nil
}

func (c *Coprocess) sendSignal(process *os.Process, signal os.Signal, name string) error {
	c.log.Tracef("sending (%v) to coprocess (%s) with pid (%d)", signal.String(),
		name, process.Pid)

	// If we are here, then the process is still running
	err := process.Signal(signal)
	if err != nil {
		shouldProcessEnd := signal == syscall.SIGTERM || signal == syscall.SIGINT || signal == syscall.SIGKILL
		// If for whatever reason the process ends before the signal can be sent and we were intending to end the
		// process, we handle the resulting error
		if shouldProcessEnd && err.Error() == "os: process already finished" {
			c.log.Debugf("(%v) failed for coprocess (%s) pid (%d): %v",
				signal.String(), name, process.Pid, err)
		} else {
			return errors.Wrapf(err, "(%v) failed for coprocess (%s) pid (%d)",
				signal.String(), name, process.Pid)
		}
	}

	return nil
}

// New creates a new instance of Coprocess and validates that the passed parameters are valid.
func New(name string, exec []string, stopExec []string, user string, restarts string, background bool,
	execEvent string, stopEvent string, settings api.Settings) (Coprocess, *InitError) {

	var initErrors []string

	if name == "" {
		initErrors = append(initErrors, "coprocess field 'name' is blank")
	}
	if len(exec) == 0 {
		initErrors = append(initErrors, "coprocess field 'exec' is empty")
	}
	// note: the Terminating policy is filtered out here and can't be set by the user
	if !(restarts == string(Unlimited) || restarts == string(Never) || isASCIIDigit(restarts)) {
		initErrors = append(initErrors, fmt.Sprintf("coprocess field 'restarts' is set to an invalid value "+
			"(%s) - it must be 'never' or 'unlimited'", restarts))
	}

	eventNames := events.GlobalEvents.EventNames()
	if !sliceContainsAny(eventNames, execEvent) {
		initErrors = append(initErrors, fmt.Sprintf("coprocess field 'exec_event' was not set to a valid event "+
			"name (%s)", execEvent))
	}
	if !sliceContainsAny(eventNames, stopEvent) {
		initErrors = append(initErrors, fmt.Sprintf("coprocess field 'stop_event' was not set to a valid event "+
			"name (%s)", stopEvent))
	}

	if len(initErrors) > 0 {
		err := InitError{InitFailureMessages: initErrors}
		return Coprocess{}, &err
	}

	interpolatedArgs := interpolateCommandWithSettings(
		map[string][]string{"exec": exec, "stopExec": stopExec},
		settings)

	instance := Coprocess{
		Name:       name,
		Exec:       interpolatedArgs["exec"],
		StopExec:   interpolatedArgs["stopExec"],
		User:       user,
		Restarts:   RestartPolicy(restarts),
		Background: background,
		ExecEvent:  execEvent,
		StopEvent:  stopEvent,
		Done:       abool.NewBool(false),

		cmdLock: &sync.RWMutex{},
		log:     slog.NewLogger(name),
	}

	return instance, nil
}

// interpolateCommandWithSettings iterates each value of a command (as represented by an array)
// and interpolates substitutions into the command. For example, if the stop_exec command contained:
// [ "curl", "-X", "POST", "http://my.service/remove/${host_id}" ] then the command could be interpolated as
// [ "curl", "-X", "POST", "http://my.service/remove/9bcc0df29af9454298607489a54040e2" ].
func interpolateCommandWithSettings(arguments map[string][]string, settings api.Settings) map[string][]string {
	replacer := buildSubstitutionsList(settings)

	processed := make(map[string][]string, len(arguments))

	for argName, v := range arguments {
		processed[argName] = interpolateAll(v, replacer)
	}

	return processed
}

// buildSubstitutionsList creates a string.Replacer containing the substitutions to use for interpolating
// exec and stop_exec commands.
func buildSubstitutionsList(settings api.Settings) *strings.Replacer {
	substitutions := []string{
		"${host_id}", settings.GetString("host_id"),
		"${modules_path}", settings.GetString("modules_path"),
		"${nginx_binary}", settings.GetString("nginx_binary"),
		"${plugin_path}", settings.GetString("plugin_path"),
		"${run_path}", settings.GetString("run_path"),
		"${vcpu_count}", settings.GetString("vcpu_count"),
		"${last_reload_time}", settings.GetString("last_reload_time"),
		"${wrapper_pid}", strconv.Itoa(os.Getpid()),
	}

	for k, v := range settings.GetStringMapString("env") {
		substitutions = append(substitutions, "${"+k+"}")
		substitutions = append(substitutions, v)
	}

	return strings.NewReplacer(substitutions...)
}
