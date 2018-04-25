// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/openfaas/faas-cli/builder"
	"github.com/openfaas/faas-cli/stack"
	"github.com/spf13/cobra"
)

var (
	list bool
)

func init() {
	newFunctionCmd.Flags().StringVar(&language, "lang", "", "Language or template to use")
	newFunctionCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL to store in YAML stack file")
	newFunctionCmd.Flags().StringVarP(&imagePrefix, "prefix", "p", "", "Set prefix for the function image")

	newFunctionCmd.Flags().BoolVar(&list, "list", false, "List available languages")

	faasCmd.AddCommand(newFunctionCmd)
}

// newFunctionCmd displays newFunction information
var newFunctionCmd = &cobra.Command{
	Use:   "new FUNCTION_NAME --lang=FUNCTION_LANGUAGE [--gateway=http://domain:port] | --list | --yaml=YAML_FILE)",
	Short: "Create a new template in the current folder with the name given as name",
	Long: `The new command creates a new function based upon hello-world in the given language
or type in --list for a list of languages available.  YAML file can be specified by
--yaml flag ( default: <func_name>.yml ). If the stack file already exists then new 
function will be appended to the existing file`,
	Example: `faas-cli new chatbot --lang node
  faas-cli new text-parser --lang python --gateway http://mydomain:8080
  faas-cli new text-reader --lang python --yaml stack.yml
  faas-cli new --list`,
	PreRunE: preRunNewFunction,
	RunE:    runNewFunction,
}

// preRunNewFunction validates args & flags
func preRunNewFunction(cmd *cobra.Command, args []string) error {
	language, _ = validateLanguageFlag(language)

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

	if len(args) < 1 {
		return fmt.Errorf("please provide a name for the function")
	}

	functionName = args[0]

	if len(language) == 0 {
		return fmt.Errorf("you must supply a function language with the --lang flag")
	}

	PullTemplates(DefaultTemplateRepository)

	if stack.IsValidTemplate(language) == false {
		return fmt.Errorf("%s is unavailable or not supported", language)
	}

	appendMode := false
	if len(yamlFile) > 0 {
		if (yamlFile == defaultYAML) && usingDefaultYaml {
			yamlFile = "./" + functionName + ".yml"
		}

		if (strings.HasSuffix(yamlFile, ".yml") || strings.HasSuffix(yamlFile, ".yaml")) == false {
			return fmt.Errorf("the stack file suffix should be .yml or .yaml")
		}
	} else {
		yamlFile = "./" + functionName + ".yml"
	}
	if _, statErr := os.Stat(yamlFile); statErr == nil {
		appendMode = true
	} else if !os.IsNotExist(statErr) {
		return fmt.Errorf("unable to access stack file %s, error %v", yamlFile, statErr)
	}
	if appendMode {
		exists, readErr := checkIfFuncExist(functionName, yamlFile)
		if readErr != nil {
			return readErr
		}
		if exists {
			return fmt.Errorf(`
Function %s already exists in %s file.
Cannot have duplicate function names in same yml file`, functionName, yamlFile)
		}
	}

	if _, err := os.Stat(functionName); err == nil {
		return fmt.Errorf("folder: %s already exists", functionName)
	}

	if err := os.Mkdir(functionName, 0700); err == nil {
		fmt.Printf("Folder: %s is created.\n", functionName)
	} else {
		return fmt.Errorf("folder: %s failed to create: %s", functionName, err)
	}

	if err := updateGitignore(); err != nil {
		return fmt.Errorf("got unexpected error while updating .gitignore file: %s", err)
	}

	var imageName string
	imagePrefix = strings.TrimSpace(imagePrefix)
	if len(imagePrefix) > 0 {
		imageName = imagePrefix + "/" + functionName
	} else {
		imageName = functionName
	}

	builder.CopyFiles(filepath.Join("template", language, "function"), functionName)

	var stackYaml string

	if !appendMode {
		stackYaml +=
			`provider:
  name: faas
  gateway: ` + gateway + `

functions:
`
	}

	stackYaml +=
		`  ` + functionName + `:
    lang: ` + language + `
    handler: ./` + functionName + `
    image: ` + imageName + `
`

	printFiglet()
	fmt.Println()
	fmt.Printf("Function created in folder: %s\n", functionName)

	var stackWriteErr error

	if appendMode {
		originalBytes, readErr := ioutil.ReadFile(yamlFile)
		if readErr != nil {
			fmt.Printf("unable to read %s to append, %s", yamlFile, readErr)
		}

		buffer := string(originalBytes) + stackYaml

		stackWriteErr = ioutil.WriteFile(yamlFile, []byte(buffer), 0600)
		if stackWriteErr != nil {
			return fmt.Errorf("error writing stack file %s", stackWriteErr)
		}

		fmt.Printf("Stack file updated: %s\n", yamlFile)
	} else {
		stackWriteErr = ioutil.WriteFile(yamlFile, []byte(stackYaml), 0600)
		if stackWriteErr != nil {
			return fmt.Errorf("error writing stack file %s", stackWriteErr)
		}

		fmt.Printf("Stack file written: %s\n", yamlFile)
	}

	return nil
}

func printAvailableTemplates(availableTemplates []string) string {
	var result string
	sort.Sort(StrSort(availableTemplates))
	for _, template := range availableTemplates {
		result += fmt.Sprintf("- %s\n", template)
	}
	return result
}

func checkIfFuncExist(functionName string, yamlFile string) (bool, error) {
	fileBytes, readErr := ioutil.ReadFile(yamlFile)
	if readErr != nil {
		return false, fmt.Errorf("unable to read %s to append, %s", yamlFile, readErr)
	}

	services, parseErr := stack.ParseYAMLData(fileBytes, "", "")

	if parseErr != nil {
		return false, fmt.Errorf("Error parsing %s yml file", yamlFile)
	}

	if _, ok := services.Functions[functionName]; ok {
		return true, nil
	}

	return false, nil
}

// StrSort sort strings
type StrSort []string

func (a StrSort) Len() int           { return len(a) }
func (a StrSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a StrSort) Less(i, j int) bool { return a[i] < a[j] }
