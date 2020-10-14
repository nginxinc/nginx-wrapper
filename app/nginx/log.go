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
	"bufio"
	"github.com/nginxinc/nginx-wrapper/lib/events"
	"github.com/pkg/errors"
	"io"
	"os/exec"
	"strings"

	slog "github.com/go-eden/slf4go"
)

// New creates a new instance of a nginx wrapper log.
func New(readers ...io.Reader) {
	runLog := slog.NewLogger("nginx")
	reader := io.MultiReader(readers...)
	scan := bufio.NewScanner(reader)
	go func() {
		for scan.Scan() {
			nginxLog(scan.Text(), *runLog)
		}
	}()
}

// Cmd process the logs for the passed Cmd object and outputs the
// logs within the nginx-wrapper logging configuration.
func Cmd(cmd *exec.Cmd) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.WithStack(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return errors.WithStack(err)
	}
	New(stdout, stderr)
	return nil
}

func nginxLog(l string, runLog slog.Logger) {
	if len(l) == 0 {
		return
	}

	var msg string

	if strings.Index(l, "nginx: ") == 0 {
		msg = strings.TrimSpace(l[7:])
	} else {
		msg = strings.TrimSpace(l)
	}

	leftBracketPos := strings.IndexRune(msg, '[')
	rightBracketPos := strings.IndexRune(msg, ']')

	var nginxLogLevel string

	if leftBracketPos < 0 || rightBracketPos < 0 {
		// do nothing because there was no runLog level indicated
	} else {
		nginxLogLevel = msg[leftBracketPos+1 : rightBracketPos]
	}

	msg = strings.TrimLeft(msg[rightBracketPos+1:], " \t")

	switch strings.ToUpper(nginxLogLevel) {
	case "NOTICE":
		runLog.Info(msg)
	case "WARNING":
		runLog.Warn(msg)
	case "ALERT":
		runLog.Error(msg)
	default:
		runLog.Info(msg)
	}

	events.GlobalEvents.ParseForTriggerableEvent(msg)
}
