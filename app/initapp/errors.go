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

package initapp

import (
	"fmt"
	"reflect"
)

// PluginMetadataError indicates that the plugin metadata is missing a required key.
type PluginMetadataError struct {
	Err error
}

// Cause returns the cause of the error - the same as the Unwrap() method.
func (p *PluginMetadataError) Cause() error {
	return p.Err
}

func (p *PluginMetadataError) Unwrap() error {
	return p.Err
}

// PluginMetadataKeyMissing indicates that the plugin metadata is missing a required key.
type PluginMetadataKeyMissing struct {
	KeyMissing string
	PluginMetadataError
}

func (p *PluginMetadataKeyMissing) Error() string {
	return fmt.Sprintf("plugin metadata missing required key (%s)", p.KeyMissing)
}

// PluginMetadataMalformedType indicates that the plugin metadata was passed the wrong data type.
type PluginMetadataMalformedType struct {
	KeyName      string
	ExpectedType reflect.Kind
	PluginMetadataError
}

func (p *PluginMetadataMalformedType) Error() string {
	return fmt.Sprintf("plugin metadata must use (%v) for (%s) value)",
		p.ExpectedType, p.KeyName)
}
