// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/openfaas/faas-cli/builder"
	"github.com/openfaas/faas-cli/stack"
	"github.com/spf13/cobra"
)

var (
	appendFile string
	list       bool
	quiet      bool
)

func init() {
	newFunctionCmd.Flags().StringVar(&language, "lang", "", "Language or template to use")
	newFunctionCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL to store in YAML stack file")
	newFunctionCmd.Flags().StringVarP(&imagePrefix, "prefix", "p", "", "Set prefix for the function image")

	newFunctionCmd.Flags().BoolVar(&list, "list", false, "List available languages")
	newFunctionCmd.Flags().StringVarP(&appendFile, "append", "a", "", "Append to existing YAML file")
	newFunctionCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Skip template notes")

	faasCmd.AddCommand(newFunctionCmd)
}

// newFunctionCmd displays newFunction information
var newFunctionCmd = &cobra.Command{
	Use:   "new FUNCTION_NAME --lang=FUNCTION_LANGUAGE [--gateway=http://domain:port] | --list | --append=STACK_FILE)",
	Short: "Create a new template in the current folder with the name given as name",
	Long: `The new command creates a new function based upon hello-world in the given
language or type in --list for a list of languages available.`,
	Example: `faas-cli new chatbot --lang node
  faas-cli new text-parser --lang python --gateway http://mydomain:8080
  faas-cli new text-reader --lang python --append stack.yml
  faas-cli new --list
  faas-cli new demo --lang python --quiet`,
	PreRunE: preRunNewFunction,
	RunE:    runNewFunction,
}

// validateFunctionName provides least-common-denominator validation - i.e. only allows valid Kubernetes services names
func validateFunctionName(functionName string) error {
	// Regex for RFC-1123 validation:
	// 	k8s.io/kubernetes/pkg/util/validation/validation.go
	var validDNS = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	if matched := validDNS.MatchString(functionName); !matched {
		return fmt.Errorf(`function name can only contain a-z, 0-9 and dashes`)
	}
	return nil
}

// preRunNewFunction validates args & flags
func preRunNewFunction(cmd *cobra.Command, args []string) error {
	if list == true {
		return nil
	}

	language, _ = validateLanguageFlag(language)

	if len(args) < 1 {
		return fmt.Errorf("please provide a name for the function")
	}
	if len(language) == 0 {
		return fmt.Errorf("you must supply a function language with the --lang flag")
	}

	functionName = args[0]

	if err := validateFunctionName(functionName); err != nil {
		return err
	}

	return nil
}

func runNewFunction(cmd *cobra.Command, args []string) error {
	if list == true {
		var availableTemplates []string

		templateFolders, err := ioutil.ReadDir(templateDirectory)
		if err != nil {
			return fmt.Errorf("no language templates were found. Please run 'faas-cli template pull'")
		}

		for _, file := range templateFolders {
			if file.IsDir() {
				availableTemplates = append(availableTemplates, file.Name())
			}
		}

		fmt.Printf("Languages available as templates:\n%s\n", printAvailableTemplates(availableTemplates))

		return nil
	}

	PullTemplates(DefaultTemplateRepository)

	if !stack.IsValidTemplate(language) {
		return fmt.Errorf("%s is unavailable or not supported", language)
	}

	var services *stack.Services
	appendMode := len(appendFile) > 0
	if appendMode {
		if (strings.HasSuffix(appendFile, ".yml") || strings.HasSuffix(appendFile, ".yaml")) == false {
			return fmt.Errorf("when appending to a stack the suffix should be .yml or .yaml")
		}

		if _, statErr := os.Stat(appendFile); statErr != nil {
			return fmt.Errorf("unable to find file: %s - %s", appendFile, statErr.Error())
		}

		var duplicateError error
		services, duplicateError = duplicateFunctionName(functionName, appendFile)

		if duplicateError != nil {
			return duplicateError
		}
	} else {
		gateway = getGatewayURL(gateway, defaultGateway, gateway, os.Getenv(openFaaSURLEnvironment))
		services = &stack.Services{
			Provider: stack.Provider{
				Name:       "faas",
				GatewayURL: gateway,
			},
			Functions: make(map[string]stack.Function),
		}
	}

	if _, err := os.Stat(functionName); err == nil {
		return fmt.Errorf("folder: %s already exists", functionName)
	}

	if err := os.Mkdir(functionName, 0700); err != nil {
		return fmt.Errorf("folder: could not create %s : %s", functionName, err)
	}
	fmt.Printf("Folder: %s created.\n", functionName)

	if err := updateGitignore(); err != nil {
		return fmt.Errorf("got unexpected error while updating .gitignore file: %s", err)
	}

	// Create function directory from template.
	builder.CopyFiles(filepath.Join("template", language, "function"), functionName)
	printFiglet()
	fmt.Printf("\nFunction created in folder: %s\n", functionName)

	// Define template of stack file.
	const stackTmpl = `{{ if .Provider.Name -}}
provider:
  name: {{ .Provider.Name }}
  gateway: {{ .Provider.GatewayURL }}

functions:
{{- end }}
{{- range $name, $function := .Functions }}
  {{ $name }}:
    lang: {{ $function.Language }}
    handler: ./{{ $name }}
    image: {{ $function.Image }}
{{- end }}
`

	var imageName string
	if imagePrefix = strings.TrimSpace(imagePrefix); len(imagePrefix) > 0 {
		imageName = fmt.Sprintf("%s/%s:latest", imagePrefix, functionName)
	} else {
		imageName = fmt.Sprintf("%s:latest", functionName)
	}

	function := stack.Function{
		Name:     functionName,
		Language: language,
		Image:    imageName,
	}
	services.Functions[functionName] = function

	var fileName string
	if appendMode {
		fileName = appendFile
	} else {
		fileName = functionName + ".yml"
	}
	f, err := os.OpenFile("./"+fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("could not open file '%s' %s", fileName, err)
	}

	t := template.Must(template.New("stack").Parse(stackTmpl))
	if err := t.Execute(f, services); err != nil {
		return fmt.Errorf("could not parse functions into stack template %s", err)
	}

	if appendMode {
		fmt.Printf("Stack file updated: %s\n", fileName)
	} else {
		fmt.Printf("Stack file written: %s\n", fileName)
	}

	if !quiet {
		languageTemplate, _ := stack.LoadLanguageTemplate(language)

		if languageTemplate.WelcomeMessage != "" {
			fmt.Printf("\nNotes:\n")
			fmt.Printf("%s\n", languageTemplate.WelcomeMessage)
		}
	}

	return nil
}

func printAvailableTemplates(availableTemplates []string) string {
	var result string
	sort.Slice(availableTemplates, func(i, j int) bool {
		return availableTemplates[i] < availableTemplates[j]
	})
	for _, template := range availableTemplates {
		result += fmt.Sprintf("- %s\n", template)
	}
	return result
}

func duplicateFunctionName(functionName string, appendFile string) (*stack.Services, error) {
	fileBytes, readErr := ioutil.ReadFile(appendFile)
	if readErr != nil {
		return nil, fmt.Errorf("unable to read %s to append, %s", appendFile, readErr)
	}

	services, parseErr := stack.ParseYAMLData(fileBytes, "", "")

	if parseErr != nil {
		return nil, fmt.Errorf("Error parsing %s yml file", appendFile)
	}

	if _, ok := services.Functions[functionName]; ok {
		return nil, fmt.Errorf(`
Function %s already exists in %s file. 
Cannot have duplicate function names in same yml file`, functionName, appendFile)
	}

	return services, nil
}
