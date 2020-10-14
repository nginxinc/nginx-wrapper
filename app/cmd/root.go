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
	"github.com/nginxinc/nginx-wrapper/lib/osenv"
	"github.com/spf13/cobra"
	"os"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "nginx-wrapper",
	Short: "A pluggable process wrapper for NGINX",
	Long: `NGINX Wrapper is a process wrapper that monitors NGINX for 
(start, reload, and exit) events, provides a templating framework for 
NGINX conf files and allows for plugins that extend its functionality.`,
	Run: run,
}

func run(cmd *cobra.Command, args []string) {
	// this closure must be present in order for OnInitialize to be called
	if len(args) == 0 {
		_ = cmd.Help()
		os.Exit(0)
	}
}

// Execute the root command thereby invoking any user specified commands.
func Execute() {
	rootCmd.Version = Version.AppVersion

	err := rootCmd.Execute()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v%s", err, osenv.LineBreak)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", config.DefaultConfigPath,
		"path to configuration file")
}
