// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"bytes"
	"fmt"
	"os"
	"sort"

	"github.com/openfaas/faas-cli/builder"
	v2 "github.com/openfaas/faas-cli/schema/store/v2"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/schema"
	knativev1 "github.com/openfaas/faas-cli/schema/knative/v1"
	openfaasv1 "github.com/openfaas/faas-cli/schema/openfaas/v1"
	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/faas-cli/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	yaml "gopkg.in/yaml.v3"
)

const (
	resourceKind      = "Function"
	defaultAPIVersion = "openfaas.com/v1"
)

var (
	api                  string
	name                 string
	functionNamespace    string
	crdFunctionNamespace string
	fromStore            string
	desiredArch          string
	annotationArgs       []string
)

func init() {

	generateCmd.Flags().StringVar(&fromStore, "from-store", "", "generate using a store image")
	generateCmd.Flags().StringVar(&name, "name", "", "for use with --from-store, override the name for the Function CR")

	generateCmd.Flags().StringVar(&api, "api", defaultAPIVersion, "CRD API version e.g openfaas.com/v1, serving.knative.dev/v1")
	generateCmd.Flags().StringVarP(&crdFunctionNamespace, "namespace", "n", "openfaas-fn", "Kubernetes namespace for functions")
	generateCmd.Flags().Var(&tagFormat, "tag", "Override latest tag on function Docker image, accepts 'digest', 'latest', 'sha', 'branch', 'describe'")
	generateCmd.Flags().BoolVar(&envsubst, "envsubst", true, "Substitute environment variables in stack.yml file")
	generateCmd.Flags().StringVar(&desiredArch, "arch", "x86_64", "Desired image arch. (Default x86_64)")
	generateCmd.Flags().StringArrayVar(&annotationArgs, "annotation", []string{}, "Any annotations you want to add (to store functions only)")

	faasCmd.AddCommand(generateCmd)
}

var generateCmd = &cobra.Command{
	Use:   "generate --api=openfaas.com/v1 --yaml stack.yml --tag sha --namespace=openfaas-fn",
	Short: "Generate Kubernetes CRD YAML file",
	Long:  `The generate command creates kubernetes CRD YAML file for functions`,
	Example: `faas-cli generate --api=openfaas.com/v1 --yaml stack.yml | kubectl apply  -f -
faas-cli generate --api=openfaas.com/v1 -f stack.yml
faas-cli generate --api=serving.knative.dev/v1 -f stack.yml
faas-cli generate --api=openfaas.com/v1 --namespace openfaas-fn -f stack.yml
faas-cli generate --api=openfaas.com/v1 -f stack.yml --tag branch -n openfaas-fn`,
	PreRunE: preRunGenerate,
	RunE:    runGenerate,
}

func preRunGenerate(cmd *cobra.Command, args []string) error {
	if len(api) == 0 {
		return fmt.Errorf("you must supply the API version with the --api flag")
	}

	return nil
}

