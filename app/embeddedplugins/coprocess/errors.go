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
	"fmt"
	"github.com/nginxinc/nginx-wrapper/lib/osenv"
	"strings"
)

// InitError is throw when a Coprocess instance can't be instantiated.
type InitError struct {
	InitFailureMessages []string
	ConfigSectionName   string
}

func (e *InitError) Error() string {
	if len(e.InitFailureMessages) < 1 {
		return fmt.Sprintf("error initializing coprocess section (%s)",
			e.ConfigSectionName)
	}

	var sb strings.Builder
	for _, e := range e.InitFailureMessages {
		sb.WriteString("    ")
		sb.WriteString(e)
		sb.WriteString(osenv.LineBreak)
	}

	return fmt.Sprintf("error initializing coprocess section (%s):%s%s",
		e.ConfigSectionName, osenv.LineBreak, sb.String())
}
