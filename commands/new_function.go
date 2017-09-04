// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/alexellis/faas-cli/builder"
	"github.com/spf13/cobra"
)

var (
	lang string
	list bool
)

func init() {
	newFunctionCmd.Flags().StringVar(&functionName, "name", "", "Name for your function")
	newFunctionCmd.Flags().StringVar(&lang, "lang", "", "Language or template to use")
	newFunctionCmd.Flags().StringVar(&gateway, "gateway", "http://localhost:8080",
		"Gateway URL or http://localhost:8080 to store in YAML stack file")

	newFunctionCmd.Flags().BoolVar(&list, "list", false, "List available languages")

	faasCmd.AddCommand(newFunctionCmd)
}

// newFunctionCmd displays newFunction information
var newFunctionCmd = &cobra.Command{
	Use:   "new [--name] [--lang] [--list]",
	Short: "Create a new template in the current folder with the name given as name",
	Long: fmt.Sprintf(`The new command creates a new function based upon hello-world in the
given language or type in --list for a list of languages available.`),
	Example: `  faas-cli new
  faas-cli new --name chatbot --lang node
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
		fmt.Printf("You must give a --function name\n")
		return
	}

	if len(lang) == 0 {
		fmt.Printf("You must give a --lang parameter\n")
		return
	}

	PullTemplates()

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

	fmt.Printf("Function created in folder: %s\n", functionName)

	stackWriteErr := ioutil.WriteFile("./"+functionName+".yml", []byte(stack), 0600)
	if stackWriteErr != nil {
		fmt.Printf("Error writing stack file %s\n", stackWriteErr)
	} else {
		fmt.Printf("Stack file written: %s\n", functionName+".yml")
	}

	return
}
