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
	"time"
)

// initLogrusDriver sets up the slf4go global configuration to map to a
// logrus global configuration.
func initLogrusDriver(prependLoggerName bool) {
	slog.SetDriver(&LogrusDriver{PrependLoggerName: prependLoggerName})
}

// LogrusDriver is a slf4go compatible logger implementation.
type LogrusDriver struct {
	// LoggerToLevelMapping allows you to specify a logger name that you want to assign a specific log level
	LoggerToLevelMapping map[string]slog.Level
	// When enabled the logger name is appended to the start of all log lines
	PrependLoggerName bool
}

// Name indicates the name of the logger implementation.
func (ld *LogrusDriver) Name() string {
	return "logrus"
}

// Print outputs the log line using the backing logrus driver implementation.
func (ld *LogrusDriver) Print(sl *slog.Log) {
	logrusLevel := ConvertSlf4goLevelToLogrusLevel(sl.Level)
	// Exit early if we can't log for the current level
	if !log.IsLevelEnabled(logrusLevel) {
		return
	}

	// In order to pass along fields to logrus, we need to convert them to
	// an logrus Entry type, so that we can invoke methods on Entry directly.
	entry := extractLogrusEntry(sl)

	// Invoke the underlying logrus driver to log the messages gathered
	// via the slf4go API
	if sl.Format == nil && ld.PrependLoggerName { // prepend logger names
		prepended := prependArg(sl.Logger+": ", sl.Args)
		entry.Log(logrusLevel, prepended...)
	} else if sl.Format != nil && ld.PrependLoggerName { // prepend logger names
		format := sl.Logger + ": " + *sl.Format
		entry.Logf(logrusLevel, format, sl.Args...)
	} else if sl.Format == nil { // leave out logger names
		entry.Log(logrusLevel, sl.Args...)
	} else { // leave out logger names
		entry.Logf(logrusLevel, *sl.Format, sl.Args...)
	}
}

func prependArg(firstArg interface{}, args []interface{}) []interface{} {
	newArgs := make([]interface{}, len(args)+1)
	newArgs[0] = firstArg

	for i, arg := range args {
		newArgs[i+1] = arg
	}

	return newArgs
}

// extractLogrusEntry returns a logrus Entry type with the slf4go fields populated.
func extractLogrusEntry(sl *slog.Log) *log.Entry {
	var entry *log.Entry

	if sl.Fields != nil {
		entry = log.WithFields(log.Fields(sl.Fields))
	} else {
		fields := log.Fields{}
		entry = log.WithFields(fields)
	}
	// slf4go uses microseconds to store logged time and we need to convert to
	// time.Time, so we convert microseconds to nanoseconds in order to not lose
	// precision.
	timeStamp := timeFromMicroseconds(sl.Time)
	entry = entry.WithTime(timeStamp)

	return entry
}

// timeFromMicroseconds converts a microseconds unix timestamp to a Golang
// time.Time data structure.
func timeFromMicroseconds(micros int64) time.Time {
	//nolint:gomnd
	microSecondConversionFactor := int64(1000)
	nsec := micros * microSecondConversionFactor
	return time.Unix(0, nsec)
}

// GetLevel retrieves the log level of the specified logger,
// it should return the lowest Level that could be print,
// which can help invoker to decide whether prepare print or not.
func (ld *LogrusDriver) GetLevel(logger string) slog.Level {
	// If a logger to level mapping was set, then we use that value
	specificLevel, ok := ld.LoggerToLevelMapping[logger]
	if ok {
		return specificLevel
	}

	// Otherwise, we use the log level set by logrus globally
	logrusGlobalLevel := log.GetLevel()
	return ConvertLogrusLevelToSlf4goLevel(logrusGlobalLevel)
}

// ConvertLogrusLevelToSlf4goLevel converts a log level from logrus to slf4go.
func ConvertLogrusLevelToSlf4goLevel(logrusLevel log.Level) slog.Level {
	switch logrusLevel {
	case log.TraceLevel:
		return slog.TraceLevel
	case log.DebugLevel:
		return slog.DebugLevel
	case log.InfoLevel:
		return slog.InfoLevel
	case log.WarnLevel:
		return slog.WarnLevel
	case log.ErrorLevel:
		return slog.ErrorLevel
	case log.PanicLevel:
		return slog.PanicLevel
	case log.FatalLevel:
		return slog.FatalLevel
	}

	// If we have a log level that isn't specified above because it is a
	// custom int value, we can map it to a logrus compatible int value
	// by recasting it.
	return slog.Level(logrusLevel)
}

// ConvertSlf4goLevelToLogrusLevel converts a log level from slf4go to logrus.
func ConvertSlf4goLevelToLogrusLevel(slfLevel slog.Level) log.Level {
	switch slfLevel {
	case slog.TraceLevel:
		return log.TraceLevel
	case slog.DebugLevel:
		return log.DebugLevel
	case slog.InfoLevel:
		return log.InfoLevel
	case slog.WarnLevel:
		return log.WarnLevel
	case slog.ErrorLevel:
		return log.ErrorLevel
	case slog.PanicLevel:
		return log.PanicLevel
	case slog.FatalLevel:
		return log.FatalLevel
	}

	// If we have a log level that isn't specified above because it is a
	// custom int value, we can map it to a logrus compatible int value
	// by recasting it.
	return log.Level(slfLevel)
}
