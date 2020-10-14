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
	"github.com/nginxinc/nginx-wrapper/lib/api"
	"github.com/spf13/viper"
	"reflect"
	"testing"
)

func TestInitErrors(t *testing.T) {
	settings := api.Settings(viper.New())
	_, err := New("", []string{}, []string{}, "", "unknown", false, "something",
		"something", settings)

	if err == nil {
		t.Error("no error returned for invalid init input values")
	} else {
		errMessageCount := len(err.InitFailureMessages)
		if errMessageCount != 5 {
			t.Errorf("unexpected number of error messages received (%d)", errMessageCount)
		}
	}
}

func TestInterpolatedSettings(t *testing.T) {
	vconfig := viper.New()
	vconfig.Set("host_id", "a-valid-host-id")
	vconfig.Set("vcpu_count", 3)
	env := map[string]string{
		"var1": "this",
		"var2": "that",
	}
	vconfig.Set("env", env)
	settings := api.Settings(vconfig)

	exec := []string{"run", "${vcpu_count}", "${var1}=${unknown}"}
	stopExec := []string{"stop", "hammertime", "${host_id}"}
	coprocess, err := New("unittest", exec, stopExec, "testy", string(Never), false,
		"start", "exit", settings)

	if err != nil {
		t.Error(err)
	}

	expectedExec := []string{"run", "3", "this=${unknown}"}
	expectedStopExec := []string{"stop", "hammertime", "a-valid-host-id"}

	if !reflect.DeepEqual(coprocess.Exec, expectedExec) {
		t.Errorf("Unexpected value in interpolated array values:\n"+
			"expected: %v\n"+
			"actual  : %v", expectedExec, exec)
	}

	if !reflect.DeepEqual(coprocess.StopExec, expectedStopExec) {
		t.Errorf("Unexpected value in interpolated array values:\n"+
			"expected: %v\n"+
			"actual  : %v", expectedStopExec, stopExec)
	}
}
