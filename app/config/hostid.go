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
	// nolint: gosec
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"github.com/denisbrodbeck/machineid"
	slog "github.com/go-eden/slf4go"
	"github.com/pkg/errors"
	"net"
)

// hostIDGenerators contains the list of functions to use to generate a host id.
var defaultHostIDGenerators = []func() (string, error){
	machineid.ID,
	parseMachineIDFromMACAddresses,
	randomMachineID,
}

// hostID finds a unique identifier for the running host.
func hostID(generators []func() (string, error)) string {
	for _, generator := range generators {
		machineID, machineIDErr := generator()
		if machineIDErr == nil && machineID != "" {
			return machineID
		} else {
			slog.NewLogger("host-id").Warn(machineIDErr)
		}
	}

	// Worst case, we give up and return an empty string. The host id isn't critical
	// for the functioning of the overall system, so it is ok that it blank.
	slog.NewLogger("host-id").Error("unable to generate a host id")

	return ""
}

// parseMachineIDFromMACAddresses generates a host id based on the MAC addresses on the host.
func parseMachineIDFromMACAddresses() (string, error) {
	interfaces, interfaceListErr := net.Interfaces()

	if len(interfaces) < 1 {
		return "", errors.New("unable to generate a machine id because there are no mac addresses available")
	}

	if interfaceListErr != nil {
		return "", errors.Wrap(interfaceListErr, "unable to list network interfaces")
	}

	// nolint: gosec
	macAddressHash := md5.New()

	for _, interfaceInfo := range interfaces {
		macAddress := interfaceInfo.HardwareAddr
		_, writeErr := macAddressHash.Write([]byte(macAddress.String()))
		if writeErr != nil {
			return "", errors.Wrapf(writeErr, "unable to write mac address (%s) to hash", macAddress)
		}
	}

	var hashSum = macAddressHash.Sum(nil)
	hexString := hex.EncodeToString(hashSum)

	return hexString, nil
}

// randomMachineId generates a purely random host id.
func randomMachineID() (string, error) {
	var randomBytes = make([]byte, 16)
	_, randErr := rand.Read(randomBytes)
	if randErr != nil {
		return "", errors.Wrapf(randErr, "unable to randomly generate host id")
	}

	machineID := hex.EncodeToString(randomBytes)

	return machineID, nil
}
