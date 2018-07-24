// Copyright (c) OpenFaaS Author(s) 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package schema

import "github.com/openfaas/faas-cli/stack"

//Metadata metadata of the object
type Metadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

//Spec describe characteristics of the object
type Spec struct {
	//Name name of the function
	Name string `yaml:"name"`
	//Image docker image name of the function
	Image string `yaml:"image"`

	Environment map[string]string `yaml:"environment,omitempty"`

	Labels *map[string]string `yaml:"labels,omitempty"`

	//Limits for the function
	Limits *stack.FunctionResources `yaml:"limits,omitempty"`

	//Requests of resources requested by function
	Requests *stack.FunctionResources `yaml:"requests,omitempty"`

	Constraints *[]string `yaml:"constraints,omitempty"`

	//Secrets list of secrets to be made available to function
	Secrets []string `yaml:"secrets,omitempty"`
}

//CRD root level YAML definition for the object
type CRD struct {
	//APIVersion CRD API version
	APIVersion string `yaml:"apiVersion"`
	//Kind kind of the object
	Kind     string   `yaml:"kind"`
	Metadata Metadata `yaml:"metadata"`
	Spec     Spec     `yaml:"spec"`
}
