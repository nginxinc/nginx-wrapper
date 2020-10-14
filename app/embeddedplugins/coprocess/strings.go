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
	"strings"
	"unicode"
)

// interpolateAll processes each string in sourceStrings and performs substitutions on the contents
// of each string using the already initialized replacer instance.
// nolint: interfacer
func interpolateAll(sourceStrings []string, replacer *strings.Replacer) []string {
	processed := make([]string, len(sourceStrings))
	for i, contents := range sourceStrings {
		processed[i] = replacer.Replace(contents)
	}

	return processed
}

// sliceContainsAny returns true if the specified slice contains at least one element matching the
// specified string.
func sliceContainsAny(slice []string, search string) bool {
	found := false

	for _, element := range slice {
		if search == element {
			found = true
			break
		}
	}

	return found
}

// isASCIIDigit returns true if the given string contains only the ascii digits 0123456789.
func isASCIIDigit(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c > unicode.MaxASCII || !unicode.IsDigit(rune(c)) {
			return false
		}
	}
	return true
}
