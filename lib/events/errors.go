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

import "fmt"

// NginxLogParseError is thrown when there is a problem parsing a
// NGINX log line.
type NginxLogParseError struct {
	Message string
	LogLine string
	Err     error
}

func (e *NginxLogParseError) Error() string {
	var msg string

	if e.Message == "" && e.LogLine == "" && e.Err == nil {
		msg = ""
	} else if e.Err == nil && e.Message != "" && e.LogLine != "" {
		msg = fmt.Sprintf(e.Message, e.LogLine)
	} else if e.Err == nil && e.Message != "" && e.LogLine == "" {
		msg = e.Message
	} else {
		msg = fmt.Sprintf(e.Message+": %v", e.LogLine, e.Err)
	}

	return msg
}

// Cause returns the cause of the error - the same as the Unwrap() method.
func (e *NginxLogParseError) Cause() error {
	return e.Err
}

func (e *NginxLogParseError) Unwrap() error {
	return e.Err
}

func newNginxLogParserError(message string, line string) *NginxLogParseError {
	return &NginxLogParseError{
		Message: message,
		LogLine: line,
	}
}
