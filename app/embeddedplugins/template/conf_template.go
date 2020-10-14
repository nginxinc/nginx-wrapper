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

// The dependency on the config package means that this plugin will always be embedded
// and not an external plugin.
import "github.com/nginxinc/nginx-wrapper/app/config"

import (
	"fmt"
	"github.com/elliotchance/orderedmap"
	"github.com/nginxinc/nginx-wrapper/lib/api"
	"github.com/nginxinc/nginx-wrapper/lib/fs"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const uRWgR = 0640
const uRWXgRX = 0740
const templateFilePerm = os.FileMode(uRWgR)
const templateDirPerm = os.FileMode(uRWXgRX)

// PathObject represents an object on the file system - either
// a directory or a file.
type PathObject struct {
	Name  string
	IsDir bool
}

func (po PathObject) String() string {
	return po.Name
}

// Template represents a set of files/directories that will be
// templated or copied.
type Template struct {
	// map of all template files as keys and output files as values
	Files *orderedmap.OrderedMap
}

// NewTemplate creates a new instance of a Template.
func NewTemplate() Template {
	return Template{
		Files: orderedmap.NewOrderedMap(),
	}
}

// DiscoverTemplateFiles finds all of the template files within the
// conf_template_path.
func (t *Template) DiscoverTemplateFiles(settings api.Settings) error {
	templatePath := settings.GetString(PluginName + ".conf_template_path")
	outputPath := settings.GetString(PluginName + ".conf_output_path")

	templatePathFileInfo, templatePathStatErr := os.Stat(templatePath)
	if templatePathStatErr != nil {
		return errors.Wrapf(templatePathStatErr,
			"error opening nginx conf template path (%s)", templatePath)
	}

	if !templatePathFileInfo.IsDir() {
		// Validate that our single template file's path
		err := validateRegularFileInfo(templatePathFileInfo, templatePath)
		if err != nil {
			return errors.Wrap(err, "template file is not a valid file")
		}
	}

	_, outputPathStatErr := os.Stat(outputPath)
	if outputPathStatErr != nil {
		return errors.Wrapf(outputPathStatErr,
			"error opening nginx conf template output path (%s)", outputPath)
	}

	// Refresh file list if it already has values
	if t.Files.Len() > 0 {
		t.Files = orderedmap.NewOrderedMap()
	}

	// Process simple logic for a single templated config file
	if !templatePathFileInfo.IsDir() {
		outputFile := PathObject{
			Name:  outputPath + fs.PathSeparator + "nginx.conf",
			IsDir: false,
		}

		log.Debug("Adding single template file mapping: ",
			templatePath, " -> ", outputFile)
		t.Files.Set(templatePath, outputFile)
		return nil
	}

	// Process templates, static files and subdirectories contained within
	// the template path directory
	return t.discoverInDirectory(settings, templatePath, outputPath)
}

func (t *Template) discoverInDirectory(settings api.Settings, templatePath string, outputPath string) error {
	// Our template conf path is a directory, so we need to walk it for all files
	walkErr := filepath.Walk(templatePath,
		func(file string, info os.FileInfo, err error) error {
			if err != nil {
				return errors.Wrapf(err, "error processing file (%s)", file)
			}

			if filepath.Base(file) == settings.GetString(PluginName+".template_suffix") {
				return errors.Errorf("can't process filename (%s) that is only a suffix", file)
			}

			var outputFile PathObject

			withoutTemplatePath := strings.TrimPrefix(file, templatePath)
			withOutputPath := outputPath + withoutTemplatePath

			if info.IsDir() {
				// We skip processing of the root directory because the
				// assumption is that it already exists and doesn't need to
				// be discovered.
				if file == templatePath {
					return nil
				}

				outputFile = PathObject{
					Name:  withOutputPath + fs.PathSeparator,
					IsDir: true,
				}
			} else {
				regularFileErr := validateRegularFileInfo(info, file)
				if regularFileErr != nil {
					log.Warnf("unable to process (%s) because it isn't a regular file",
						file)
					return nil
				}

				withoutSuffix := strings.TrimSuffix(withOutputPath,
					settings.GetString(PluginName+".template_suffix"))

				outputFile = PathObject{
					Name:  withoutSuffix,
					IsDir: false,
				}
			}

			log.Debug("Adding template file mapping: ",
				file, " -> ", outputFile)
			t.Files.Set(file, outputFile)
			return nil
		})

	if walkErr != nil {
		return errors.Wrapf(walkErr, "error walking conf_template_path (%s)", templatePath)
	}

	return nil
}

// ApplyTemplating runs all templates and performs substitutions using
// the supplied Viper config instance as a data source.
func (t *Template) ApplyTemplating(settings api.Settings) *ProcessingError {
	if t.Files == nil || t.Files.Len() < 1 {
		return nil
	}

	for el := t.Files.Front(); el != nil; el = el.Next() {
		source := el.Key.(string)
		output := el.Value.(PathObject)

		// Make new directory at destination
		if output.IsDir {
			log.Tracef("Recreating directory from (%s) at (%s)", source, output.Name)
			mkdirErr := os.MkdirAll(output.Name, templateDirPerm)
			if mkdirErr != nil {
				return &ProcessingError{
					Message:              fmt.Sprintf("couldn't make directory: %s", output.Name),
					IsATemplatingProblem: false,
					Err:                  mkdirErr,
				}
			}
			// Template source file and write to destination
		} else if isTemplateFile(source, settings) {
			log.Tracef("Templating file from (%s) to (%s)", source, output.Name)
			templateErr := applyFileTemplate(source, output.Name, settings)
			if templateErr != nil {
				if templateErr.TemplateFile == "" {
					templateErr.TemplateFile = source
				}
				if templateErr.OutputFile == "" {
					templateErr.OutputFile = output.Name
				}

				return templateErr
			}
			// Just copy other static files to destination
		} else {
			log.Tracef("Copying file from (%s) to (%s)", source, output.Name)
			_, copyErr := fs.CopyFile(source, output.Name, 256)
			if copyErr != nil {
				return &ProcessingError{
					Message:              fmt.Sprintf("unable to copy file (%s) to (%s)", source, output.Name),
					IsATemplatingProblem: false,
					Err:                  copyErr,
				}
			}
		}
	}

	return nil
}

// CleanOutputConfiguration removes the configuration that was already
// templatized.
func (t *Template) CleanOutputConfiguration() []error {
	log.Trace("removing nginx configuration")

	if t.Files == nil {
		return []error{}
	}

	var errorPile []error

	for el := t.Files.Front(); el != nil; el = el.Next() {
		pathObject := el.Value.(PathObject)

		filename := pathObject.Name

		if filename == "" {
			log.Warnf("empty filename encountered when cleaning configuration")
			continue
		}

		log.Tracef("removing (%s)", filename)
		err := os.RemoveAll(filename)

		if err != nil {
			wrapped := errors.Wrapf(err, "error removing templated file (%s)",
				pathObject)
			errorPile = append(errorPile, wrapped)
		}
	}

	return errorPile
}

// isTemplateFile determines if a file is a candidate for templating based
// of the file extension.
func isTemplateFile(source string, settings api.Settings) bool {
	suffix := settings.GetString(PluginName + ".template_suffix")
	return strings.HasSuffix(source, suffix)
}

// applyFileTemplate templates the source file and writes it to the destination.
func applyFileTemplate(source string, output string, settings api.Settings) *ProcessingError {
	// use the delimiters specified in the config
	nginxConfTemplate := template.New("nginx-conf")
	nginxConfTemplate.Delims(
		settings.GetString(PluginName+".template_var_left_delim"),
		settings.GetString(PluginName+".template_var_right_delim"))

	parsedTemplate, parseErr := nginxConfTemplate.ParseFiles(source)

	if parseErr != nil {
		return &ProcessingError{
			Message:              "unable to parse",
			TemplateFile:         source,
			IsATemplatingProblem: true,
			Err:                  parseErr,
		}
	}

	writer, openErr := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, templateFilePerm)
	if openErr != nil {
		return &ProcessingError{
			Message:              "unable to open destination for write",
			TemplateFile:         source,
			IsATemplatingProblem: false,
			Err:                  openErr,
		}
	}
	defer func() {
		err := writer.Close()
		if err != nil {
			log.Warnf("problem closing template file (%s) after write: %v",
				output, err)
		}
	}()

	return applyTemplate(parsedTemplate, writer, settings)
}

