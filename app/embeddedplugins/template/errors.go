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

package template

import "strings"

// ProcessingError is an error that occurs when templates are being applied to nginx configuration.
type ProcessingError struct {
	Message              string
	TemplateFile         string
	OutputFile           string
	TemplateName         string
	IsATemplatingProblem bool
	Err                  error
}

func (e *ProcessingError) Error() string {
	builder := strings.Builder{}

	builder.WriteString(e.Message)

	if e.TemplateName != "" {
		builder.WriteString(" with template named (")
		builder.WriteString(e.TemplateName)
		builder.WriteString(")")
	}

	if e.TemplateFile != "" {
		builder.WriteString(" with template file (")
		builder.WriteString(e.TemplateFile)
		builder.WriteString(")")
	}

	if e.OutputFile != "" {
		builder.WriteString(" to output file (")
		builder.WriteString(e.OutputFile)
		builder.WriteString(")")
	}

	if e.Err != nil {
		builder.WriteString(": ")
		builder.WriteString(e.Err.Error())
	}

	return builder.String()
}

// Cause returns the cause of the error - the same as the Unwrap() method.
func (e *ProcessingError) Cause() error {
	return e.Err
}

func (e *ProcessingError) Unwrap() error {
	return e.Err
}
