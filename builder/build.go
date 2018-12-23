// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package builder

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/openfaas/faas-cli/schema"
	"github.com/openfaas/faas-cli/stack"
)

// AdditionalPackageBuildArg holds the special build-arg keyname for use with build-opts.
// Can also be passed as a build arg hence needs to be accessed from commands
const AdditionalPackageBuildArg = "ADDITIONAL_PACKAGE"

// BuildImage construct Docker image from function parameters
func BuildImage(image string, handler string, functionName string, language string, nocache bool, squash bool, shrinkwrap bool, buildArgMap map[string]string, buildOptions []string, tag string, buildLabelMap map[string]string) error {

	if stack.IsValidTemplate(language) {

		format := schema.DefaultFormat

		var version string

		if strings.ToLower(tag) == "sha" {
			version = GetGitSHA()
			if len(version) == 0 {
				return fmt.Errorf("cannot tag image with Git SHA as this is not a Git repository")

			}
			format = schema.SHAFormat
		}

		var branch string

		if strings.ToLower(tag) == "branch" {
			branch = GetGitBranch()
			if len(branch) == 0 {
				return fmt.Errorf("cannot tag image with Git branch and SHA as this is not a Git repository")

			}
			version = GetGitSHA()
			if len(version) == 0 {
				return fmt.Errorf("cannot tag image with Git SHA as this is not a Git repository")

			}
			format = schema.BranchAndSHAFormat
		}

		imageName := schema.BuildImageName(format, image, version, branch)

		var tempPath string

		if strings.ToLower(language) == "dockerfile" {

			tempPath = handler
			if shrinkwrap {
				tempPath = dockerBuildFolder(functionName, handler, language)
				fmt.Printf("%s shrink-wrapped to %s\n", functionName, tempPath)
				return nil
			}

			if err := ensureHandlerPath(handler); err != nil {

				return fmt.Errorf("building %s, %s is an invalid path", imageName, handler)
			}

			fmt.Printf("Building: %s with Dockerfile. Please wait..\n", imageName)

		} else {

			if err := ensureHandlerPath(handler); err != nil {
				return fmt.Errorf("building %s, %s is an invalid path", imageName, handler)
			}

			tempPath = createBuildTemplate(functionName, handler, language)
			fmt.Printf("Building: %s with %s template. Please wait..\n", imageName, language)

			if shrinkwrap {
				fmt.Printf("%s shrink-wrapped to %s\n", functionName, tempPath)

				return nil
			}
		}

		buildOptPackages, buildPackageErr := getBuildOptionPackages(buildOptions, language)

		if buildPackageErr != nil {
			return buildPackageErr

		}

		dockerBuildVal := dockerBuild{
			Image:            imageName,
			NoCache:          nocache,
			Squash:           squash,
			HTTPProxy:        os.Getenv("http_proxy"),
			HTTPSProxy:       os.Getenv("https_proxy"),
			BuildArgMap:      buildArgMap,
			BuildOptPackages: buildOptPackages,
			BuildLabelMap:    buildLabelMap,
		}

		spaceSafeCmdLine := getDockerBuildCommand(dockerBuildVal)

		ExecCommand(tempPath, spaceSafeCmdLine)
		fmt.Printf("Image: %s built.\n", imageName)

	} else {
		return fmt.Errorf("language template: %s not supported, build a custom Dockerfile", language)
	}

	return nil
}

func getDockerBuildCommand(build dockerBuild) []string {
	flagSlice := buildFlagSlice(build.NoCache, build.Squash, build.HTTPProxy, build.HTTPSProxy, build.BuildArgMap, build.BuildOptPackages, build.BuildLabelMap)
	command := []string{"docker", "build"}
	command = append(command, flagSlice...)
	command = append(command, "-t", build.Image, ".")

	return command
}

type dockerBuild struct {
	Image            string
	Version          string
	NoCache          bool
	Squash           bool
	HTTPProxy        string
	HTTPSProxy       string
	BuildArgMap      map[string]string
	BuildOptPackages []string
	BuildLabelMap    map[string]string
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

	CopyFiles("./template/"+language, tempPath)

	// Overlay in user-function
	// CopyFiles(handler, functionPath)
	infos, readErr := ioutil.ReadDir(handler)
	if readErr != nil {
		fmt.Printf("Error reading the handler %s - %s.\n", handler, readErr.Error())
	}

	for _, info := range infos {
		switch info.Name() {
		case "build", "template":
			fmt.Printf("Skipping \"%s\" folder\n", info.Name())
			continue
		default:
			copyErr := CopyFiles(
				filepath.Clean(handler+"/"+info.Name()),
				filepath.Clean(functionPath+"/"+info.Name()),
			)

			if copyErr != nil {
				log.Fatal(copyErr)
			}
		}
	}

	return tempPath
}

