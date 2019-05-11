// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package v1alpha1

import "github.com/openfaas/faas-cli/schema"

const APIVersionLatest = "serving.knative.dev/v1alpha1"

//ServingSpec describe characteristics of the object
type ServingSpec struct {
	RunLatest ServingSpecRunLatest `yaml:"runLatest"`
}

type ServingSpecRunLatest struct {
	Configuration ServingSpecRunLatestConfiguration `yaml:"configuration"`
}
type ServingSpecRunLatestConfiguration struct {
	RevisionTemplate ServingSpecRunLatestConfigurationRevisionTemplate `yaml:"revisionTemplate"`
}

type ServingSpecRunLatestConfigurationRevisionTemplate struct {
	Spec ServingSpecRunLatestConfigurationRevisionTemplateSpec `yaml:"spec"`
}

type ServingSpecRunLatestConfigurationRevisionTemplateSpec struct {
	Container ServingSpecRunLatestConfigurationRevisionTemplateSpecContainer `yaml:"container"`
	Volumes   []Volume                                                       `yaml:"volumes,omitempty"`
}

type ServingSpecRunLatestConfigurationRevisionTemplateSpecContainer struct {
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

//ServingCRD root level YAML definition for the object
type ServingCRD struct {
	//APIVersion CRD API version
	APIVersion string `yaml:"apiVersion"`
	//Kind kind of the object
	Kind     string          `yaml:"kind"`
	Metadata schema.Metadata `yaml:"metadata"`
	Spec     ServingSpec     `yaml:"spec"`
}
