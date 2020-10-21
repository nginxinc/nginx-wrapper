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
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

// DiscoverTemplateFiles finds all of the template files within the
// conf_template_path.
func (t *Template) DiscoverTemplateFiles() error {
	templatePath := t.ConfTemplatePath
	outputPath := t.ConfOutputPath

	templatePathFileInfo, templatePathStatErr := os.Stat(templatePath)
	if templatePathStatErr != nil {
		return errors.Wrapf(templatePathStatErr,
			"error opening nginx conf template path (%s)", templatePath)
	}

	if !fs.IsRegularFileOrDirectory(templatePathFileInfo) {
		return errors.Errorf("template path (%s) is not a valid file or directory", templatePath)
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
	return t.discoverInDirectory(templatePath)
}

func (t *Template) discoverInDirectory(templatePath string) error {
	// Our template conf path is a directory, so we need to walk it for all files
	walkErr := filepath.Walk(templatePath, t.processTemplatePath)

	if walkErr != nil {
		return errors.Wrapf(walkErr, "error walking conf_template_path (%s)", templatePath)
	}

	return nil
}

func (t *Template) processTemplatePath(inputPath string, info os.FileInfo, err error) error {
	if err != nil {
		return errors.Wrapf(err, "error processing path (%s)", inputPath)
	}

	// normalize the path because who knows what form it may be in
	path := filepath.Clean(inputPath)

	if filepath.Base(path) == t.TemplateFileSuffix {
		return errors.Errorf("can't process filename (%s) that is only a suffix", path)
	}

	var outputFile PathObject

	withoutTemplatePath := t.removeTemplatePath(path)
	withOutputPath := filepath.Clean(t.ConfOutputPath + fs.PathSeparator + withoutTemplatePath)

	if !fs.IsRegularFileOrDirectory(info) {
		log.Warnf("unable to process path (%s) because it isn't a regular file or directory", path)
		return nil
	}

	if info.IsDir() {
		// We skip processing of the root directory because the
		// assumption is that it already exists and doesn't need to
		// be discovered.
		if path == t.ConfTemplatePath {
			return nil
		}

		outputFile = PathObject{
			Name:  withOutputPath + fs.PathSeparator,
			IsDir: true,
		}
	} else {
		withoutSuffix := strings.TrimSuffix(withOutputPath, t.TemplateFileSuffix)

		outputFile = PathObject{
			Name:  withoutSuffix,
			IsDir: false,
		}
	}

	log.Debug("Adding template file mapping: ",
		path, " -> ", outputFile)
	t.Files.Set(path, outputFile)
	return nil
}

func (t *Template) removeTemplatePath(file string) string {
	return strings.TrimPrefix(file, t.ConfTemplatePath)
}
