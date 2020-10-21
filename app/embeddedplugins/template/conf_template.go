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
import (
	"github.com/nginxinc/nginx-wrapper/app/config"
	"path/filepath"
)

import (
	"fmt"
	"github.com/elliotchance/orderedmap"
	"github.com/nginxinc/nginx-wrapper/lib/api"
	"github.com/nginxinc/nginx-wrapper/lib/fs"
	"github.com/pkg/errors"
	"io"
	"os"
	"strings"
	"text/template"
)

const uRWgR = fs.OS_USER_RW | fs.OS_GROUP_R
const uRWXgRX = os.ModeDir | (fs.OS_USER_RWX | fs.OS_GROUP_R | fs.OS_GROUP_X)
const templateFilePerm = os.FileMode(uRWgR)
const templateDirPerm = uRWXgRX

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
	Files                 *orderedmap.OrderedMap
	TemplateFileSuffix    string
	ConfTemplatePath      string
	ConfOutputPath        string
	TemplateVarLeftDelim  string
	TemplateVarRightDelim string
}

// NewTemplate creates a new instance of a Template.
func NewTemplate(settings api.Settings) Template {
	return Template{
		Files:                 orderedmap.NewOrderedMap(),
		TemplateFileSuffix:    settings.GetString(PluginName + ".template_suffix"),
		ConfTemplatePath:      filepath.Clean(settings.GetString(PluginName + ".conf_template_path")),
		ConfOutputPath:        filepath.Clean(settings.GetString(PluginName + ".conf_output_path")),
		TemplateVarLeftDelim:  settings.GetString(PluginName + ".template_var_left_delim"),
		TemplateVarRightDelim: settings.GetString(PluginName + ".template_var_right_delim"),
	}
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
		} else if t.isTemplateFile(source) {
			log.Tracef("Templating file from (%s) to (%s)", source, output.Name)
			templateErr := t.applyFileTemplate(source, output.Name, settings)
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
func (t *Template) isTemplateFile(source string) bool {
	return strings.HasSuffix(source, t.TemplateFileSuffix)
}

// applyFileTemplate templates the source file and writes it to the destination.
func (t *Template) applyFileTemplate(source string, output string, settings api.Settings) *ProcessingError {
	// use the delimiters specified in the config
	nginxConfTemplate := template.New("nginx-conf")
	nginxConfTemplate.Delims(t.TemplateVarLeftDelim, t.TemplateVarRightDelim)

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
