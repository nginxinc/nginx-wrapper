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
	"github.com/nginxinc/nginx-wrapper/lib/osenv"
	"github.com/spf13/cobra"
	"os"
)

// AppVersionInfo is a data structure that represents the version information
// we want to display to users.
type AppVersionInfo struct {
	AppVersion    string
	GitCommitHash string
	UTCBuildTime  string
}

func (avi AppVersionInfo) String() string {
	return fmt.Sprintf("nginx-wrapper %s (%s) %s", avi.AppVersion,
		avi.GitCommitHash, avi.UTCBuildTime)
}

// Version contains the build details for nginx-wrapper as generated
// from the Makefile.
var Version AppVersionInfo

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the nginx-wrapper version",
	Run:   appVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func appVersion(*cobra.Command, []string) {
	fmt.Print(Version)
	fmt.Print(osenv.LineBreak)
	os.Exit(0)
}
