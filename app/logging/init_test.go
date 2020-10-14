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

package logging

import (
	"github.com/spf13/viper"
	"testing"
)

func TestCanInitLoggerWithValidLevel(t *testing.T) {
	config := viper.New()
	config.Set("log.level", "TRACE")

	err := InitLogger(config)

	if err != nil {
		t.Error(err)
	}
}

func TestCantInitLoggerWithInvalidLevel(t *testing.T) {
	config := viper.New()
	config.Set("log.level", "INVALID")

	err := InitLogger(config)

	if err == nil {
		t.Error("No error thrown for invalid level")
	}
}
