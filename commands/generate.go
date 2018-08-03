// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"strings"

	"github.com/openfaas/faas-cli/builder"
	"github.com/openfaas/faas-cli/schema"
	"github.com/openfaas/faas-cli/stack"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

const (
	defaultFunctionNamespace = "openfaas-fn"
	resourceKind             = "Function"
	defaultAPIVersion        = "openfaas.com/v1alpha2"
)

var (
	api               string
	functionNamespace string
)

func init() {
	generateCmd.Flags().StringVar(&api, "api", defaultAPIVersion, "OpenFaaS CRD API version")
	generateCmd.Flags().StringVarP(&functionNamespace, "namespace", "n", defaultFunctionNamespace, "Kubernetes namespace for functions")
	generateCmd.Flags().StringVar(&tag, "tag", "file", "Tag Docker image for function, specify file or SHA")
	faasCmd.AddCommand(generateCmd)
}

var generateCmd = &cobra.Command{
	Use:   "generate --api=openfaas.com/v1alpha2 --yaml stack.yml --tag=sha --namespace=openfaas-fn",
	Short: "Generate Kubernetes CRD YAML file",
	Long:  `The generate command creates kubernetes CRD YAML file for functions`,
	Example: `faas-cli generate --api=openfaas.com/v1alpha2 --yaml stack.yml | kubectl apply  -f -
faas-cli generate --api=openfaas.com/v1alpha2 -f stack.yml
faas-cli generate --api=openfaas.com/v1alpha2 --namespace openfaas-fn -f stack.yml
faas-cli generate --api=openfaas.com/v1alpha2 -f stack.yml --tag=sha -n openfaas-fn`,
	PreRunE: preRunGenerate,
	RunE:    runGenerate,
}

func preRunGenerate(cmd *cobra.Command, args []string) error {
	if len(api) == 0 {
		return fmt.Errorf("You must supply api version with the --api flag")
	}
	return nil
}

func runGenerate(cmd *cobra.Command, args []string) error {

	var services stack.Services

	//Process function stack file
	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAMLFile(yamlFile, regex, filter)
		if err != nil {
			return err
		}

		if parsedServices != nil {
			services = *parsedServices
		}
	}

	format := schema.DefaultFormat
	var version string
	var branch string

	if strings.ToLower(tag) == "sha" {
		version = builder.GetGitSHA()
		if len(version) == 0 {
			return fmt.Errorf("cannot tag image with Git SHA as this is not a Git repository")

		}
		format = schema.SHAFormat
	}

	if strings.ToLower(tag) == "branch" {
		branch = builder.GetGitBranch()
		if len(branch) == 0 {
			return fmt.Errorf("cannot tag image with Git branch and SHA as this is not a Git repository")

		}
		version = builder.GetGitSHA()
		if len(version) == 0 {
			return fmt.Errorf("cannot tag image with Git SHA as this is not a Git repository")

		}
		format = schema.BranchAndSHAFormat
	}

	objectsString, err := generateCRDYAML(services, format, api, functionNamespace, branch, version)
	if err != nil {
		return err
	}

	if len(objectsString) > 0 {
		fmt.Println(objectsString)
	}
	return nil
}

//generateCRDYAML generates CRD YAML for functions
func generateCRDYAML(services stack.Services, format schema.BuildFormat, apiVersion, namespace, branch, version string) (string, error) {

	var objectsString string

	if len(services.Functions) > 0 {
		for name, function := range services.Functions {
			//read environment variables from the file
			fileEnvironment, err := readFiles(function.EnvironmentFile)
			if err != nil {
				return "", err
			}

			//combine all environment variables
			allEnvironment, envErr := compileEnvironment([]string{}, function.Environment, fileEnvironment)
			if envErr != nil {
				return "", envErr
			}

			metadata := schema.Metadata{Name: name, Namespace: namespace}
			imageName := schema.BuildImageName(format, function.Image, version, branch)

			spec := schema.Spec{
				Name:        name,
				Image:       imageName,
				Environment: allEnvironment,
				Labels:      function.Labels,
				Limits:      function.Limits,
				Requests:    function.Requests,
				Constraints: function.Constraints,
				Secrets:     function.Secrets,
			}

			crd := schema.CRD{
				APIVersion: apiVersion,
				Kind:       resourceKind,
				Metadata:   metadata,
				Spec:       spec,
			}

			//Marshal the object definition to yaml
			objectString, err := yaml.Marshal(crd)
			if err != nil {
				return "", err
			}
			objectsString += "---\n" + string(objectString)
		}
	}

	return objectsString, nil
}
