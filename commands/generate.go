// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/openfaas/faas-cli/builder"
	v2 "github.com/openfaas/faas-cli/schema/store/v2"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/schema"
	openfaasv1 "github.com/openfaas/faas-cli/schema/openfaas/v1"
	"github.com/openfaas/faas-cli/util"
	"github.com/openfaas/go-sdk/stack"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	yaml "gopkg.in/yaml.v3"
)

const (
	resourceKind      = "Function"
	defaultAPIVersion = "openfaas.com/v1"
)

var apiVersions = []string{"openfaas.com/v1"}

var (
	api                  string
	name                 string
	functionNamespace    string
	crdFunctionNamespace string
	fromStore            string
	desiredArch          string
	outputFormat         string
	annotationArgs       []string
	labelArgs            []string
	envArgs              []string
)

func init() {

	generateCmd.Flags().StringVar(&fromStore, "from-store", "", "generate using a store image")
	generateCmd.Flags().StringVar(&name, "name", "", "for use with --from-store, override the name for the generated function")
	generateCmd.Flags().StringVar(&outputFormat, "output", "", "output format e.g stack.yaml, for use with --from-store to generate an OpenFaaS stack.yaml")

	generateCmd.Flags().StringVar(&api, "api", defaultAPIVersion, "CRD API version e.g openfaas.com/v1")
	generateCmd.Flags().StringVarP(&crdFunctionNamespace, "namespace", "n", "openfaas-fn", "Kubernetes namespace for functions")
	generateCmd.Flags().Var(&tagFormat, "tag", "Override latest tag on function Docker image, accepts 'digest', 'latest', 'sha', 'branch', 'describe'")
	generateCmd.Flags().BoolVar(&envsubst, "envsubst", true, "Substitute environment variables in stack.yaml file")
	generateCmd.Flags().StringVar(&desiredArch, "arch", "x86_64", "Desired image arch. (Default x86_64)")
	generateCmd.Flags().StringArrayVar(&annotationArgs, "annotation", []string{}, "Any annotations you want to add (to store functions only)")
	generateCmd.Flags().StringArrayVar(&labelArgs, "label", []string{}, "Any labels you want to add (to store functions only)")
	generateCmd.Flags().StringArrayVarP(&envArgs, "env", "e", []string{}, "Set one or more environment variables (ENVVAR=VALUE) (to store functions only)")

	faasCmd.AddCommand(generateCmd)
}

var generateCmd = &cobra.Command{
	Use:   "generate --api=openfaas.com/v1 --yaml stack.yaml --tag sha --namespace=openfaas-fn",
	Short: "Generate Kubernetes CRD YAML file",
	Long:  `The generate command creates kubernetes CRD YAML file for functions`,
	Example: `  faas-cli generate --api=openfaas.com/v1 --yaml stack.yaml | kubectl apply  -f -
  faas-cli generate --api=openfaas.com/v1 -f stack.yaml
  faas-cli generate --api=openfaas.com/v1 --namespace openfaas-fn -f stack.yaml
  faas-cli generate --api=openfaas.com/v1 -f stack.yaml --tag branch -n openfaas-fn
  faas-cli generate --from-store nodeinfo --output stack.yaml > stack.yaml`,
	PreRunE: preRunGenerate,
	RunE:    runGenerate,
}

func preRunGenerate(cmd *cobra.Command, args []string) error {
	if outputFormat != "" && !isStackOutput(outputFormat) {
		return fmt.Errorf("unsupported output format %q, must be one of: stack.yaml, stack", outputFormat)
	}

	if isStackOutput(outputFormat) {
		if len(fromStore) == 0 {
			return fmt.Errorf("--output %s can only be used with --from-store", outputFormat)
		}

		return nil
	}

	if len(api) == 0 {
		return fmt.Errorf("you must supply the API version with the --api flag")
	}

	for _, version := range apiVersions {
		if api == version {
			return nil
		}
	}

	return fmt.Errorf("unsupported API version %q, must be one of: %s", api, strings.Join(apiVersions, ", "))
}

func isStackOutput(format string) bool {
	return format == "stack.yaml" || format == "stack"
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

	labels, err := util.ParseMap(labelArgs, "label")
	if err != nil {
		return fmt.Errorf("error parsing labels: %v", err)
	}

	envVars, err := util.ParseMap(envArgs, "env")
	if err != nil {
		return fmt.Errorf("error parsing environment variables: %v", err)
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

		allLabels := util.MergeMap(item.Labels, labels)

		allEnvironment := util.MergeMap(item.Environment, envVars)

		fullName := item.Name
		if len(name) > 0 {
			fullName = name
		}

		services.Functions[fullName] = stack.Function{
			Name:        fullName,
			Image:       item.Images[desiredArch],
			Labels:      &allLabels,
			Annotations: &allAnnotations,
			Environment: allEnvironment,
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
			`No "stack.yaml" or "stack.yml" file was found in the current directory.
Use "--yaml" / "-f" to specify a custom filename.

Alternatively, to generate a definition for store functions, use "--from-store"`)
		os.Exit(1)
	}

	if isStackOutput(outputFormat) {
		stackYAML, err := generateStackYAML(services, gateway)
		if err != nil {
			return err
		}

		fmt.Println(stackYAML)
		return nil
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

// stackYAML is the output format for --output stack.yaml, mirroring an
// OpenFaaS stack.yaml with only the values known from the store item
type stackYAML struct {
	Version   string                 `yaml:"version"`
	Provider  stack.Provider         `yaml:"provider"`
	Functions map[string]stackYAMLFn `yaml:"functions"`
}

type stackYAMLFn struct {
	Image       string             `yaml:"image"`
	SkipBuild   bool               `yaml:"skip_build"`
	Environment map[string]string  `yaml:"environment,omitempty"`
	Labels      *map[string]string `yaml:"labels,omitempty"`
	Annotations *map[string]string `yaml:"annotations,omitempty"`
}

// generateStackYAML generates an OpenFaaS stack.yaml for functions, with
// skip_build set so the pre-built image is deployed without a rebuild
func generateStackYAML(services stack.Services, gateway string) (string, error) {
	functions := make(map[string]stackYAMLFn, len(services.Functions))
	for name, function := range services.Functions {
		if function.Labels != nil && len(*function.Labels) == 0 {
			function.Labels = nil
		}

		if function.Annotations != nil && len(*function.Annotations) == 0 {
			function.Annotations = nil
		}

		functions[name] = stackYAMLFn{
			Image:       function.Image,
			SkipBuild:   true,
			Environment: function.Environment,
			Labels:      function.Labels,
			Annotations: function.Annotations,
		}
	}

	output := stackYAML{
		Version: "1.0",
		Provider: stack.Provider{
			Name:       "openfaas",
			GatewayURL: gateway,
		},
		Functions: functions,
	}

	var buff bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&buff)
	yamlEncoder.SetIndent(2)
	if err := yamlEncoder.Encode(&output); err != nil {
		return "", err
	}

	return buff.String(), nil
}

// generateCRDYAML generates CRD YAML for functions
func generateCRDYAML(services stack.Services, format schema.BuildFormat, apiVersion, namespace string, metadataSource builder.FunctionMetadataSource) (string, error) {

	var objectsString string

	if len(services.Functions) > 0 {

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

func generateFunctionOrder(functions map[string]stack.Function) []string {

	var functionNames []string

	for functionName := range functions {
		functionNames = append(functionNames, functionName)
	}

	sort.Strings(functionNames)

	return functionNames
}
