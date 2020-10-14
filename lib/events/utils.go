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

import (
	"strconv"
	"strings"
	"unicode"
)

func reverse(runes []rune, limit int) string {
	output := make([]rune, limit)

	for i := 0; i < limit && i < len(runes); i++ {
		output[limit-1-i] = runes[i]
	}

	return string(output)
}

func parsePidAndTid(msg string) (int64, int64, error) {
	semicolonPos := strings.Index(msg, ":")
	if semicolonPos == -1 {
		err := newNginxLogParserError(
			"unable to find expected semicolon in message: %s", msg)
		return -1, -1, err
	}

	hashPos := strings.Index(msg, "#")
	if hashPos == -1 {
		err := newNginxLogParserError(
			"unable to find expected hash in message: %s", msg)
		return -1, -1, err
	}

	pidStr := msg[0:hashPos]

	if hashPos >= semicolonPos {
		err := newNginxLogParserError(
			"hash is beyond semicolon position - can't parse: %s", msg)
		return -1, -1, err
	}

	tidStr := msg[hashPos+1 : semicolonPos]

	pid, pidParseErr := strconv.ParseInt(pidStr, 10, 64)
	if pidParseErr != nil {
		return -1, -1, pidParseErr
	}

	tid, tidParseErr := strconv.ParseInt(tidStr, 10, 64)
	if tidParseErr != nil {
		return -1, -1, tidParseErr
	}

	return pid, tid, nil
}

func extractPidFromStartWorkerProcess(text string) int64 {
	matchRunes := []rune("start worker process ")
	maxLen := len(matchRunes) + maxPidWidth

	if text == "" {
		return -1
	}

	var runes []rune

	if len(text)-maxLen < 0 {
		runes = []rune(text)
	} else {
		runes = []rune(text[len(text)-maxLen:])
	}

	reversedPidDigits := make([]rune, maxPidWidth)
	reversedPidDigitsWritten := 0
	startPos := len(runes) - 1
	lastCharWasDigit := false
	matchPos := len(matchRunes)

	for i := startPos; i >= 0; i-- {
		c := runes[i]

		// trim whitespace off of the end
		if i == startPos && unicode.IsSpace(c) {
			startPos = i - 1
			continue
		}

		if i == startPos || lastCharWasDigit {
			if unicode.IsDigit(c) {
				// we have found too many digits
				if reversedPidDigitsWritten >= maxPidWidth {
					return -1
				}

				reversedPidDigits[reversedPidDigitsWritten] = c
				reversedPidDigitsWritten++

				lastCharWasDigit = true
				continue
				// the last character in the string must be a digit
			} else if i == startPos {
				return -1
				// this wasn't a digit so, we flip the bit
			} else if lastCharWasDigit {
				lastCharWasDigit = false
			}
		}

		// if we are here, the current character is not a digit
		matchPos--

		// we have matched the string but there is still preceding text
		if matchPos < 0 {
			break
		}

		m := matchRunes[matchPos]
		if c != m {
			return -1
		}
	}

	stringPid := reverse(reversedPidDigits, reversedPidDigitsWritten)
	pid, err := strconv.ParseInt(stringPid, 10, 64)

	if err != nil {
		log.Fatalf("couldn't parse pid (%s) as integer: %v",
			stringPid, err)
	}

	return pid
}
