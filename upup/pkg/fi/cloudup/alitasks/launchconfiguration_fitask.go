// +build !ignore_autogenerated

/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by fitask. DO NOT EDIT.

package alitasks

import (
	"k8s.io/kops/upup/pkg/fi"
)

// LaunchConfiguration

var _ fi.HasLifecycle = &LaunchConfiguration{}

// GetLifecycle returns the Lifecycle of the object, implementing fi.HasLifecycle
func (o *LaunchConfiguration) GetLifecycle() fi.Lifecycle {
	return o.Lifecycle
}

// SetLifecycle sets the Lifecycle of the object, implementing fi.SetLifecycle
func (o *LaunchConfiguration) SetLifecycle(lifecycle fi.Lifecycle) {
	o.Lifecycle = lifecycle
}

var _ fi.HasName = &LaunchConfiguration{}

// GetName returns the Name of the object, implementing fi.HasName
func (o *LaunchConfiguration) GetName() *string {
	return o.Name
}

// String is the stringer function for the task, producing readable output using fi.TaskAsString
func (o *LaunchConfiguration) String() string {
	return fi.TaskAsString(o)
}
