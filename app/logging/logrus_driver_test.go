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

package logging

import (
	slog "github.com/go-eden/slf4go"
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestName(t *testing.T) {
	driver := LogrusDriver{}
	name := driver.Name()
	if name != "logrus" {
		t.Error("log driver's name doesn't match convention")
	}
}

func TestConvertLogrusLevelToSlf4goLevelKnownLevels(t *testing.T) {
	attemptToConvertLogrusLevel(log.TraceLevel, slog.TraceLevel, t)
	attemptToConvertLogrusLevel(log.DebugLevel, slog.DebugLevel, t)
	attemptToConvertLogrusLevel(log.InfoLevel, slog.InfoLevel, t)
	attemptToConvertLogrusLevel(log.WarnLevel, slog.WarnLevel, t)
	attemptToConvertLogrusLevel(log.ErrorLevel, slog.ErrorLevel, t)
	attemptToConvertLogrusLevel(log.PanicLevel, slog.PanicLevel, t)
	attemptToConvertLogrusLevel(log.FatalLevel, slog.FatalLevel, t)
}

func TestConvertLogrusLevelToSlf4goLevelUnknownLevel(t *testing.T) {
	unknownLevelInt := 1000
	logrusLevel := log.Level(unknownLevelInt)
	slfLevel := ConvertLogrusLevelToSlf4goLevel(logrusLevel)
	if int(slfLevel) != unknownLevelInt {
		t.Errorf("Unexpected level conversion expected(%d) was (%d)",
			unknownLevelInt, slfLevel)
	}
}

func TestConvertSlf4goLevelToLogrusLevelKnownLevels(t *testing.T) {
	attemptToConvertSlfLevel(slog.TraceLevel, log.TraceLevel, t)
	attemptToConvertSlfLevel(slog.DebugLevel, log.DebugLevel, t)
	attemptToConvertSlfLevel(slog.InfoLevel, log.InfoLevel, t)
	attemptToConvertSlfLevel(slog.WarnLevel, log.WarnLevel, t)
	attemptToConvertSlfLevel(slog.ErrorLevel, log.ErrorLevel, t)
	attemptToConvertSlfLevel(slog.PanicLevel, log.PanicLevel, t)
	attemptToConvertSlfLevel(slog.FatalLevel, log.FatalLevel, t)
}

func TestConvertSlf4goLevelToLogrusLevelUnknownLevel(t *testing.T) {
	unknownLevelInt := 1000
	slfLevel := slog.Level(unknownLevelInt)
	logrusLevel := ConvertSlf4goLevelToLogrusLevel(slfLevel)
	if int(logrusLevel) != unknownLevelInt {
		t.Errorf("Unexpected level conversion expected(%d) was (%d)",
			unknownLevelInt, slfLevel)
	}
}

func TestTimeFromMicroseconds(t *testing.T) {
	unixMicros := int64(1594328693237135)
	unixNanos := int64(1594328693237135000)

	timeStamp := timeFromMicroseconds(unixMicros)

	if timeStamp.UnixNano() != unixNanos {
		t.Errorf("Failed to convert microseconds to nanoseconds timestamp "+
			"for input (%d) ms and output (%d) ns", unixMicros, timeStamp.UnixNano())
	}
}

func attemptToConvertLogrusLevel(expectedLogrusLevel log.Level, slfLevel slog.Level, t *testing.T) {
	actualLogrusLevel := ConvertSlf4goLevelToLogrusLevel(slfLevel)
	if actualLogrusLevel != expectedLogrusLevel {
		t.Errorf("Couldn't convert level, actual (%v) doesn't match expected (%v)",
			actualLogrusLevel, expectedLogrusLevel)
	}
}

func attemptToConvertSlfLevel(expectedSlfLevel slog.Level, logrusLevel log.Level, t *testing.T) {
	actualSlfLevel := ConvertLogrusLevelToSlf4goLevel(logrusLevel)
	if actualSlfLevel != expectedSlfLevel {
		t.Errorf("Couldn't convert level, actual (%v) doesn't match expected (%v)",
			actualSlfLevel, expectedSlfLevel)
	}
}
