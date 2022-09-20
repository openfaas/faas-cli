// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package v1

import (
	"github.com/openfaas/faas-cli/schema"
	"github.com/openfaas/faas-cli/stack"
)

//APIVersionLatest latest API version of CRD
const APIVersionLatest = "openfaas.com/v1"

//Spec describe characteristics of the object
type Spec struct {
	//Name name of the function
	Name string `yaml:"name"`
	//Image docker image name of the function
	Image string `yaml:"image"`

	Environment map[string]string `yaml:"environment,omitempty"`

	Labels *map[string]string `yaml:"labels,omitempty"`

	Annotations *map[string]string `yaml:"annotations,omitempty"`

	//Limits for the function
	Limits *stack.FunctionResources `yaml:"limits,omitempty"`

	//Requests of resources requested by function
	Requests *stack.FunctionResources `yaml:"requests,omitempty"`

	Constraints *[]string `yaml:"constraints,omitempty"`

	//Secrets list of secrets to be made available to function
	Secrets []string `yaml:"secrets,omitempty"`

	ReadOnlyRootFilesystem bool `yaml:"readOnlyRootFilesystem,omitempty"`
}

//CRD root level YAML definition for the object
type CRD struct {
	//APIVersion CRD API version
	APIVersion string `yaml:"apiVersion"`
	//Kind kind of the object
	Kind     string          `yaml:"kind"`
	Metadata schema.Metadata `yaml:"metadata"`
	Spec     Spec            `yaml:"spec"`
}
