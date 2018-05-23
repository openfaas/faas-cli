// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package stack

// Provider for the FaaS set of functions.
type Provider struct {
	Name       string `yaml:"name"`
	GatewayURL string `yaml:"gateway"`
	Network    string `yaml:"network"`
}

// Function as deployed or built on FaaS
type Function struct {
	// Name of deployed function
	Name     string `yaml:"-"`
	Language string `yaml:"lang"`

	// Handler Local folder to use for function
	Handler string `yaml:"handler"`

	// Image Docker image name
	Image string `yaml:"image"`

	// Docker registry Authorization
	RegistryAuth string `yaml:"registry_auth,omitempty"`

	FProcess string `yaml:"fprocess"`

	Environment map[string]string `yaml:"environment"`

	// Secrets list of secrets to be made available to function
	Secrets []string `yaml:"secrets"`

	SkipBuild bool `yaml:"skip_build"`

	Constraints *[]string `yaml:"constraints"`

	// EnvironmentFile is a list of files to import and override environmental variables.
	// These are overriden in order.
	EnvironmentFile []string `yaml:"environment_file"`

	Labels *map[string]string `yaml:"labels"`

	// Limits for function
	Limits *FunctionResources `yaml:"limits"`

	// Requests of resources requested by function
	Requests *FunctionResources `yaml:"requests"`
}

// FunctionResources Memory and CPU
type FunctionResources struct {
	Memory string `yaml:"memory"`
	CPU    string `yaml:"cpu"`
}

// EnvironmentFile represents external file for environment data
type EnvironmentFile struct {
	Environment map[string]string `yaml:"environment"`
}

// Services root level YAML file to define FaaS function-set
type Services struct {
	Functions map[string]Function `yaml:"functions,omitempty"`
	Provider  Provider            `yaml:"provider,omitempty"`
}

// LanguageTemplate read from template.yml within root of a language template folder
type LanguageTemplate struct {
	Language     string        `yaml:"language"`
	FProcess     string        `yaml:"fprocess"`
	BuildOptions []BuildOption `yaml:"build_options"`
}

type BuildOption struct {
	Name     string   `yaml:"name"`
	Packages []string `yaml:"packages"`
}
