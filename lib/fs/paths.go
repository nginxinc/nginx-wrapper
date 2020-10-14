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
	"github.com/pkg/errors"
	"os"
	"os/exec"
)

// PathSeparator is the OS dependent path separator character as a string.
const PathSeparator = string(os.PathSeparator)

// FindInPath searches the PATH and the local directory structure for a file.
func FindInPath(fileName string) (string, error) {
	path, lookPathErr := exec.LookPath(fileName)

	if lookPathErr == nil {
		return path, nil
	}

	return "", errors.Wrapf(lookPathErr,
		"unable to find valid file in path")
}

// TempDirectoryPath returns a path under the system temporary directory.
func TempDirectoryPath(suffix string) string {
	return os.TempDir() + PathSeparator + suffix
}
