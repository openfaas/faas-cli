// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package builder

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/openfaas/faas-cli/stack"
)

// BuildImage construct Docker image from function parameters
func BuildImage(image string, handler string, functionName string, language string, nocache bool, squash bool, shrinkwrap bool, buildArgMap map[string]string) {

	if stack.IsValidTemplate(language) {

		var tempPath string
		if strings.ToLower(language) == "dockerfile" {

			if shrinkwrap {
				fmt.Printf("Nothing to do for: %s.\n", functionName)

				return
			}

			tempPath = handler
			if err := ensureHandlerPath(handler); err != nil {
				fmt.Printf("Unable to build %s, %s is an invalid path\n", image, handler)
				fmt.Printf("Image: %s not built.\n", image)

				return
			}
			fmt.Printf("Building: %s with Dockerfile. Please wait..\n", image)

		} else {

			if err := ensureHandlerPath(handler); err != nil {
				fmt.Printf("Unable to build %s, %s is an invalid path\n", image, handler)
				fmt.Printf("Image: %s not built.\n", image)

				return
			}
			tempPath = createBuildTemplate(functionName, handler, language)
			fmt.Printf("Building: %s with %s template. Please wait..\n", image, language)

			if shrinkwrap {
				fmt.Printf("%s shrink-wrapped to %s\n", functionName, tempPath)

				return
			}
		}

		flagSlice := buildFlagSlice(nocache, squash, os.Getenv("http_proxy"), os.Getenv("https_proxy"), buildArgMap)
		spaceSafeCmdLine := []string{"docker", "build"}
		spaceSafeCmdLine = append(spaceSafeCmdLine, flagSlice...)
		spaceSafeCmdLine = append(spaceSafeCmdLine, "-t", image, ".")
		ExecCommand(tempPath, spaceSafeCmdLine)
		fmt.Printf("Image: %s built.\n", image)

	} else {
		log.Fatalf("Language template: %s not supported. Build a custom Dockerfile instead.", language)
	}
}

// createBuildTemplate creates temporary build folder to perform a Docker build with language template
func createBuildTemplate(functionName string, handler string, language string) string {
	tempPath := fmt.Sprintf("./build/%s/", functionName)
	fmt.Printf("Clearing temporary build folder: %s\n", tempPath)

	clearErr := os.RemoveAll(tempPath)
	if clearErr != nil {
		fmt.Printf("Error clearing temporary build folder %s\n", tempPath)
	}

	fmt.Printf("Preparing %s %s\n", handler+"/", tempPath+"function")

	functionPath := tempPath + "/function"
	mkdirErr := os.MkdirAll(functionPath, 0700)
	if mkdirErr != nil {
		fmt.Printf("Error creating path %s - %s.\n", functionPath, mkdirErr.Error())
	}

	// Both Dockerfile and dockerfile are accepted
	if language == "Dockerfile" {
		language = "dockerfile"
	}
	CopyFiles("./template/"+language, tempPath)

	// Overlay in user-function
	CopyFiles(handler, functionPath)

	return tempPath
}

func buildFlagSlice(nocache bool, squash bool, httpProxy string, httpsProxy string, buildArgMap map[string]string) []string {

	var spaceSafeBuildFlags []string

	if nocache {
		spaceSafeBuildFlags = append(spaceSafeBuildFlags, "--no-cache")
	}
	if squash {
		spaceSafeBuildFlags = append(spaceSafeBuildFlags, "--squash")
	}

	if len(httpProxy) > 0 {
		spaceSafeBuildFlags = append(spaceSafeBuildFlags, "--build-arg", fmt.Sprintf("http_proxy=%s", httpProxy))
	}

	if len(httpsProxy) > 0 {
		spaceSafeBuildFlags = append(spaceSafeBuildFlags, "--build-arg", fmt.Sprintf("https_proxy=%s", httpsProxy))
	}

	for k, v := range buildArgMap {
		spaceSafeBuildFlags = append(spaceSafeBuildFlags, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}

	return spaceSafeBuildFlags
}

func ensureHandlerPath(handler string) error {
	if _, err := os.Stat(handler); err != nil {
		return err
	}

	return nil
}
