// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"

	"github.com/openfaas/faas-cli/builder"
	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/schema"
	knativev1alpha1 "github.com/openfaas/faas-cli/schema/knative/v1alpha1"
	openfaasv1alpha2 "github.com/openfaas/faas-cli/schema/openfaas/v1alpha2"
	"github.com/openfaas/faas-cli/stack"
	"github.com/pkg/errors"
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
	fromStore         string
)

func init() {

	generateCmd.Flags().StringVar(&fromStore, "from-store", "", "generate using a store image")

	generateCmd.Flags().StringVar(&api, "api", defaultAPIVersion, "CRD API version e.g openfaas.com/v1alpha2, serving.knative.dev/v1alpha1")
	generateCmd.Flags().StringVarP(&functionNamespace, "namespace", "n", defaultFunctionNamespace, "Kubernetes namespace for functions")
	generateCmd.Flags().Var(&tagFormat, "tag", "Override latest tag on function Docker image, accepts 'latest', 'sha', 'branch', 'describe'")
	generateCmd.Flags().BoolVar(&envsubst, "envsubst", true, "Substitute environment variables in stack.yml file")

	faasCmd.AddCommand(generateCmd)
}

var generateCmd = &cobra.Command{
	Use:   "generate --api=openfaas.com/v1alpha2 --yaml stack.yml --tag sha --namespace=openfaas-fn",
	Short: "Generate Kubernetes CRD YAML file",
	Long:  `The generate command creates kubernetes CRD YAML file for functions`,
	Example: `faas-cli generate --api=openfaas.com/v1alpha2 --yaml stack.yml | kubectl apply  -f -
faas-cli generate --api=openfaas.com/v1alpha2 -f stack.yml
faas-cli generate --api=serving.knative.dev/v1alpha1 -f stack.yml
faas-cli generate --api=openfaas.com/v1alpha2 --namespace openfaas-fn -f stack.yml
faas-cli generate --api=openfaas.com/v1alpha2 -f stack.yml --tag branch -n openfaas-fn`,
	PreRunE: preRunGenerate,
	RunE:    runGenerate,
}

func preRunGenerate(cmd *cobra.Command, args []string) error {
	if len(api) == 0 {
		return fmt.Errorf("You must supply api version with the --api flag")
	}
	return nil
}

func filterStoreItem(items []schema.StoreItem, fromStore string) (*schema.StoreItem, error) {
	var item *schema.StoreItem

	for _, val := range items {
		if val.Name == fromStore {
			item = &val
			break
		}
	}

	if item == nil {
		return nil, fmt.Errorf("unable to find '%s' in store", fromStore)
	}

	return item, nil
}

func runGenerate(cmd *cobra.Command, args []string) error {

	var services stack.Services

	if len(fromStore) > 0 {
		services = stack.Services{
			Provider: stack.Provider{
				Name: "openfaas",
			},
			Version: "1.0",
		}

		services.Functions = make(map[string]stack.Function)

		items, err := proxy.FunctionStoreList(storeAddress)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Unable to retrieve functions from URL %s", storeAddress))
		}

		item, err := filterStoreItem(items, fromStore)
		if err != nil {
			return err
		}

		services.Functions[item.Name] = stack.Function{
			Name:        item.Name,
			Image:       item.Image,
			Labels:      &item.Labels,
			Annotations: &item.Annotations,
			Environment: item.Environment,
			FProcess:    item.Fprocess,
		}

	} else if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAMLFile(yamlFile, regex, filter, envsubst)
		if err != nil {
			return err
		}

		if parsedServices != nil {
			services = *parsedServices
		}
	}

	branch, version, err := builder.GetImageTagValues(tagFormat)
	if err != nil {
		return err
	}

	objectsString, err := generateCRDYAML(services, tagFormat, api, functionNamespace, branch, version)
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

		if apiVersion == knativev1alpha1.APIVersionLatest {
			return generateknativev1alpha1ServingCRDYAML(services, format, api, functionNamespace, branch, version)
		}

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

			spec := openfaasv1alpha2.Spec{
				Name:        name,
				Image:       imageName,
				Environment: allEnvironment,
				Labels:      function.Labels,
				Limits:      function.Limits,
				Requests:    function.Requests,
				Constraints: function.Constraints,
				Secrets:     function.Secrets,
			}

			crd := openfaasv1alpha2.CRD{
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

func generateknativev1alpha1ServingCRDYAML(services stack.Services, format schema.BuildFormat, apiVersion, namespace, branch, version string) (string, error) {
	crds := []knativev1alpha1.ServingCRD{}

	for name, function := range services.Functions {

		fileEnvironment, err := readFiles(function.EnvironmentFile)
		if err != nil {
			return "", err
		}

		//combine all environment variables
		allEnvironment, envErr := compileEnvironment([]string{}, function.Environment, fileEnvironment)
		if envErr != nil {
			return "", envErr
		}

		var env []knativev1alpha1.EnvPair
		for k, v := range allEnvironment {
			env = append(env, knativev1alpha1.EnvPair{Name: k, Value: v})
		}

		crd := knativev1alpha1.ServingCRD{
			Metadata: schema.Metadata{
				Name:      name,
				Namespace: namespace,
			},
			APIVersion: apiVersion,
			Kind:       "Service",
			Spec: knativev1alpha1.ServingSpec{
				RunLatest: knativev1alpha1.ServingSpecRunLatest{

					Configuration: knativev1alpha1.ServingSpecRunLatestConfiguration{
						RevisionTemplate: knativev1alpha1.ServingSpecRunLatestConfigurationRevisionTemplate{
							Spec: knativev1alpha1.ServingSpecRunLatestConfigurationRevisionTemplateSpec{
								Container: knativev1alpha1.ServingSpecRunLatestConfigurationRevisionTemplateSpecContainer{
									Image: function.Image,
									Env:   env,
								},
							},
						},
					},
				},
			},
		}

		var mounts []knativev1alpha1.VolumeMount
		var volumes []knativev1alpha1.Volume

		for _, secret := range function.Secrets {
			mounts = append(mounts, knativev1alpha1.VolumeMount{
				MountPath: "/var/openfaas/secrets/" + secret,
				ReadOnly:  true,
				Name:      secret,
			})
			volumes = append(volumes, knativev1alpha1.Volume{
				Name: secret,
				Secret: knativev1alpha1.Secret{
					SecretName: secret,
				},
			})
		}

		crd.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.VolumeMounts = mounts
		crd.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Volumes = volumes

		crds = append(crds, crd)
	}

	var objectsString string
	for _, crd := range crds {
		//Marshal the object definition to yaml
		objectString, err := yaml.Marshal(crd)
		if err != nil {
			return "", err
		}
		objectsString += "---\n" + string(objectString)
	}

	return objectsString, nil
}
