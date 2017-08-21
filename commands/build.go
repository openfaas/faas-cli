// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/alexellis/faas-cli/builder"
	"github.com/alexellis/faas-cli/stack"
	"github.com/spf13/cobra"
)

// Flags that are to be added to commands.
var (
	nocache bool
	squash  bool
)

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	buildCmd.Flags().StringVar(&image, "image", "", "Docker image name to build")
	buildCmd.Flags().StringVar(&handler, "handler", "", "Directory with handler for function, e.g. handler.js")
	buildCmd.Flags().StringVar(&functionName, "name", "", "Name of the deployed function")
	buildCmd.Flags().StringVar(&language, "lang", "node", "Programming language template")

	// Setup flags that are used only by this command (variables defined above)
	buildCmd.Flags().BoolVar(&nocache, "no-cache", false, "Do not use Docker's build cache")
	buildCmd.Flags().BoolVar(&squash, "squash", false, `Use Docker's squash flag for smaller images
                         [experimental] `)

	// Set bash-completion.
	_ = buildCmd.Flags().SetAnnotation("handler", cobra.BashCompSubdirsInDir, []string{})

	faasCmd.AddCommand(buildCmd)
}

// buildCmd allows the user to build an OpenFaaS function container
var buildCmd = &cobra.Command{
	Use: `build -f YAML_FILE [--no-cache] [--squash]
  faas-cli build --image IMAGE_NAME
                 --handler HANDLER_DIR
                 --name FUNCTION_NAME
                 [--lang <ruby|python|python-armf|node|node-armf|csharp>]
                 [--no-cache] [--squash]`,
	Short: "Builds OpenFaaS function containers",
	Long: `Builds OpenFaaS function containers either via the supplied YAML config using
the "--yaml" flag (which may contain multiple function definitions), or directly
via flags.`,
	Example: `  faas-cli build -f https://domain/path/myfunctions.yml
  faas-cli build -f ./samples.yml --no-cache
  faas-cli build --image=my_image --lang=python --handler=/path/to/fn/ 
                 --name=my_fn --squash`,
	Run: runBuild,
}

func runBuild(cmd *cobra.Command, args []string) {

	var services stack.Services
	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAML(yamlFile)
		if err != nil {
			log.Fatalln(err.Error())
			return
		}

		if parsedServices != nil {
			services = *parsedServices
		}
	}

	if pullErr := pullTemplates(); pullErr != nil {
		log.Fatalln("Could not pull templates for FaaS.", pullErr)
	}

	if len(services.Functions) > 0 {
		for k, function := range services.Functions {
			if function.SkipBuild {
				fmt.Printf("Skipping build of: %s.\n", function.Name)
			} else {
				function.Name = k
				// fmt.Println(k, function)
				fmt.Printf("Building: %s.\n", function.Name)
				builder.BuildImage(function.Image, function.Handler, function.Name, function.Language, nocache, squash)
			}
		}
	} else {
		if len(image) == 0 {
			fmt.Println("Please provide a valid -image name for your Docker image.")
			return
		}
		if len(handler) == 0 {
			fmt.Println("Please provide the full path to your function's handler.")
			return
		}
		if len(functionName) == 0 {
			fmt.Println("Please provide the deployed -name of your function.")
			return
		}
		builder.BuildImage(image, handler, functionName, language, nocache, squash)
	}

}

func pullTemplates() error {
	var err error
	exists, err := os.Stat("./template")
	if err != nil || exists == nil {
		log.Println("No templates found in current directory.")

		err = fetchTemplates()
		if err != nil {
			log.Println("Unable to download templates from Github.")
			return err
		}
	}
	return err
}
