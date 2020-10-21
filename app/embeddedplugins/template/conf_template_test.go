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

package template

import (
	"bytes"
	"fmt"
	"github.com/nginxinc/nginx-wrapper/app/config"
	"github.com/nginxinc/nginx-wrapper/lib/api"
	"github.com/nginxinc/nginx-wrapper/lib/fs"
	"github.com/nginxinc/nginx-wrapper/lib/osenv"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

func TestDiscoverTemplateFilesWithSingleFile(t *testing.T) {
	cfgTemplate := NewTemplate()
	vconfig := viper.New()

	tempDir := tempDir("nginx-wrapper-unit-tests")
	mkdir(tempDir, t)
	confTemplateDir := tempDir + fs.PathSeparator + "conf"
	mkdir(confTemplateDir, t)
	confTemplatePath := confTemplateDir + fs.PathSeparator + "nginx.conf.tmpl"
	outputPath := tempDir + fs.PathSeparator + "out"
	mkdir(outputPath, t)

	defer os.RemoveAll(tempDir)

	data := []byte("hello")
	err := ioutil.WriteFile(confTemplatePath, data, os.ModePerm)
	if err != nil {
		t.Error(err)
	}

	vconfig.Set(PluginName+".conf_template_path", confTemplatePath)
	vconfig.Set(PluginName+".conf_output_path", outputPath)
	vconfig.Set(PluginName+".template_suffix", ".tmpl")

	err = cfgTemplate.DiscoverTemplateFiles(vconfig)
	if err != nil {
		t.Error(err)
	}

	if cfgTemplate.Files.Len() != 1 {
		t.Errorf("unexpected number of template files: %d", cfgTemplate.Files.Len())
	}
	singleFile := cfgTemplate.Files.Front()
	assertEquals(singleFile.Key, confTemplatePath, t)

	expectedVal := PathObject{
		Name:  outputPath + fs.PathSeparator + "nginx.conf",
		IsDir: false,
	}
	assertEquals(singleFile.Value, expectedVal, t)
}

func TestDiscoverTemplateFilesWithDirectory(t *testing.T) {
	cfgTemplate := NewTemplate()
	vconfig := viper.New()

	tempDir := tempDir("nginx-wrapper-unit-tests")
	mkdir(tempDir, t)
	confTemplatePath := tempDir + fs.PathSeparator + "conf"
	mkdir(confTemplatePath, t)
	outputPath := tempDir + fs.PathSeparator + "out"
	mkdir(outputPath, t)

	defer os.RemoveAll(tempDir)

	data := []byte("hello")
	err := ioutil.WriteFile(confTemplatePath+fs.PathSeparator+"nginx.conf.tmpl", data, os.ModePerm)
	if err != nil {
		t.Error(err)
	}
	err = ioutil.WriteFile(confTemplatePath+fs.PathSeparator+"ordinary.text", data, os.ModePerm)
	if err != nil {
		t.Error(err)
	}
	subdir := confTemplatePath + fs.PathSeparator + "subdir"
	mkdir(subdir, t)
	secondLevel := confTemplatePath + fs.PathSeparator + "subdir" + fs.PathSeparator + "2nd-level"
	mkdir(secondLevel, t)
	err = ioutil.WriteFile(secondLevel+fs.PathSeparator+"another.conf.tmpl", data, os.ModePerm)
	if err != nil {
		t.Error(err)
	}

	vconfig.Set(PluginName+".conf_template_path", confTemplatePath)
	vconfig.Set(PluginName+".conf_output_path", outputPath)
	vconfig.Set(PluginName+".template_suffix", ".tmpl")

	err = cfgTemplate.DiscoverTemplateFiles(vconfig)
	if err != nil {
		t.Error(err)
	}

	if cfgTemplate.Files.Len() != 5 {
		t.Errorf("unexpected number of template files: %d", cfgTemplate.Files.Len())
	}

	first := cfgTemplate.Files.Front()
	assertEquals(first.Key, confTemplatePath+fs.PathSeparator+"nginx.conf.tmpl", t)
	expectedFirstVal := PathObject{
		Name:  outputPath + fs.PathSeparator + "nginx.conf",
		IsDir: false,
	}
	assertEquals(first.Value, expectedFirstVal, t)

	second := first.Next()
	assertEquals(second.Key, confTemplatePath+fs.PathSeparator+"ordinary.text", t)
	expectedSecondVal := PathObject{
		Name:  outputPath + fs.PathSeparator + "ordinary.text",
		IsDir: false,
	}
	assertEquals(second.Value, expectedSecondVal, t)

	third := second.Next()
	assertEquals(third.Key, subdir, t)
	expectedThirdVal := PathObject{
		Name:  outputPath + fs.PathSeparator + "subdir" + fs.PathSeparator,
		IsDir: true,
	}
	assertEquals(third.Value, expectedThirdVal, t)

	fourth := third.Next()
	assertEquals(fourth.Key, secondLevel, t)
	expectedFourthVal := PathObject{
		Name:  expectedThirdVal.Name + "2nd-level" + fs.PathSeparator,
		IsDir: true,
	}
	assertEquals(fourth.Value, expectedFourthVal, t)

	fifth := fourth.Next()
	assertEquals(fifth.Key, secondLevel+fs.PathSeparator+"another.conf.tmpl", t)
	expectedFifthVal := PathObject{
		Name:  expectedFourthVal.Name + "another.conf",
		IsDir: false,
	}
	assertEquals(fifth.Value, expectedFifthVal, t)
}

func TestCanTemplateCompositeDefaultsFromConfig(t *testing.T) {
	vconfig := viper.New()
	vconfig.SetDefault(PluginName+".template_suffix", ".tmpl2")
	expected := "./nginx.conf.tmpl2"
	templateText := "{{.template_conf_template_path}}"
	assertCanTemplate(expected, templateText, vconfig, t)
}

func TestCanTemplateTopLevelUndefinedDefaultFromConfig(t *testing.T) {
	expected := config.CoreDefaults["nginx_binary"].(string)
	templateText := "{{.nginx_binary}}"
	vconfig := viper.New()
	assertCanTemplate(expected, templateText, vconfig, t)
}

func TestCanTemplateTopLevelViperDefaultFromConfig(t *testing.T) {
	expected := "something-different"
	templateText := "{{.nginx_binary}}"
	vconfig := viper.New()
	vconfig.SetDefault("nginx_binary", expected)
	assertCanTemplate(expected, templateText, vconfig, t)
}

func TestCanTemplateTopLevelFromConfig(t *testing.T) {
	expected := config.CoreDefaults["nginx_binary"].(string)
	templateText := "{{.nginx_binary}}"
	vconfig := viper.New()
	vconfig.Set("nginx_binary", expected)
	assertCanTemplate(expected, templateText, vconfig, t)
}

func TestCanTemplateTopLevelOverridesDefaultFromConfig(t *testing.T) {
	expected := "another"
	templateText := "{{.nginx_binary}}"
	vconfig := viper.New()
	vconfig.SetDefault("nginx_binary", "something weird")
	vconfig.Set("nginx_binary", expected)
	assertCanTemplate(expected, templateText, vconfig, t)
}

func TestCanTemplateSubLevelUndefinedDefaultFromConfig(t *testing.T) {
	expected := config.LogDefaults["level"].(string)
	templateText := "{{.log_level}}"
	vconfig := viper.New()
	assertCanTemplate(expected, templateText, vconfig, t)
}

func TestCanTemplateSubLevelViperDefaultFromConfig(t *testing.T) {
	expected := "WARN"
	templateText := "{{.log_level}}"
	vconfig := viper.New()
	vconfig.SetDefault("log.level", expected)
	assertCanTemplate(expected, templateText, vconfig, t)
}

func TestCanTemplateSubLevelFromConfig(t *testing.T) {
	expected := config.LogDefaults["level"].(string)
	templateText := "{{.log_level}}"
	vconfig := viper.New()
	vconfig.Set("log.level", expected)
	assertCanTemplate(expected, templateText, vconfig, t)
}

func TestCanTemplateSubLevelOverridesDefaultFromConfig(t *testing.T) {
	expected := "DEBUG"
	templateText := "{{.log_level}}"
	vconfig := viper.New()
	vconfig.SetDefault("log.level", "ERROR")
	vconfig.Set("log.level", expected)
	assertCanTemplate(expected, templateText, vconfig, t)
}

func TestCanTemplateSubLevelMapUndefinedDefaultFromConfig(t *testing.T) {
	expected := fmt.Sprintf("%v", config.TextFormatterOptionsDefaults["full_timestamp"])
	templateText := "{{index .log_formatter_options \"full_timestamp\"}}"
	vconfig := viper.New()
	assertCanTemplate(expected, templateText, vconfig, t)
}

func TestCanTemplateSubLevelMapViperDefaultFromConfig(t *testing.T) {
	expected := "false"
	templateText := "{{index .log_formatter_options \"full_timestamp\"}}"
	vconfig := viper.New()
	options := map[string]interface{}{"full_timestamp": false}
	vconfig.SetDefault("log.formatter_options", options)
	assertCanTemplate(expected, templateText, vconfig, t)
}

func TestCanTemplateSubLevelMapFromConfig(t *testing.T) {
	expected := "false"
	templateText := "{{index .log_formatter_options \"full_timestamp\"}}"
	vconfig := viper.New()
	options := map[string]interface{}{"full_timestamp": false}
	vconfig.Set("log.formatter_options", options)
	assertCanTemplate(expected, templateText, vconfig, t)
}

func TestCanTemplateSubLevelMapOverridesDefaultFromConfig(t *testing.T) {
	expected := "false"
	templateText := "{{index .log_formatter_options \"full_timestamp\"}}"
	vconfig := viper.New()
	options := map[string]interface{}{"full_timestamp": false}
	defaults := map[string]interface{}{"full_timestamp": true}
	vconfig.SetDefault("log.formatter_options", defaults)
	vconfig.Set("log.formatter_options", options)
	assertCanTemplate(expected, templateText, vconfig, t)
}

func TestCanTemplateFile(t *testing.T) {
	unitTestTempDir := fs.TempDirectoryPath("nginx-wrapper-unit-tests")
	mkdir(unitTestTempDir, t)
	defer os.RemoveAll(unitTestTempDir)

	vconfig := viper.New()
	vconfig.SetDefault("run_path", fs.TempDirectoryPath("nginx-wrapper"))
	vconfig.SetDefault("conf_path", vconfig.GetString("run_path")+fs.PathSeparator+"conf")
	vconfig.Set(PluginName+".template_var_left_delim", "{{")
	vconfig.Set(PluginName+".template_var_right_delim", "}}")

	metadata := Metadata(vconfig)
	configDefaults := metadata["config_defaults"].(map[string]interface{})

	for k, v := range configDefaults {
		configKey := PluginName + "." + k
		vconfig.SetDefault(configKey, v)
	}

	config.PluginDefaults[PluginName] = configDefaults

	source := unitTestTempDir + fs.PathSeparator + "test.txt.tmpl"
	output := unitTestTempDir + fs.PathSeparator + "test.txt"

	templateText := `
nginx_binary = {{.nginx_binary}}
template.run_path_subdirs = {{.template_run_path_subdirs}}
template.conf_template_path = {{.template_conf_template_path}}
template.delete_run_path_on_exit = {{.template_delete_run_path_on_exit}}
log.level = {{.log_level}}
log.level.formatter_options.full_timestamp = {{index .log_formatter_options "full_timestamp"}}`

	expected := `
nginx_binary = nginx
template.run_path_subdirs = [client_body conf proxy fastcgi uswsgi scgi]
template.conf_template_path = ./nginx.conf.tmpl
template.delete_run_path_on_exit = false
log.level = INFO
log.level.formatter_options.full_timestamp = true`

	writeErr := ioutil.WriteFile(source, []byte(templateText), 0600)
	if writeErr != nil {
		t.Error(writeErr)
	}

	applyErr := applyFileTemplate(source, output, vconfig)
	if applyErr != nil {
		t.Error(applyErr)
	}

	actualBytes, readErr := ioutil.ReadFile(output)
	if readErr != nil {
		t.Error(readErr)
	}
	actual := string(actualBytes)

	assertEquals(actual, expected, t)
}

func assertCanTemplate(expected string, templateText string, settings api.Settings, t *testing.T) {
	metadata := Metadata(settings)
	configDefaults := metadata["config_defaults"].(map[string]interface{})
	config.PluginDefaults[metadata["name"].(string)] = configDefaults

	for k, v := range config.SubViperConfig(settings, "template").AllSettings() {
		configKey := metadata["name"].(string) + "." + k
		settings.SetDefault(configKey, v)
	}

	cfgTemplate, newTemplateErr := template.New("test").Parse(templateText)
	if newTemplateErr != nil {
		t.Error(newTemplateErr)
	}

	writer := bytes.NewBufferString("")
	applyErr := applyTemplate(cfgTemplate, writer, settings)
	if applyErr != nil {
		t.Error(applyErr)
	}

	actual := writer.String()

	assertEquals(actual, expected, t)
}

func assertEquals(actual interface{}, expected interface{}, t *testing.T) {
	if actual != expected {
		t.Errorf(osenv.LineBreak+"actual:   %v"+
			osenv.LineBreak+osenv.LineBreak+"expected: %v", actual, expected)
	}
}

func mkdir(path string, t *testing.T) {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		t.Error(err)
	}
}
