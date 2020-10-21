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
	"github.com/iancoleman/strcase"
	"github.com/nginxinc/nginx-wrapper/app/config"
	"github.com/nginxinc/nginx-wrapper/lib/api"
	"github.com/nginxinc/nginx-wrapper/lib/fs"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gopkg.in/oleiade/reflections.v1"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

const uRWgR = fs.OS_USER_RW | fs.OS_GROUP_R
const logPerm = os.FileMode(uRWgR)

// InitLogger creates and returns a new logger instance based on
// the specified command line logger level flag.
func InitLogger(settings api.Settings) error {
	logConfig := config.SubViperConfig(settings, "log")

	// Apply defaults to log configuration because they were lost when
	// we extracted the sub element.
	for k, v := range config.LogDefaults {
		logConfig.SetDefault(k, v)
	}

	// normalize to lowercase log level
	logLevel := logConfig.GetString("level")
	level, logLevelErr := logrus.ParseLevel(logLevel)
	if logLevelErr != nil {
		return logLevelErr
	}
	logrus.SetLevel(level)

	destination, destErr := chooseDestination(logConfig.GetString("destination"))
	if destErr != nil {
		return destErr
	}
	logrus.SetOutput(destination)

	formatterOptions := logConfig.Sub("formatter_options")
	if formatterOptions == nil {
		formatterOptions = viper.New()
	}

	for k, v := range config.TextFormatterOptionsDefaults {
		formatterOptions.SetDefault(k, v)
	}

	var formatterName string

	if logConfig.IsSet("formatter_name") && formatterOptions != nil {
		formatterName = logConfig.GetString("formatter_name")
		formatter, formatConfigErr := chooseFormatter(formatterName, formatterOptions)
		if formatConfigErr != nil {
			return formatConfigErr
		}

		logrus.SetFormatter(formatter)
	} else {
		formatterName = ""
	}

	initLogrusDriver(strings.ToLower(formatterName) == "textformatter")

	return nil
}

func chooseDestination(destination string) (io.Writer, error) {
	lowercase := strings.ToLower(destination)
	var writer io.Writer

	switch lowercase {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		file, err := os.OpenFile(destination,
			os.O_WRONLY|os.O_CREATE|os.O_APPEND,
			logPerm)

		if err != nil {
			return nil, errors.Wrapf(err,
				"unable to open location (%s) for writing logs",
				destination)
		}

		writer = file
	}

	return writer, nil
}

func chooseFormatter(formatter string, options *viper.Viper) (logrus.Formatter, error) {
	lowercase := strings.ToLower(formatter)

	var f logrus.Formatter

	switch lowercase {
	case "textformatter":
		tf := &logrus.TextFormatter{}
		err := assignTextFormatterFields(tf, options)
		if err != nil {
			return nil, err
		}
		f = tf
	case "jsonformatter":
		jf := &logrus.JSONFormatter{}
		err := assignJSONFormatterFields(jf, options)
		if err != nil {
			return nil, err
		}
		f = jf
	default:
		return nil, errors.Errorf("can't assign unknown formatter (%s)",
			formatter)
	}

	return f, nil
}

func assignJSONFormatterFields(formatter *logrus.JSONFormatter, options *viper.Viper) error {
	fields, fieldsErr := reflections.Fields(formatter)

	if fieldsErr != nil {
		return fieldsErr
	}

	// Alias all of the options fields to also accept title case so that we
	// have parity with the struct's case conventions
	for _, titleCaseName := range fields {
		snakeName := strcase.ToSnake(titleCaseName)
		options.RegisterAlias(titleCaseName, snakeName)
	}

	if options.IsSet("pretty_print") {
		formatter.PrettyPrint = options.GetBool("pretty_print")
	}
	if options.IsSet("disable_timestamp") {
		formatter.DisableTimestamp = options.GetBool("disable_timestamp")
	}
	if options.IsSet("timestamp_format") {
		formatter.TimestampFormat = options.GetString("timestamp_format")
	}
	if options.IsSet("data_key") {
		formatter.DataKey = options.GetString("data_key")
	}

	// The following fields aren't implemented:
	// FieldMap
	// CallerPrettyfier: nil,

	return nil
}

func assignTextFormatterFields(formatter *logrus.TextFormatter, options *viper.Viper) error {
	fields, fieldsErr := reflections.Fields(formatter)

	if fieldsErr != nil {
		return fieldsErr
	}

	// Alias all of the options fields to also accept title case so that we
	// have parity with the struct's case conventions
	for _, titleCaseName := range fields {
		snakeName := strcase.ToSnake(titleCaseName)
		options.RegisterAlias(titleCaseName, snakeName)
	}

	if options.IsSet("force_colors") {
		formatter.ForceColors = options.GetBool("force_colors")
	}
	if options.IsSet("disable_colors") {
		formatter.DisableColors = options.GetBool("disable_colors")
	}
	if options.IsSet("force_quote") {
		formatter.ForceQuote = options.GetBool("force_quote")
	}
	if options.IsSet("environment_override_colors") {
		formatter.EnvironmentOverrideColors = options.GetBool("environment_override_colors")
	}
	if options.IsSet("disable_timestamp") {
		formatter.DisableTimestamp = options.GetBool("disable_timestamp")
	}
	if options.IsSet("full_timestamp") {
		formatter.FullTimestamp = options.GetBool("full_timestamp")
	}
	if options.IsSet("disable_sorting") {
		formatter.DisableSorting = options.GetBool("disable_sorting")
	}
	if options.IsSet("disable_level_truncation") {
		formatter.DisableLevelTruncation = options.GetBool("disable_level_truncation")
	}
	if options.IsSet("pad_level_text") {
		formatter.PadLevelText = options.GetBool("pad_level_text")
	}
	if options.IsSet("quote_empty_fields") {
		formatter.QuoteEmptyFields = options.GetBool("quote_empty_fields")
	}
	if options.IsSet("timestamp_format") {
		formatter.TimestampFormat = options.GetString("timestamp_format")
	}

	// The following fields aren't implemented:
	// SortingFunc
	// FieldMap
	// CallerPrettyfier

	return nil
}
