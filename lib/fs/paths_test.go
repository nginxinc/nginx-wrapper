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
	"os"
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

func TestPathExistsAndIsFileOrDirectoryWithDirectory(t *testing.T) {
	tempDir := makeTestDirectory(t)
	defer os.RemoveAll(tempDir)
	err := PathExistsAndIsFileOrDirectory(tempDir)

	if err != nil {
		t.Error(err)
	}
}

func TestPathExistsAndIsFileOrDirectoryWithFile(t *testing.T) {
	tempDir := makeTestDirectory(t)
	defer os.RemoveAll(tempDir)

	filePath := tempDir + PathSeparator + "original"
	_, createErr := os.Create(filePath)
	if createErr != nil {
		t.Error(createErr)
	}

	err := PathExistsAndIsFileOrDirectory(filePath)

	if err != nil {
		t.Error(err)
	}
}

func TestPathExistsAndIsFileOrDirectoryWithBlankPath(t *testing.T) {
	err := PathExistsAndIsFileOrDirectory("")

	if err == nil {
		t.Error("expected error not thrown")
	}
}

func TestPathExistsAndIsFileOrDirectoryWithSymlinkFilePath(t *testing.T) {
	tempDir := makeTestDirectory(t)
	defer os.RemoveAll(tempDir)

	sourcePath := tempDir + PathSeparator + "original"
	_, createErr := os.Create(sourcePath)
	if createErr != nil {
		t.Error(createErr)
	}
	symLinkPath := tempDir + PathSeparator + "symlink"
	symLinkErr := os.Symlink(sourcePath, symLinkPath)
	if symLinkErr != nil {
		t.Error(symLinkErr)
	}

	err := PathExistsAndIsFileOrDirectory(symLinkPath)

	if err == nil {
		t.Error("expected error not thrown")
	}
}

func TestPathExistsAndIsFileOrDirectoryWithSymlinkDirectoryPath(t *testing.T) {
	tempDir := makeTestDirectory(t)
	defer os.RemoveAll(tempDir)

	sourcePath := tempDir + PathSeparator + "original"
	mkdirErr := os.Mkdir(sourcePath, os.ModeDir|(OS_USER_RWX|OS_ALL_R))
	if mkdirErr != nil {
		t.Error(mkdirErr)
	}
	symLinkPath := tempDir + PathSeparator + "symlink"
	symLinkErr := os.Symlink(sourcePath, symLinkPath)
	if symLinkErr != nil {
		t.Error(symLinkErr)
	}

	err := PathExistsAndIsFileOrDirectory(symLinkPath)

	if err == nil {
		t.Error("expected error not thrown")
	}
}

func TestPathExistsAndIsFileOrDirectoryWithNonRegularFile(t *testing.T) {
	err := PathExistsAndIsFileOrDirectory("/dev/null")

	if err == nil {
		t.Error("expected error not thrown")
	}
}

func makeTestDirectory(t *testing.T) string {
	tempDir := TempDirectoryPath("nginx-wrapper-lib-test" + PathSeparator + t.Name())
	mkdirErr := os.MkdirAll(tempDir, os.ModeDir|(OS_USER_RWX|OS_ALL_R))
	if mkdirErr != nil {
		t.Error(mkdirErr)
	}

	return tempDir
}
