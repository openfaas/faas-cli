// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/openfaas/faas-cli/builder"
	"github.com/openfaas/faas-cli/stack"
	"github.com/spf13/cobra"
)

// Flags that are to be added to commands.
var (
	nocache    bool
	squash     bool
	parallel   int
	shrinkwrap bool
)

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	buildCmd.Flags().StringVar(&image, "image", "", "Docker image name to build")
	buildCmd.Flags().StringVar(&handler, "handler", "", "Directory with handler for function, e.g. handler.js")
	buildCmd.Flags().StringVar(&functionName, "name", "", "Name of the deployed function")
	buildCmd.Flags().StringVar(&language, "lang", "", "Programming language template")

	// Setup flags that are used only by this command (variables defined above)
	buildCmd.Flags().BoolVar(&nocache, "no-cache", false, "Do not use Docker's build cache")
	buildCmd.Flags().BoolVar(&squash, "squash", false, `Use Docker's squash flag for smaller images
						 [experimental] `)
	buildCmd.Flags().IntVar(&parallel, "parallel", 1, "Build in parallel to depth specified.")

	buildCmd.Flags().BoolVar(&shrinkwrap, "shrinkwrap", false, "Just write files to ./build/ folder for shrink-wrapping")

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
                 [--lang <ruby|python|python3|node|csharp|Dockerfile>]
                 [--no-cache] [--squash]
                 [--regex "REGEX"]
				 [--filter "WILDCARD"]
				 [--parallel PARALLEL_DEPTH]`,
	Short: "Builds OpenFaaS function containers",
	Long: `Builds OpenFaaS function containers either via the supplied YAML config using
the "--yaml" flag (which may contain multiple function definitions), or directly
via flags.`,
	Example: `  faas-cli build -f https://domain/path/myfunctions.yml
  faas-cli build -f ./samples.yml --no-cache
  faas-cli build -f ./samples.yml --filter "*gif*"
  faas-cli build -f ./samples.yml --regex "fn[0-9]_.*"
  faas-cli build --image=my_image --lang=python --handler=/path/to/fn/ 
                 --name=my_fn --squash`,
	RunE: runBuild,
}

func runBuild(cmd *cobra.Command, args []string) error {

	var services stack.Services
	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAMLFile(yamlFile, regex, filter)
		if err != nil {
			return err
		}

		if parsedServices != nil {
			services = *parsedServices
		}
	}

	if pullErr := PullTemplates(""); pullErr != nil {
		return fmt.Errorf("could not pull templates for OpenFaaS: %v", pullErr)
	}

	if len(services.Functions) > 0 {
		build(&services, parallel, shrinkwrap)
	} else {
		if len(image) == 0 {
			return fmt.Errorf("please provide a valid --image name for your Docker image")
		}
		if len(handler) == 0 {
			return fmt.Errorf("please provide the full path to your function's handler")
		}
		if len(functionName) == 0 {
			return fmt.Errorf("please provide the deployed --name of your function")
		}
		builder.BuildImage(image, handler, functionName, language, nocache, squash, shrinkwrap)
	}

	return nil
}

func build(services *stack.Services, queueDepth int, shrinkwrap bool) {
	wg := sync.WaitGroup{}

	workChannel := make(chan stack.Function)

	for i := 0; i < queueDepth; i++ {

		go func(index int) {
			wg.Add(1)
			for function := range workChannel {
				fmt.Printf("[%d] > Building: %s.\n", index, function.Name)
				if len(function.Language) == 0 {
					fmt.Println("Please provide a valid --lang or 'Dockerfile' for your function.")

				} else {
					builder.BuildImage(function.Image, function.Handler, function.Name, function.Language, nocache, squash, shrinkwrap)
				}
			}

			fmt.Printf("[%d] < Builder done.\n", index)
			wg.Done()
		}(i)
	}

	for k, function := range services.Functions {
		if function.SkipBuild {
			fmt.Printf("Skipping build of: %s.\n", function.Name)
		} else {
			function.Name = k
			workChannel <- function
		}
	}

	close(workChannel)

	wg.Wait()

}

// PullTemplates pulls templates from Github from the master zip download file.
func PullTemplates(templateUrl string) error {
	var err error
	exists, err := os.Stat("./template")
	if err != nil || exists == nil {
		log.Println("No templates found in current directory.")

		err = fetchTemplates(templateUrl, false)
		if err != nil {
			log.Println("Unable to download templates from Github.")
			return err
		}
	}
	return err
}