func filterStoreItem(items []v2.StoreFunction, fromStore string) (*v2.StoreFunction, error) {
	var item *v2.StoreFunction

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

	desiredArch, _ := cmd.Flags().GetString("arch")
	var services stack.Services

	var annotations map[string]string

	annotations, annotationErr := util.ParseMap(annotationArgs, "annotation")
	if annotationErr != nil {
		return fmt.Errorf("error parsing annotations: %v", annotationErr)
	}

	if fromStore != "" {
		if tagFormat == schema.DigestFormat {
			return fmt.Errorf("digest tag format is not supported for store functions")
		}

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

		_, ok := item.Images[desiredArch]
		if !ok {
			var keys []string
			for k := range item.Images {
				keys = append(keys, k)
			}
			return errors.New(fmt.Sprintf("image for %s not found in store. \noptions: %s", desiredArch, keys))
		}

		allAnnotations := util.MergeMap(item.Annotations, annotations)

		if len(item.Fprocess) > 0 {
			if item.Environment == nil {
				item.Environment = make(map[string]string)
			}

			if _, ok := item.Environment["fprocess"]; !ok {
				item.Environment["fprocess"] = item.Fprocess
			}
		}

		fullName := item.Name
		if len(name) > 0 {
			fullName = name
		}

		services.Functions[fullName] = stack.Function{
			Name:        fullName,
			Image:       item.Images[desiredArch],
			Labels:      &item.Labels,
			Annotations: &allAnnotations,
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
	} else {
		fmt.Println(
			`"stack.yml" file not found in the current directory.
Use "--yaml" to pass a file or "--from-store" to generate using function store.`)
		os.Exit(1)
	}

	objectsString, err := generateCRDYAML(services, tagFormat, api, crdFunctionNamespace,
		builder.NewFunctionMetadataSourceLive())
	if err != nil {
		return err
	}

	if len(objectsString) > 0 {
		fmt.Println(objectsString)
	}
	return nil
}

// generateCRDYAML generates CRD YAML for functions
func generateCRDYAML(services stack.Services, format schema.BuildFormat, apiVersion, namespace string, metadataSource builder.FunctionMetadataSource) (string, error) {

	var objectsString string

	if len(services.Functions) > 0 {

		if apiVersion == knativev1.APIVersionLatest {
			return generateknativev1ServingServiceCRDYAML(services, format, api, crdFunctionNamespace)
		}

		orderedNames := generateFunctionOrder(services.Functions)

		for _, name := range orderedNames {

			function := services.Functions[name]
			//read environment variables from the file
			fileEnvironment, err := readFiles(function.EnvironmentFile)
			if err != nil {
				return "", err
			}

			// combine all environment variables
			allEnvironment, envErr := compileEnvironment([]string{}, function.Environment, fileEnvironment)
			if envErr != nil {
				return "", envErr
			}

			branch, version, err := metadataSource.Get(tagFormat, function.Handler)
			if err != nil {
				return "", err
			}

			metadata := schema.Metadata{Name: name, Namespace: namespace}
			imageName := schema.BuildImageName(format, function.Image, version, branch)

			spec := openfaasv1.Spec{
				Name:                   name,
				Image:                  imageName,
				Environment:            allEnvironment,
				Labels:                 function.Labels,
				Annotations:            function.Annotations,
				Limits:                 function.Limits,
				Requests:               function.Requests,
				Constraints:            function.Constraints,
				Secrets:                function.Secrets,
				ReadOnlyRootFilesystem: function.ReadOnlyRootFilesystem,
			}

			crd := openfaasv1.CRD{
				APIVersion: apiVersion,
				Kind:       resourceKind,
				Metadata:   metadata,
				Spec:       spec,
			}

			var buff bytes.Buffer
			yamlEncoder := yaml.NewEncoder(&buff)
			yamlEncoder.SetIndent(2) // this is what you're looking for
			if err := yamlEncoder.Encode(&crd); err != nil {
				return "", err
			}

			objectString := buff.String()

			objectsString += "---\n" + string(objectString)
		}
	}

	return objectsString, nil
}

func generateknativev1ServingServiceCRDYAML(services stack.Services, format schema.BuildFormat, apiVersion, namespace string) (string, error) {
	crds := []knativev1.ServingServiceCRD{}

	orderedNames := generateFunctionOrder(services.Functions)

	for _, name := range orderedNames {

		function := services.Functions[name]

		fileEnvironment, err := readFiles(function.EnvironmentFile)
		if err != nil {
			return "", err
		}

		//combine all environment variables
		allEnvironment, envErr := compileEnvironment([]string{}, function.Environment, fileEnvironment)
		if envErr != nil {
			return "", envErr
		}

		env := orderknativeEnv(allEnvironment)

		var annotations map[string]string

		if function.Annotations != nil {
			annotations = *function.Annotations
		}

		branch, version, err := builder.GetImageTagValues(tagFormat, function.Handler)
		if err != nil {
			return "", err
		}

		imageName := schema.BuildImageName(format, function.Image, version, branch)

		crd := knativev1.ServingServiceCRD{
			Metadata: schema.Metadata{
				Name:        name,
				Namespace:   namespace,
				Annotations: annotations,
			},
			APIVersion: apiVersion,
			Kind:       "Service",

			Spec: knativev1.ServingServiceSpec{
				ServingServiceSpecTemplate: knativev1.ServingServiceSpecTemplate{
					Template: knativev1.ServingServiceSpecTemplateSpec{
						Containers: []knativev1.ServingSpecContainersContainerSpec{},
					},
				},
			},
		}

		crd.Spec.Template.Containers = append(crd.Spec.Template.Containers, knativev1.ServingSpecContainersContainerSpec{
			Image: imageName,
			Env:   env,
		})

		var mounts []knativev1.VolumeMount
		var volumes []knativev1.Volume

		for _, secret := range function.Secrets {
			mounts = append(mounts, knativev1.VolumeMount{
				MountPath: "/var/openfaas/secrets/" + secret,
				ReadOnly:  true,
				Name:      secret,
			})
			volumes = append(volumes, knativev1.Volume{
				Name: secret,
				Secret: knativev1.Secret{
					SecretName: secret,
				},
			})
		}

		crd.Spec.Template.Volumes = volumes
		crd.Spec.Template.Containers[0].VolumeMounts = mounts

		crds = append(crds, crd)
	}

	var objectsString string
	for _, crd := range crds {

		var buff bytes.Buffer
		yamlEncoder := yaml.NewEncoder(&buff)
		yamlEncoder.SetIndent(2) // this is what you're looking for
		if err := yamlEncoder.Encode(&crd); err != nil {
			return "", err
		}

		objectsString += "---\n" + string(buff.Bytes())
	}

	return objectsString, nil
}

func generateFunctionOrder(functions map[string]stack.Function) []string {

	var functionNames []string

	for functionName := range functions {
		functionNames = append(functionNames, functionName)
	}

	sort.Strings(functionNames)

	return functionNames
}

func orderknativeEnv(environment map[string]string) []knativev1.EnvPair {

	var orderedEnvironment []string
	var envVars []knativev1.EnvPair

	for k := range environment {
		orderedEnvironment = append(orderedEnvironment, k)
	}

	sort.Strings(orderedEnvironment)

	for _, envVar := range orderedEnvironment {
		envVars = append(envVars, knativev1.EnvPair{Name: envVar, Value: environment[envVar]})
	}

	return envVars
}
