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
	"github.com/elliotchance/orderedmap"
	"github.com/nginxinc/nginx-wrapper/lib/fs"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestDiscoverTemplateFilesWithSingleFile(t *testing.T) {
	tempDir := makeTestDirectory(t)
	defer os.RemoveAll(tempDir)

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

	vconfig := viper.New()
	vconfig.Set(PluginName+".conf_template_path", confTemplatePath)
	vconfig.Set(PluginName+".conf_output_path", outputPath)
	vconfig.Set(PluginName+".template_suffix", ".tmpl")
	cfgTemplate := NewTemplate(vconfig)

	err = cfgTemplate.DiscoverTemplateFiles()
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
	tempDir := makeTestDirectory(t)
	defer os.RemoveAll(tempDir)

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

	vconfig := viper.New()
	vconfig.Set(PluginName+".conf_template_path", confTemplatePath)
	vconfig.Set(PluginName+".conf_output_path", outputPath)
	vconfig.Set(PluginName+".template_suffix", ".tmpl")

	cfgTemplate := NewTemplate(vconfig)

	err = cfgTemplate.DiscoverTemplateFiles()
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

func TestProcessTemplatePathFile(t *testing.T) {
	template := Template{
		Files:              orderedmap.NewOrderedMap(),
		TemplateFileSuffix: ".tmpl",
		ConfTemplatePath:   "template",
		ConfOutputPath:     "/var/run/nginx-wrapper/conf",
	}

	file := "template/conf.d/default.conf.tmpl"
	info := FakeFileInfo{
		isDir: false,
		mode:  os.ModePerm,
	}
	err := template.processTemplatePath(file, info, nil)
	if err != nil {
		t.Error(err)
	}

	if template.Files.Len() != 1 {
		t.Errorf("unexpected number of files (%d) added to template files - expected 1",
			template.Files.Len())
	}
}

func TestProcessTemplatePathDirectory(t *testing.T) {
	template := Template{
		Files:              orderedmap.NewOrderedMap(),
		TemplateFileSuffix: ".tmpl",
		ConfTemplatePath:   "template",
		ConfOutputPath:     "/var/run/nginx-wrapper/conf",
	}

	file := "template/conf.d"
	info := FakeFileInfo{
		isDir: true,
		mode:  os.ModeDir,
	}
	err := template.processTemplatePath(file, info, nil)
	if err != nil {
		t.Error(err)
	}

	if template.Files.Len() != 1 {
		t.Errorf("unexpected number of files (%d) added to template files - expected 1",
			template.Files.Len())
	}
}

func TestProcessTemplatePathSymbolicLink(t *testing.T) {
	template := Template{
		Files:              orderedmap.NewOrderedMap(),
		TemplateFileSuffix: ".tmpl",
		ConfTemplatePath:   "template",
		ConfOutputPath:     "/var/run/nginx-wrapper/conf",
	}

	file := "template/conf.d/etc/nginx.conf"
	info := FakeFileInfo{
		isDir: false,
		mode:  os.ModeSymlink,
	}
	err := template.processTemplatePath(file, info, nil)
	if err != nil {
		t.Error(err)
	}

	if template.Files.Len() != 0 {
		t.Errorf("unexpected number of files (%d) added to template files - expected 0",
			template.Files.Len())
	}
}

type FakeFileInfo struct {
	size    int64
	name    string
	modTime time.Time
	mode    os.FileMode
	isDir   bool
}

func (f FakeFileInfo) Mode() os.FileMode {
	return f.mode
}

func (f FakeFileInfo) ModTime() time.Time {
	return f.modTime
}

func (f FakeFileInfo) IsDir() bool {
	return f.isDir
}

func (f FakeFileInfo) Sys() interface{} {
	return nil
}

func (f FakeFileInfo) Name() string {
	return f.name
}

func (f FakeFileInfo) Size() int64 {
	return f.size
}

func makeTestDirectory(t *testing.T) string {
	tempDir := fs.TempDirectoryPath("nginx-wrapper-test" + fs.PathSeparator + t.Name())
	mkdirErr := os.MkdirAll(tempDir, os.ModeDir|(fs.OS_USER_RWX|fs.OS_ALL_R))
	if mkdirErr != nil {
		t.Error(mkdirErr)
	}

	return tempDir
}
