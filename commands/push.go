// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"strings"
	"sync"

	"github.com/openfaas/faas-cli/exec"

	"github.com/morikuni/aec"
	"github.com/openfaas/faas-cli/builder"
	"github.com/openfaas/faas-cli/schema"
	"github.com/openfaas/faas-cli/stack"
	"github.com/spf13/cobra"
)

func init() {
	faasCmd.AddCommand(pushCmd)

	pushCmd.Flags().IntVar(&parallel, "parallel", 1, "Push images in parallel to depth specified.")
	pushCmd.Flags().Var(&tagFormat, "tag", "Override latest tag on function Docker image, accepts 'latest', 'sha', 'branch', 'describe'")
	pushCmd.Flags().BoolVar(&envsubst, "envsubst", true, "Substitute environment variables in stack.yml file")

}

// pushCmd handles pushing function container images to a remote repo
var pushCmd = &cobra.Command{
	Use:   `push -f YAML_FILE [--regex "REGEX"] [--filter "WILDCARD"] [--parallel] [--tag <sha|branch>]`,
	Short: "Push OpenFaaS functions to remote registry (Docker Hub)",
	Long: `Pushes the OpenFaaS function container image(s) defined in the supplied YAML
config to a remote repository.

These container images must already be present in your local image cache.`,

	Example: `  faas-cli push -f https://domain/path/myfunctions.yml
  faas-cli push -f ./stack.yml
  faas-cli push -f ./stack.yml --parallel 4
  faas-cli push -f ./stack.yml --filter "*gif*"
  faas-cli push -f ./stack.yml --regex "fn[0-9]_.*"
  faas-cli push -f ./stack.yml --tag sha
  faas-cli push -f ./stack.yml --tag branch
  faas-cli push -f ./stack.yml --tag describe`,
	RunE: runPush,
}

func runPush(cmd *cobra.Command, args []string) error {

	var services stack.Services
	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAMLFile(yamlFile, regex, filter, envsubst)
		if err != nil {
			return err
		}

		if parsedServices != nil {
			services = *parsedServices
		}
	}

	if len(services.Functions) > 0 {
		invalidImages := validateImages(services.Functions)
		if len(invalidImages) > 0 {
			imageList := strings.Join(invalidImages, "\n- ")
			return fmt.Errorf(`
Unable to push one or more of your functions to Docker Hub:
- ` + imageList + `

You must provide a username or registry prefix to the Function's image such as user1/function1`)
		}

		pushStack(&services, parallel, tagFormat)
	} else {
		return fmt.Errorf("you must supply a valid YAML file")
	}
	return nil
}

func pushImage(image string) {
	exec.Command("./", []string{"docker", "push", image})
}

func pushStack(services *stack.Services, queueDepth int, tagMode schema.BuildFormat) {
	wg := sync.WaitGroup{}

	workChannel := make(chan stack.Function)

	wg.Add(queueDepth)
	for i := 0; i < queueDepth; i++ {
		go func(index int) {
			for function := range workChannel {
				branch, sha, err := builder.GetImageTagValues(tagMode)
				if err != nil {
					tagMode = schema.DefaultFormat
				}
				imageName := schema.BuildImageName(tagMode, function.Image, sha, branch)

				fmt.Printf(aec.YellowF.Apply("[%d] > Pushing %s [%s].\n"), index, function.Name, imageName)
				if len(function.Image) == 0 {
					fmt.Println("Please provide a valid Image value in the YAML file.")
				} else if function.SkipBuild {
					fmt.Printf("Skipping %s\n", function.Name)
				} else {

					pushImage(imageName)
					fmt.Printf(aec.YellowF.Apply("[%d] < Pushing %s [%s] done.\n"), index, function.Name, imageName)
				}
			}

			fmt.Printf(aec.YellowF.Apply("[%d] Worker done.\n"), index)
			wg.Done()
		}(i)
	}

	for k, function := range services.Functions {
		function.Name = k
		workChannel <- function
	}

	close(workChannel)

	wg.Wait()

}

func validateImages(functions map[string]stack.Function) []string {
	invalidImages := []string{}

	for name, function := range functions {

		if !function.SkipBuild && !strings.Contains(function.Image, `/`) {
			invalidImages = append(invalidImages, name)
		}
	}
	return invalidImages
}
