// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/morikuni/aec"
	"github.com/openfaas/faas-cli/builder"
	"github.com/openfaas/faas-cli/stack"
	"github.com/spf13/cobra"
)

var (
	lang string
	list bool
)

// Implement interface for sorting array of strings
type StrSort []string

func (a StrSort) Len() int           { return len(a) }
func (a StrSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a StrSort) Less(i, j int) bool { return a[i] < a[j] }

func init() {
	newFunctionCmd.Flags().StringVar(&lang, "lang", "", "Language or template to use")
	newFunctionCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL to store in YAML stack file")

	newFunctionCmd.Flags().BoolVar(&list, "list", false, "List available languages")

	faasCmd.AddCommand(newFunctionCmd)
}

// newFunctionCmd displays newFunction information
var newFunctionCmd = &cobra.Command{
	Use:   "new FUNCTION_NAME --lang=FUNCTION_LANGUAGE [--gateway=http://domain:port] | --list)",
	Short: "Create a new template in the current folder with the name given as name",
	Long: `The new command creates a new function based upon hello-world in the given
language or type in --list for a list of languages available.`,
	Example: `faas-cli new chatbot --lang node
  faas-cli new textparser --lang python --gateway http://mydomain:8080
  faas-cli new --list`,
	RunE: runNewFunction,
}

func runNewFunction(cmd *cobra.Command, args []string) error {
	if list == true {
		var availableTemplates []string

		if templateFolders, err := ioutil.ReadDir(templateDirectory); err != nil {
			return fmt.Errorf("no language templates were found. Please run 'faas-cli template pull'")
		} else {
			for _, file := range templateFolders {
				if file.IsDir() {
					availableTemplates = append(availableTemplates, file.Name())
				}
			}
		}

		fmt.Printf(`Languages available as templates:
` + printAvailableTemplates(availableTemplates) + `

Or alternatively create a folder containing a Dockerfile, then pick
the "Dockerfile" lang type in your YAML file.
`)
		return nil
	}

	if len(args) < 1 {
		return fmt.Errorf("please provide a name for the function")
	}
	functionName = args[0]

	if len(lang) == 0 {
		return fmt.Errorf("you must supply a function language with the --lang flag")
	}

	PullTemplates("")

	if stack.IsValidTemplate(lang) == false {
		return fmt.Errorf("%s is unavailable or not supported", lang)
	}

	if _, err := os.Stat(functionName); err == nil {
		return fmt.Errorf("folder: %s already exists", functionName)
	}

	if err := os.Mkdir("./"+functionName, 0700); err == nil {
		fmt.Printf("Folder: %s created.\n", functionName)
	}

	if err := updateGitignore(); err != nil {
		return fmt.Errorf("got unexpected error while updating .gitignore file: %s", err)
	}

	// Only "template" language templates - Dockerfile must be custom, so start with empty directory.
	if strings.ToLower(lang) != "dockerfile" {
		builder.CopyFiles("./template/"+lang+"/function/", "./"+functionName+"/", true)
	} else {
		ioutil.WriteFile("./"+functionName+"/Dockerfile", []byte(`FROM alpine:3.6
# Use any image as your base image, or "scratch"
# Add fwatchdog binary via https://github.com/openfaas/faas/releases/
# Then set fprocess to the process you want to invoke per request - i.e. "cat" or "my_binary"

ADD https://github.com/openfaas/faas/releases/download/0.6.9/fwatchdog /usr/bin
# COPY ./fwatchdog /usr/bin/
RUN chmod +x /usr/bin/fwatchdog

# Populate example here - i.e. "cat", "sha512sum" or "node index.js"
ENV fprocess="wc -l"

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1
CMD ["fwatchdog"]
`), 0600)
	}

	stack := `provider:
  name: faas
  gateway: ` + gateway + `

functions:
  ` + functionName + `:
    lang: ` + lang + `
    handler: ./` + functionName + `
    image: ` + functionName + `
`

	fmt.Printf(aec.BlueF.Apply(figletStr))
	fmt.Println()
	fmt.Printf("Function created in folder: %s\n", functionName)

	stackWriteErr := ioutil.WriteFile("./"+functionName+".yml", []byte(stack), 0600)
	if stackWriteErr != nil {
		return fmt.Errorf("error writing stack file %s", stackWriteErr)
	}

	fmt.Printf("Stack file written: %s\n", functionName+".yml")
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
