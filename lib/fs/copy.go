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
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

// CopyFile copies file from source to destination on the file system.
// Returns the number of bytes copied and an optional error.
func CopyFile(source string, destination string, bufferSize int64) (int64, error) {
	stat, statErr := os.Stat(source)
	if statErr != nil {
		return -1, statErr
	}

	if !stat.Mode().IsRegular() {
		return -1, errors.Errorf("(%s) is not a regular file", source)
	}

	sourceFile, openSrcErr := os.Open(source)
	if openSrcErr != nil {
		return -1, openSrcErr
	}
	defer func() {
		err := sourceFile.Close()
		if err != nil {
			log.Warnf("problem closing file (%s) after read: %v",
				source, err)
		}
	}()

	destFile, openDstErr := os.Create(destination)
	if openDstErr != nil {
		return -1, openDstErr
	}
	defer func() {
		err := destFile.Close()
		if err != nil {
			log.Warnf("problem closing file (%s) after write: %v",
				destination, err)
		}
	}()

	buffer := make([]byte, bufferSize)

	var totalBytesWritten int64 = 0

	for {
		bytesRead, readErr := sourceFile.Read(buffer)
		if readErr != nil && readErr != io.EOF {
			return -1, readErr
		}
		if bytesRead == 0 {
			break
		}

		bytesWritten, writeErr := destFile.Write(buffer[:bytesRead])
		if writeErr != nil {
			return -1, writeErr
		}

		totalBytesWritten = int64(bytesWritten) + totalBytesWritten
	}

	return totalBytesWritten, nil
}
