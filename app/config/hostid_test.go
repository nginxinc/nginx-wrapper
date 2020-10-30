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

package config

import (
	"github.com/pkg/errors"
	"net"
	"testing"
)

func TestHostIDDefaultGenerators(t *testing.T) {
	hostId := hostID(defaultHostIDGenerators)
	if hostId == "" {
		t.Error("blank machine id returned")
	}
}

func TestHostIDGeneratorFallback(t *testing.T) {
	badGenerator := func() (string, error) {
		return "", errors.New("this generator will always error")
	}

	goodGenerator := func() (string, error) {
		return "00000000000000000000000000000001", nil
	}

	generators := []func() (string, error){
		badGenerator,
		goodGenerator,
	}

	hostId := hostID(generators)
	if hostId == "" {
		t.Error("blank machine id returned")
	}
}

func TestHostIDGeneratorNoValidIds(t *testing.T) {
	var generators []func() (string, error)

	hostId := hostID(generators)
	if hostId != "" {
		t.Error("blank machine id expected")
	}
}

func TestParseMachineIDFromMACAddresses(t *testing.T) {
	interfaces, interfaceListErr := net.Interfaces()

	if len(interfaces) > 0 && interfaceListErr != nil {
		machineId, err := parseMachineIDFromMACAddresses()
		if err != nil {
			t.Error(err)
		}

		if machineId == "" {
			t.Error("blank machine id returned")
		}
	}
}

func TestRandomMachineID(t *testing.T) {
	machineID, machineIDErr := randomMachineID()
	if machineIDErr != nil {
		t.Error(machineIDErr)
	}

	if machineID == "" {
		t.Error("unable to generate a random machine id")
	}
}
