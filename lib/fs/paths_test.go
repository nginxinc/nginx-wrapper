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

package fs

import (
	"testing"
)

func TestCantFindInPath(t *testing.T) {
	_, err := FindInPath("something345345")

	if err == nil {
		t.Error("no error thrown for missing binary")
	}
}

func TestCanFindInPath(t *testing.T) {
	path, err := FindInPath("go")

	if err != nil {
		t.Error(err)
	}

	if path == "" {
		t.Error("empty path returned")
	}
}
