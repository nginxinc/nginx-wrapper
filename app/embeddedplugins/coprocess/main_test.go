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
	"fmt"
	"github.com/nginxinc/nginx-wrapper/lib/api"
	"github.com/nginxinc/nginx-wrapper/lib/events"
	"github.com/spf13/viper"
	"os"
	"sync"
	"testing"
	"time"
)

func TestCoprocessCanBeEndedWithTERM(t *testing.T) {
	exec := []string{"bash", "ignores_int_signal.sh"}
	err := runTestCoProcess(exec)
	if err != nil {
		t.Error(err)
	}
}

func TestCoprocessCanBeEndedWithINT(t *testing.T) {
	exec := []string{"bash", "ignores_term_signal.sh"}
	err := runTestCoProcess(exec)
	if err != nil {
		t.Error(err)
	}
}

func TestCoprocessCanBeKilled(t *testing.T) {
	exec := []string{"bash", "ignores_int_term_signals.sh"}
	err := runTestCoProcess(exec)
	if err != nil {
		t.Error(err)
	}
}

func runTestCoProcess(exec []string) error {
	events.Init()

	fmt.Println(os.Getwd())

	vconfig := viper.New()
	vconfig.Set("coprocess.consul.name", "test-process")
	vconfig.Set("coprocess.consul.exec", exec)
	vconfig.Set("coprocess.consul.restarts", "unlimited")
	vconfig.Set("coprocess.consul.background", true)
	vconfig.Set("coprocess.consul.exec_event", "pre-start")
	vconfig.Set("coprocess.consul.stop_event", "exit")
	context := api.PluginStartupContext{
		Settings: vconfig,
	}

	err := Start(context)
	if err != nil {
		return err
	}

	waitGroup := sync.WaitGroup{}

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		time.Sleep(3 * time.Second)
		events.GlobalEvents.NginxExit.Trigger(map[string]interface{}{})
	}()

	events.GlobalEvents.NginxPreStart.Trigger(map[string]interface{}{})

	waitGroup.Wait()

	return nil
}
