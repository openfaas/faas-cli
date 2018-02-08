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
	appendFile string
	list       bool
)

func init() {
	newFunctionCmd.Flags().StringVar(&language, "lang", "", "Language or template to use")
	newFunctionCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL to store in YAML stack file")

	newFunctionCmd.Flags().BoolVar(&list, "list", false, "List available languages")
	newFunctionCmd.Flags().StringVarP(&appendFile, "append", "a", "", "Append to existing YAML file")

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

	appendMode := len(appendFile) > 0
	if appendMode {
		if (strings.HasSuffix(appendFile, ".yml") || strings.HasSuffix(appendFile, ".yaml")) == false {
			return fmt.Errorf("when appending to a stack the suffix should be .yml or .yaml")
		}

		if _, statErr := os.Stat(appendFile); statErr != nil {
			return fmt.Errorf("unable to find file: %s - %s", appendFile, statErr.Error())
		}
	}

	if _, err := os.Stat(functionName); err == nil {
		return fmt.Errorf("folder: %s already exists", functionName)
	}

	if err := os.Mkdir(functionName, 0700); err == nil {
		fmt.Printf("Folder: %s created.\n", functionName)
	} else {
		return fmt.Errorf("folder: could not create %s : %s", functionName, err)
	}

	if err := updateGitignore(); err != nil {
		return fmt.Errorf("got unexpected error while updating .gitignore file: %s", err)
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
    image: ` + functionName + `
`

	printFiglet()
	fmt.Println()
	fmt.Printf("Function created in folder: %s\n", functionName)

	var stackWriteErr error

	if appendMode {
		originalBytes, readErr := ioutil.ReadFile(appendFile)
		if readErr != nil {
			fmt.Printf("unable to read %s to append, %s", appendFile, readErr)
		}
		buffer := string(originalBytes) + stackYaml

		stackWriteErr = ioutil.WriteFile(appendFile, []byte(buffer), 0600)
		if stackWriteErr != nil {
			return fmt.Errorf("error writing stack file %s", stackWriteErr)
		}

		fmt.Printf("Stack file updated: %s\n", appendFile)
	} else {

		stackWriteErr = ioutil.WriteFile("./"+functionName+".yml", []byte(stackYaml), 0600)
		if stackWriteErr != nil {
			return fmt.Errorf("error writing stack file %s", stackWriteErr)
		}

		fmt.Printf("Stack file written: %s\n", functionName+".yml")
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

// StrSort sort strings
type StrSort []string

func (a StrSort) Len() int           { return len(a) }
func (a StrSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a StrSort) Less(i, j int) bool { return a[i] < a[j] }
