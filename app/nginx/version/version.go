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

package version

import (
	"bufio"
	"fmt"
	"github.com/elliotchance/orderedmap"
	"github.com/nginxinc/nginx-wrapper/lib/osenv"
	"github.com/pkg/errors"
	"os/exec"
	"regexp"
	"strings"
)

const versionPrefix = "nginx version: "
const configArgPrefix = "configure arguments:"

var versionMatchPattern = regexp.MustCompile(
	`^.*:?\s*(nginx/(\d+\.\d+\.\d+)\s*(.*))`)
var configArgSplitPatter = regexp.MustCompile(`\s+--`)

// NginxVersion contains details parsed from the "nginx -v" command.
type NginxVersion struct {
	Full             string
	Version          string
	Detail           string
	IsPlus           bool
	AdditionalDetail []string
	ConfigureArgs    *orderedmap.OrderedMap
}

func (nv NginxVersion) String() string {
	return nv.Full
}

// Info contains a dump of all version information.
func (nv NginxVersion) Info() string {
	var builder strings.Builder

	builder.WriteString(nv.Full)
	builder.WriteRune(osenv.LineBreakRune)

	for i := 0; i < len(nv.AdditionalDetail); i++ {
		builder.WriteString(nv.AdditionalDetail[i])
		builder.WriteRune(osenv.LineBreakRune)
	}

	if nv.ConfigureArgs.Len() > 0 {
		builder.WriteString(configArgPrefix)

		for el := nv.ConfigureArgs.Front(); el != nil; el = el.Next() {
			key := fmt.Sprintf("%v", el.Key)
			val := fmt.Sprintf("%v", el.Value)
			builder.WriteString(" --")
			builder.WriteString(key)

			if val != "" {
				builder.WriteString("=")
				builder.WriteString(val)
			}
		}
	}

	return builder.String()
}

// ReadNginxVersion parses the output of the "nginx -V" command into
// a consumable data structure.
func ReadNginxVersion(nginxBinPath string) (NginxVersion, error) {
	cmd := exec.Command(nginxBinPath, "-V")
	output, cmdErr := cmd.CombinedOutput()

	if cmdErr != nil {
		return NginxVersion{}, cmdErr
	}

	// we load all of the output into memory because it shouldn't be
	// a lot of data and has a clear bounds
	versionOutput := string(output)

	scanner := bufio.NewScanner(strings.NewReader(versionOutput))

	if !scanner.Scan() {
		return NginxVersion{}, errors.Errorf(
			"no output from NGINX executable (%s)", nginxBinPath)
	}

	firstLine := scanner.Text()
	version, versionErr := parseVersion(firstLine)
	if versionErr != nil {
		return NginxVersion{}, errors.Wrapf(versionErr,
			"unable to parse first version line from NGINX executable (%s)",
			nginxBinPath)
	}

	var details []string
	var configureArgs *orderedmap.OrderedMap

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, configArgPrefix) {
			substr := line[len(configArgPrefix):]

			args := parseConfigureArgs(substr)
			configureArgs = args
		} else {
			details = append(details, line)
		}
	}

	version.AdditionalDetail = details
	version.ConfigureArgs = configureArgs

	return version, nil
}

func parseConfigureArgs(line string) *orderedmap.OrderedMap {
	args := orderedmap.NewOrderedMap()

	chunked := configArgSplitPatter.Split(line, -1)

	for i := 0; i < len(chunked); i++ {
		s := chunked[i]

		if s == "" {
			continue
		}

		equalsPos := strings.Index(s, "=")
		var key string
		var val string

		if equalsPos > 0 {
			key = s[0:equalsPos]
			val = s[equalsPos+1:]
		} else {
			key = s
			val = ""
		}

		args.Set(key, val)
	}

	return args
}

func parseVersion(line string) (NginxVersion, error) {
	var substring string

	if strings.HasPrefix(line, versionPrefix) {
		substring = line[len(versionPrefix):]
	} else {
		substring = line
	}

	matches := versionMatchPattern.FindStringSubmatch(substring)
	expectedMatches := 4
	if len(matches) != expectedMatches {
		return NginxVersion{}, errors.Errorf(
			"can't extract invalid version text: %s", line)
	}

	detail := strings.Trim(matches[3], "()")

	return NginxVersion{
		Full:    matches[1],
		Version: matches[2],
		Detail:  detail,
		IsPlus:  strings.HasPrefix(detail, "nginx-plus"),
	}, nil
}
