// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/openfaas/faas-cli/builder"
	"github.com/spf13/cobra"
)

var (
	lang string
	list bool
)

func init() {
	newFunctionCmd.Flags().StringVar(&functionName, "name", "", "Name for your function")
	newFunctionCmd.Flags().StringVar(&lang, "lang", "", "Language or template to use")
	newFunctionCmd.Flags().StringVar(&gateway, "gateway", defaultGateway,
		"Gateway URL to store in YAML stack file")

	newFunctionCmd.Flags().BoolVar(&list, "list", false, "List available languages")

	faasCmd.AddCommand(newFunctionCmd)
}

// newFunctionCmd displays newFunction information
var newFunctionCmd = &cobra.Command{
	Use:   "new (--name=FUNCTION_NAME --lang=FUNCTION_LANGUAGE [--gateway=http://domain:port] | --list)",
	Short: "Create a new template in the current folder with the name given as name",
	Long: `The new command creates a new function based upon hello-world in the given
language or type in --list for a list of languages available.`,
	Example: `faas-cli new --name chatbot --lang node
  faas-cli new --name textparser --lang python --gateway http://mydomain:8080
  faas-cli new --list`,
	Run: runNewFunction,
}

func runNewFunction(cmd *cobra.Command, args []string) {
	if list == true {
		fmt.Printf(`Languages available as templates:
- node
- python
- ruby
- csharp

Or alternatively create a folder and a new Dockerfile, then pick
the "Dockerfile" lang type in your YAML file.
`)
		return
	}
	if len(functionName) == 0 {
		fmt.Println("You must supply a function name with the --name flag")
		return
	}

	if len(lang) == 0 {
		fmt.Println("You must supply a function language with the --lang flag")
		return
	}

	PullTemplates("")

	if _, err := os.Stat(functionName); err == nil {
		fmt.Printf("Folder: %s already exists\n", functionName)
		return
	}

	if err := os.Mkdir("./"+functionName, 0700); err == nil {
		fmt.Printf("Folder: %s created.\n", functionName)
	}

	builder.CopyFiles("./template/"+lang+"/function/", "./"+functionName+"/", true)

	stack := `provider:
  name: faas
  gateway: ` + gateway + `

functions:
  ` + functionName + `:
    lang: ` + lang + `
    handler: ./` + functionName + `
    image: ` + functionName + `
`

	fmt.Printf(figletStr)
	fmt.Println()
	fmt.Printf("Function created in folder: %s\n", functionName)

	stackWriteErr := ioutil.WriteFile("./"+functionName+".yml", []byte(stack), 0600)
	if stackWriteErr != nil {
		fmt.Printf("Error writing stack file %s\n", stackWriteErr)
	} else {
		fmt.Printf("Stack file written: %s\n", functionName+".yml")
	}

	return
}
