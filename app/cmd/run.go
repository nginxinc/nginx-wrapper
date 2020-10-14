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
	slog "github.com/go-eden/slf4go"
	"github.com/nginxinc/nginx-wrapper/app/initapp"
	"github.com/nginxinc/nginx-wrapper/app/nginx"
	"github.com/nginxinc/nginx-wrapper/lib/fs"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run NGINX in a process wrapper",
	Run:   startWrapper,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func startWrapper(*cobra.Command, []string) {
	vconfig := initapp.Init(true, configPath)

	log := slog.NewLogger("run")

	pm := nginx.NewProcessMonitor()
	pm.Increment()

	// Start up NGINX
	go func() {
		defer pm.Done()

		configPath := vconfig.GetString("conf_path") + fs.PathSeparator +
			"nginx.conf"
		runPath := vconfig.GetString("run_path")
		nginxBinPath := vconfig.GetString("nginx_binary")
		startErr := nginx.Start(nginxBinPath, runPath, configPath, pm)
		if startErr != nil {
			log.Fatalf("error starting NGINX: %+v", startErr)
		}
	}()

	pm.Wait()
}