// applyTemplate executes the go templates object and creates a list of the parameters
// to be passed to it.
func applyTemplate(parsedTemplate *template.Template, writer io.Writer, settings api.Settings) *ProcessingError {
	// go templates do not support the dot character (.) easily, so
	// we template variables using the underscore (_) as a delimiter
	params := config.AllKnownElements(settings, "_")

	// Preferably, we could use parsedTemplate.Execute() and not have to do
	// this two step operation where we get the template name and then explicitly
	// execute it. However, in applyFileTemplate() we make use of the
	// nginxConfTemplate.Delims() method which must be called after a new template
	// instance is created, so we are stuck using this API.
	templateName := parsedTemplate.Templates()[0].Name()
	templateErr := parsedTemplate.ExecuteTemplate(writer, templateName, params)
	if templateErr != nil {
		return &ProcessingError{
			Message:              "unable to apply template",
			TemplateName:         templateName,
			IsATemplatingProblem: true,
			Err:                  templateErr,
		}
	}

	return nil
}

// validateRegularFileInfo throws an error if the pathInfo passed refers to anything
// that isn't a readable regular file.
func validateRegularFileInfo(fileInfo os.FileInfo, path string) error {
	if fileInfo.IsDir() {
		return errors.Errorf("path (%s) is a directory and not a file",
			path)
	}

	if !fileInfo.Mode().IsRegular() {
		return errors.Errorf("path (%s) is not a regular file", path)
	}

	return nil
}
