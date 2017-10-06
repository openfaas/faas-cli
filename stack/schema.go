// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package stack

// Provider for the FaaS set of functions.
type Provider struct {
	Name       string `yaml:"name"`
	GatewayURL string `yaml:"gateway"`
	Network    string `yaml:"network,omitempty"`
}

// Function as deployed or built on FaaS
type Function struct {
	// Name of deployed function
	Name     string `yaml:"-,omitempty"`
	Language string `yaml:"lang"`

	// Handler Local folder to use for function
	Handler string `yaml:"handler"`

	// Image Docker image name
	Image string `yaml:"image"`

	FProcess string `yaml:"fprocess,omitempty"`

	Environment map[string]string `yaml:"environment,omitempty"`

	SkipBuild bool `yaml:"skip_build,omitempty"`

	Constraints *[]string `yaml:"constraints,omitempty"`

	// EnvironmentFile is a list of files to import and override environmental variables.
	// These are overriden in order.
	EnvironmentFile []string `yaml:"environment_file,omitempty"`

	Labels *map[string]string `yaml:"labels,omitempty"`
}

// EnvironmentFile represents external file for environment data
type EnvironmentFile struct {
	Environment map[string]string `yaml:"environment"`
}

// Services root level YAML file to define FaaS function-set
type Services struct {
	Provider  Provider            `yaml:"provider,omitempty"`
	Functions map[string]Function `yaml:"functions,omitempty"`
}
