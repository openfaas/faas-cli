// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/morikuni/aec"
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
	Use:   "new FUNCTION_NAME --lang=FUNCTION_LANGUAGE [--gateway=http://domain:port] | --list)",
	Short: "Create a new template in the current folder with the name given as name",
	Long: `The new command creates a new function based upon hello-world in the given
language or type in --list for a list of languages available.`,
	Example: `faas-cli new chatbot --lang node
  faas-cli new textparser --lang python --gateway http://mydomain:8080
  faas-cli new --list`,
	Run: runNewFunction,
}

func runNewFunction(cmd *cobra.Command, args []string) {
	if list == true {
		fmt.Printf(`Languages available as templates:
- node
- python
- python3
- ruby
- csharp
- Dockerfile
- go

Or alternatively create a folder containing a Dockerfile, then pick
the "Dockerfile" lang type in your YAML file.
`)
		return
	}

	if len(args) < 1 {
		fmt.Println("Please provide a name for the function")
		return
	}
	functionName = args[0]

	if len(lang) == 0 {
		fmt.Println("You must supply a function language with the --lang flag")
		return
	}

	PullTemplates("")

	if validTemplate(lang) == false {
		fmt.Printf("%s is unavailable or not supported.\n", lang)
	}

	if _, err := os.Stat(functionName); err == nil {
		fmt.Printf("Folder: %s already exists\n", functionName)
		return
	}

	if err := os.Mkdir("./"+functionName, 0700); err == nil {
		fmt.Printf("Folder: %s created.\n", functionName)
	}

	if err := updateGitignore(); err != nil {
		fmt.Println("Got unexpected error while updating .gitignore file.")
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
		fmt.Printf("Error writing stack file %s\n", stackWriteErr)
	} else {
		fmt.Printf("Stack file written: %s\n", functionName+".yml")
	}

	return
}

func validTemplate(lang string) bool {
	var found bool
	if strings.ToLower(lang) != "dockerfile" {
		found = true
	}
	if _, err := os.Stat(path.Join("./template/", lang)); err == nil {
		found = true
	}

	return found
}