func dockerBuildFolder(functionName string, handler string, language string) string {
	tempPath := fmt.Sprintf("./build/%s/", functionName)
	fmt.Printf("Clearing temporary build folder: %s\n", tempPath)

	clearErr := os.RemoveAll(tempPath)
	if clearErr != nil {
		fmt.Printf("Error clearing temporary build folder %s\n", tempPath)
	}

	fmt.Printf("Preparing %s %s\n", handler+"/", tempPath)

	// Both Dockerfile and dockerfile are accepted
	if language == "Dockerfile" {
		language = "dockerfile"
	}

	// CopyFiles(handler, tempPath)
	infos, readErr := ioutil.ReadDir(handler)
	if readErr != nil {
		fmt.Printf("Error reading the handler %s - %s.\n", handler, readErr.Error())
	}

	for _, info := range infos {
		switch info.Name() {
		case "build", "template":
			fmt.Printf("Skipping \"%s\" folder\n", info.Name())
			continue
		default:
			copyErr := CopyFiles(
				filepath.Clean(handler+"/"+info.Name()),
				filepath.Clean(tempPath+"/"+info.Name()),
			)

			if copyErr != nil {
				log.Fatal(copyErr)
			}
		}
	}

	return tempPath
}

func buildFlagSlice(nocache bool, squash bool, httpProxy string, httpsProxy string, buildArgMap map[string]string, buildOptionPackages []string, buildLabelMap map[string]string) []string {

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

		if k != AdditionalPackageBuildArg {
			spaceSafeBuildFlags = append(spaceSafeBuildFlags, "--build-arg", fmt.Sprintf("%s=%s", k, v))
		} else {
			buildOptionPackages = append(buildOptionPackages, strings.Split(v, " ")...)
		}
	}
	if len(buildOptionPackages) > 0 {
		buildOptionPackages = deDuplicate(buildOptionPackages)
		spaceSafeBuildFlags = append(spaceSafeBuildFlags, "--build-arg", fmt.Sprintf("%s=%s", AdditionalPackageBuildArg, strings.Join(buildOptionPackages, " ")))
	}

	for k, v := range buildLabelMap {
		spaceSafeBuildFlags = append(spaceSafeBuildFlags, "--label", fmt.Sprintf("%s=%s", k, v))
	}

	return spaceSafeBuildFlags
}

func ensureHandlerPath(handler string) error {
	if _, err := os.Stat(handler); err != nil {
		return err
	}

	return nil
}

func getBuildOptionPackages(requestedBuildOptions []string, language string) ([]string, error) {

	var buildPackages []string

	if len(requestedBuildOptions) > 0 {

		var allFound bool

		availableBuildOptions, err := getBuildOptionsFor(language)

		if err != nil {
			return nil, err
		}

		buildPackages, allFound = getPackages(availableBuildOptions, requestedBuildOptions)

		if !allFound {
			err = fmt.Errorf("Error: You're using a build option unavailable for %s. Please check /template/%s/template.yml for supported build options", language, language)
			return nil, err
		}

	}
	return buildPackages, nil
}

func getBuildOptionsFor(language string) ([]stack.BuildOption, error) {

	var buildOptions = []stack.BuildOption{}

	pathToTemplateYAML := "./template/" + language + "/template.yml"
	if _, err := os.Stat(pathToTemplateYAML); os.IsNotExist(err) {
		return buildOptions, err
	}

	var langTemplate stack.LanguageTemplate
	parsedLangTemplate, err := stack.ParseYAMLForLanguageTemplate(pathToTemplateYAML)

	if err != nil {
		return buildOptions, err
	}

	if parsedLangTemplate != nil {
		langTemplate = *parsedLangTemplate
		buildOptions = langTemplate.BuildOptions
	}

	return buildOptions, nil
}

func getPackages(availableBuildOptions []stack.BuildOption, requestedBuildOptions []string) ([]string, bool) {
	var buildPackages []string

	for _, requestedOption := range requestedBuildOptions {

		requestedOptionAvailable := false

		for _, availableOption := range availableBuildOptions {

			if availableOption.Name == requestedOption {
				buildPackages = append(buildPackages, availableOption.Packages...)
				requestedOptionAvailable = true
				break
			}
		}
		if requestedOptionAvailable == false {
			return buildPackages, false
		}
	}

	return deDuplicate(buildPackages), true
}

func deDuplicate(buildOptPackages []string) []string {

	seenPackages := map[string]bool{}
	retPackages := []string{}

	for _, packageName := range buildOptPackages {

		if _, alreadySeen := seenPackages[packageName]; !alreadySeen {

			seenPackages[packageName] = true
			retPackages = append(retPackages, packageName)
		}
	}
	return retPackages
}
