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
	"reflect"
	"strings"
	"testing"
)

func TestInterpolateAllNoMatches(t *testing.T) {
	original := []string{"dog", "cat", "rabbit", "mouse"}
	substitutions := []string{"${donkey}", "burro", "${cow}", "vaca"}
	replacer := strings.NewReplacer(substitutions...)
	actual := interpolateAll(original, replacer)

	if !reflect.DeepEqual(original, actual) {
		t.Errorf("Unexpected value in interpolated array values:\n"+
			"expected: %v\n"+
			"actual  : %v", original, actual)
	}
}

func TestInterpolateOneMatch(t *testing.T) {
	original := []string{"dog friend", "${donkey} friend", "rabbit friend", "mouse friend"}
	substitutions := []string{"${donkey}", "burro", "${cow}", "vaca"}
	expected := []string{"dog friend", "burro friend", "rabbit friend", "mouse friend"}
	replacer := strings.NewReplacer(substitutions...)
	actual := interpolateAll(original, replacer)

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Unexpected value in interpolated array values:\n"+
			"expected: %v\n"+
			"actual  : %v", original, actual)
	}
}

func TestInterpolateTwoMatches(t *testing.T) {
	original := []string{"dog friend", "${donkey} friend", "rabbit friend", "${cow} friend", "mouse friend"}
	substitutions := []string{"${donkey}", "burro", "${cow}", "vaca"}
	expected := []string{"dog friend", "burro friend", "rabbit friend", "vaca friend", "mouse friend"}
	replacer := strings.NewReplacer(substitutions...)
	actual := interpolateAll(original, replacer)

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Unexpected value in interpolated array values:\n"+
			"expected: %v\n"+
			"actual  : %v", original, actual)
	}
}

func TestContainsFound(t *testing.T) {
	values := []string{"dog", "cat", "rabbit", "mouse"}
	found := sliceContainsAny(values, "rabbit")

	if !found {
		t.Error("expected value not found")
	}
}

func TestContainsNotFound(t *testing.T) {
	values := []string{"dog", "cat", "rabbit", "mouse"}
	found := sliceContainsAny(values, "kangaroo")

	if found {
		t.Error("unexpected value found")
	}
}

func TestIsAsciiDigitAllDigits(t *testing.T) {
	value := "123"
	actual := isASCIIDigit(value)
	if !actual {
		t.Errorf("failed to test properly for digits value: %s", value)
	}
}

func TestIsAsciiDigitNotAllDigits(t *testing.T) {
	value := "-123"
	actual := isASCIIDigit(value)
	if actual {
		t.Errorf("failed to test properly for digits value: %s", value)
	}
}

func TestIsAsciiDigitUnicodeDigitShouldFail(t *testing.T) {
	value := "１２３"
	actual := isASCIIDigit(value)
	if actual {
		t.Errorf("failed to test properly for digits value: %s", value)
	}
}
