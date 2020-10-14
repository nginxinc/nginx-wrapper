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

package cmd

import (
	"fmt"
	"github.com/nginxinc/nginx-wrapper/app/config"
	"github.com/nginxinc/nginx-wrapper/app/initapp"
	"github.com/nginxinc/nginx-wrapper/lib/osenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"sort"
	"strconv"
	"strings"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Display runtime configuration",
	Run:   debug,
}

func init() {
	rootCmd.AddCommand(debugCmd)
}

func debug(*cobra.Command, []string) {
	vconfig := initapp.Init(false, configPath)

	orderedKeys := config.Keys()
	elements := config.AllKnownElements(vconfig, ".")

	maxKeySize := 1
	for _, k := range orderedKeys {
		if len(k) > maxKeySize {
			maxKeySize = len(k)
		}
	}

	format := "%" + strconv.Itoa(maxKeySize) + "s: %v" + osenv.LineBreak

	for _, k := range orderedKeys {
		v := elements[k]
		fmt.Printf(format, k, v)
	}

	unknownKeys := findUnknownKeys(vconfig, orderedKeys)
	if len(unknownKeys) > 0 {
		fmt.Println(osenv.LineBreak + "The following configuration settings are unknown:")
		for _, k := range unknownKeys {
			v := elements[k]
			fmt.Printf(format, k, v)
		}
	}
}

func findUnknownKeys(vconfig *viper.Viper, expectedKeys []string) []string {
	const subSubKeyCount = 2
	var unknownKeys []string
	var loadedKeys []string

	// Skip sub-subkeys (two dots or more) because we don't validate the expected values
	// associated with those keys. For example, in the sub-subkey for log.formatter_options.*,
	// we allow any sub-subkeys so that users can flexibly configure the logger with options
	// that are relevant to the logger implementation.
	for _, key := range vconfig.AllKeys() {
		if strings.Count(key, ".") < subSubKeyCount {
			loadedKeys = append(loadedKeys, key)
		}
	}

	sort.Strings(loadedKeys)

	for _, key := range loadedKeys {
		indexToInsert := sort.SearchStrings(expectedKeys, key)
		var found bool
		if indexToInsert == len(expectedKeys) {
			found = false
		} else {
			found = expectedKeys[indexToInsert] == key
		}

		if !found {
			unknownKeys = append(unknownKeys, key)
		}
	}

	return unknownKeys
}
