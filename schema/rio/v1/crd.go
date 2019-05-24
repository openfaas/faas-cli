// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package v1

import "github.com/openfaas/faas-cli/schema"

const APIVersionLatest = "rio.cattle.io/v1"

//RioSpec describe characteristics of the object
type Spec struct {
	Env   []EnvPair `yaml:"env,omitempty"`
	Image string    `yaml:"image"`
	Ports []Port    `yaml:"ports"`
}

type Port struct {
	Port       int    `yaml:"port"`
	Protocol   string `yaml:"protocol"`
	TargetPort int    `yaml:"targetPort"`
}

type EnvPair struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
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
