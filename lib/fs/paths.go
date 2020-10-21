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
	"path/filepath"
)

// PathSeparator is the OS dependent path separator character as a string.
const PathSeparator = string(os.PathSeparator)

// Useful permissions constants example provided via StackOverflow:
// https://stackoverflow.com/a/42718395/33611
// noinspection GoSnakeCaseUsage,GoUnusedConst
const (
	OS_READ        = 04
	OS_WRITE       = 02
	OS_EX          = 01
	OS_USER_SHIFT  = 6
	OS_GROUP_SHIFT = 3
	OS_OTH_SHIFT   = 0

	OS_USER_R   = OS_READ << OS_USER_SHIFT
	OS_USER_W   = OS_WRITE << OS_USER_SHIFT
	OS_USER_X   = OS_EX << OS_USER_SHIFT
	OS_USER_RW  = OS_USER_R | OS_USER_W
	OS_USER_RWX = OS_USER_RW | OS_USER_X

	OS_GROUP_R   = OS_READ << OS_GROUP_SHIFT
	OS_GROUP_W   = OS_WRITE << OS_GROUP_SHIFT
	OS_GROUP_X   = OS_EX << OS_GROUP_SHIFT
	OS_GROUP_RW  = OS_GROUP_R | OS_GROUP_W
	OS_GROUP_RWX = OS_GROUP_RW | OS_GROUP_X

	OS_OTH_R   = OS_READ << OS_OTH_SHIFT
	OS_OTH_W   = OS_WRITE << OS_OTH_SHIFT
	OS_OTH_X   = OS_EX << OS_OTH_SHIFT
	OS_OTH_RW  = OS_OTH_R | OS_OTH_W
	OS_OTH_RWX = OS_OTH_RW | OS_OTH_X

	OS_ALL_R   = OS_USER_R | OS_GROUP_R | OS_OTH_R
	OS_ALL_W   = OS_USER_W | OS_GROUP_W | OS_OTH_W
	OS_ALL_X   = OS_USER_X | OS_GROUP_X | OS_OTH_X
	OS_ALL_RW  = OS_ALL_R | OS_ALL_W
	OS_ALL_RWX = OS_ALL_RW | OS_GROUP_X
)

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
	// os.TempDir() generates a temporary directory path that has directory separators normalized.
	// Go on MacOS in some cases will generate paths with double directory separators and this
	// causes the path matching tests to fail. With this function, we can be assured that the
	// generated path is normalized.
	return filepath.Clean(os.TempDir() + PathSeparator + suffix)
}

// PathExistsAndIsFileOrDirectory checks to see if a given path exists and if it is a
// directory or regular file. If the check is successful this function will return nil.
// Otherwise, it will return an error indicating the problem with the path.
func PathExistsAndIsFileOrDirectory(path string) error {
	fileInfo, statErr := os.Lstat(path)

	if statErr != nil {
		return errors.Wrapf(statErr, "unable to stat path (%s)", path)
	}

	if !IsRegularFileOrDirectory(fileInfo) {
		return errors.Errorf("path (%s) is not a directory or regular file", path)
	}

	return nil
}

// IsRegularFileOrDirectory returns true if the referenced FileInfo object
// is a normal file or directory and not a symlink, device or other specialized
// file type.
func IsRegularFileOrDirectory(fileInfo os.FileInfo) bool {
	return fileInfo.IsDir() || fileInfo.Mode().IsRegular()
}
