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

	FProcess string `yaml:"fprocess"`

	Environment map[string]string `yaml:"environment"`

	SkipBuild bool `yaml:"skip_build"`

	Constraints *[]string `yaml:"constraints"`

	EnvironmentFile []string `yaml:"environment_file"`
}

type EnvironmentFile struct {
	Environment map[string]string `yaml:"environment"`
}

// Services root level YAML file to define FaaS function-set
type Services struct {
	Functions map[string]Function `yaml:"functions,omitempty"`
	Provider  Provider            `yaml:"provider,omitempty"`
}
