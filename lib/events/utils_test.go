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

import "testing"

func TestReverseLengthIsLimit(t *testing.T) {
	input := "87654321"
	expected := "12345678"
	limit := 8

	actual := reverse([]rune(input), limit)
	if actual != expected {
		t.Errorf("actual: %s expected: %s", actual, expected)
	}
}

func TestReverseLengthIsGreaterThanLimit(t *testing.T) {
	input := make([]rune, 8)
	input[0] = '4'
	input[1] = '3'
	input[2] = '2'
	input[3] = '1'

	expected := "1234"
	limit := 4

	actual := reverse(input, limit)
	if actual != expected {
		t.Errorf("actual: %s expected: %s", actual, expected)
	}
}

func TestParsePidAndTidValidMsg(t *testing.T) {
	msg := "21247#11546: epoll_wait() failed (4: Interrupted system call)"
	pid, tid, err := parsePidAndTid(msg)

	if err != nil {
		t.Error(err)
	}

	if pid != 21247 {
		t.Errorf("pid din't match 21247 and was %d", pid)
	}

	if tid != 11546 {
		t.Errorf("tid din't match 21247 and was %d", tid)
	}
}

func TestParsePidAndTidNoSemicolon(t *testing.T) {
	msg := "21247#11546 epoll_wait() failed (4: Interrupted system call)"
	_, _, err := parsePidAndTid(msg)

	if err == nil {
		t.Error("expected error not thrown")
	}
}

func TestParsePidAndTidNoHash(t *testing.T) {
	msg := "21247-11546: epoll_wait() failed (4: Interrupted system call)"
	_, _, err := parsePidAndTid(msg)

	if err == nil {
		t.Error("expected error not thrown")
	}
}

func TestParsePidAndTidHashInWrongPlace(t *testing.T) {
	msg := "21247-11546: epoll_wait() failed (#4: Interrupted system call)"
	_, _, err := parsePidAndTid(msg)

	if err == nil {
		t.Error("expected error not thrown")
	}
}

func TestExtractPidFromStartWorkerProcessEmptyString(t *testing.T) {
	if extractPidFromStartWorkerProcess("") != -1 {
		t.Error("empty string doesn't have digits")
	}
}

func TestExtractPidFromStartWorkerProcessNoPid(t *testing.T) {
	if extractPidFromStartWorkerProcess("23303#23303: start worker process ") != -1 {
		t.Error("shouldn't match because there is no pid")
	}
}

func TestExtractPidFromStartWorkerProcessNoLastSpace(t *testing.T) {
	if extractPidFromStartWorkerProcess("23303#23303: start worker process_") != -1 {
		t.Error("shouldn't match because there is no space")
	}
}

func TestExtractPidFromStartWorkerProcessSingleTrailingSpace(t *testing.T) {
	if extractPidFromStartWorkerProcess("23303#23303: start worker process 2 ") != 2 {
		t.Error("failed to match valid text")
	}
}

func TestExtractPidFromStartWorkerProcessManyTrailingSpaces(t *testing.T) {
	if extractPidFromStartWorkerProcess("23303#23303: start worker process 22342    ") != 22342 {
		t.Error("failed to match valid text")
	}
}

func TestExtractPidFromStartWorkerProcessNoLeadingText(t *testing.T) {
	if extractPidFromStartWorkerProcess("start worker process 23325") != 23325 {
		t.Error("failed to match valid text")
	}
}

func TestExtractPidFromStartWorkerProcessLongLeadingText(t *testing.T) {
	if extractPidFromStartWorkerProcess("1234567890 1234567890 1234567890 1234567890: start worker process 23325") != 23325 {
		t.Error("failed to match valid text")
	}
}

func TestExtractPidFromStartWorkerProcessPidTooBig(t *testing.T) {
	if extractPidFromStartWorkerProcess("1234567890#1234567890: start worker process 12345678901234567890") != -1 {
		t.Error("shouldn't match if pid is abnormally large")
	}
}

func TestExtractPidFromStartWorkerProcessOneDigits(t *testing.T) {
	if extractPidFromStartWorkerProcess("5#5: start worker process 5") != 5 {
		t.Error("failed to match valid text")
	}
}

func TestExtractPidFromStartWorkerProcessFiveDigits(t *testing.T) {
	if extractPidFromStartWorkerProcess("23303#23303: start worker process 23325") != 23325 {
		t.Error("failed to match valid text")
	}
}

func TestExtractPidFromStartWorkerProcessEightDigits(t *testing.T) {
	if extractPidFromStartWorkerProcess("12345678#12345678: start worker process 12345678") != 12345678 {
		t.Error("failed to match valid text")
	}
}

func TestExtractPidFromStartWorkerProcessNineteenDigits(t *testing.T) {
	if extractPidFromStartWorkerProcess("12345678#12345678: start worker process 1234567890123456789") != 1234567890123456789 {
		t.Error("failed to match valid text")
	}
}
