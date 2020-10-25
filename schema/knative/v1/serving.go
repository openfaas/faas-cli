// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package v1

import "github.com/openfaas/faas-cli/schema"

const APIVersionLatest = "serving.knative.dev/v1"

//ServingServiceCRD root level YAML definition for the object
type ServingServiceCRD struct {
	//APIVersion CRD API version
	APIVersion string `yaml:"apiVersion"`
	//Kind kind of the object
	Kind     string             `yaml:"kind"`
	Metadata schema.Metadata    `yaml:"metadata,omitempty"`
	Spec     ServingServiceSpec `yaml:"spec"`
}

type ServingServiceSpec struct {
	ServingServiceSpecTemplate `yaml:"template"`
}

type ServingServiceSpecTemplateSpec struct {
	Containers []ServingSpecContainersContainerSpec `yaml:"containers"`
	Volumes    []Volume                             `yaml:"volumes,omitempty"`
}
type ServingServiceSpecTemplate struct {
	Template ServingServiceSpecTemplateSpec `yaml:"spec"`
}

type ServingSpecContainersContainerSpec struct {
	Image        string        `yaml:"image"`
	Env          []EnvPair     `yaml:"env,omitempty"`
	VolumeMounts []VolumeMount `yaml:"volumeMounts,omitempty"`
}

type VolumeMount struct {
	Name      string `yaml:"name"`
	MountPath string `yaml:"mountPath"`
	ReadOnly  bool   `yaml:"readOnly"`
}

type Volume struct {
	Name   string `yaml:"name"`
	Secret Secret `yaml:"secret"`
}

type Secret struct {
	SecretName string `yaml:"secretName"`
}

type EnvPair struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}
