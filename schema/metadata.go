// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package schema

// Metadata metadata of the object
type Metadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace,omitempty"`
}
