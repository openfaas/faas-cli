// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package builder

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/openfaas/faas-cli/execute"
	"github.com/openfaas/faas-cli/schema"
	"github.com/openfaas/faas-cli/stack"
	vcs "github.com/openfaas/faas-cli/versioncontrol"
)

// AdditionalPackageBuildArg holds the special build-arg keyname for use with build-opts.
// Can also be passed as a build arg hence needs to be accessed from commands
const AdditionalPackageBuildArg = "ADDITIONAL_PACKAGE"

// BuildImage construct Docker image from function parameters
// TODO: refactor signature to a struct to simplify the length of the method header
func BuildImage(ctx context.Context, image string, handler string, functionName string, language string, nocache bool, squash bool, shrinkwrap bool, buildArgMap map[string]string, buildOptions []string, tagFormat schema.BuildFormat, buildLabelMap map[string]string, quietBuild bool, copyExtraPaths []string, remoteBuilder, payloadSecretPath string) error {

	if stack.IsValidTemplate(language) {
		pathToTemplateYAML := fmt.Sprintf("./template/%s/template.yml", language)
		if _, err := os.Stat(pathToTemplateYAML); os.IsNotExist(err) {
			return err
		}

		langTemplate, err := stack.ParseYAMLForLanguageTemplate(pathToTemplateYAML)
		if err != nil {
			return fmt.Errorf("error reading language template: %s", err.Error())
		}

		mountSSH := false
		if langTemplate.MountSSH {
			mountSSH = true
		}

		if err := ensureHandlerPath(handler); err != nil {
			return fmt.Errorf("building %s, %s is an invalid path", functionName, handler)
		}

		tempPath, err := createBuildContext(functionName, handler, language, isLanguageTemplate(language), langTemplate.HandlerFolder, copyExtraPaths)
		if err != nil {
			return err
		}

		if shrinkwrap {
			fmt.Printf("%s shrink-wrapped to %s\n", functionName, tempPath)
			return nil
		}

		branch, version, err := GetImageTagValues(tagFormat, handler)
		if err != nil {
			return err
		}

		imageName := schema.BuildImageName(tagFormat, image, version, branch)
		fmt.Printf("Building: %s with %s template. Please wait..\n", imageName, language)

		if remoteBuilder != "" {
			tempDir, err := os.MkdirTemp(os.TempDir(), "builder-*")
			if err != nil {
				return fmt.Errorf("failed to create temporary directory for %s, error: %w", functionName, err)
			}
			defer os.RemoveAll(tempDir)

			tarPath := path.Join(tempDir, "req.tar")

			if err := makeTar(builderConfig{Image: imageName, BuildArgs: buildArgMap}, path.Join("build", functionName), tarPath); err != nil {
				return fmt.Errorf("failed to create tar file for %s, error: %w", functionName, err)
			}

			res, err := callBuilder(tarPath, tempPath, remoteBuilder, functionName, payloadSecretPath)
			if err != nil {
				return err
			}
			defer res.Body.Close()

			data, _ := io.ReadAll(res.Body)

			result := builderResult{}
			if err := json.Unmarshal(data, &result); err != nil {
				return err
			}

			if !quietBuild {
				for _, logMsg := range result.Log {
					fmt.Printf("%s\n", logMsg)
				}
			}

			if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
				fmt.Println(res.StatusCode)
				return fmt.Errorf("%s failure while building or pushing image %s: %s", functionName, imageName, result.Status)
			}

			log.Printf("%s success building and pushing image: %s", functionName, result.Image)

		} else {

			buildOptPackages, err := getBuildOptionPackages(buildOptions, language, langTemplate.BuildOptions)
			if err != nil {
				return err

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

			command, args := getDockerBuildCommand(dockerBuildVal)

			envs := os.Environ()
			if mountSSH {
				envs = append(envs, "DOCKER_BUILDKIT=1")
			}

			task := execute.ExecTask{
				Cwd:         tempPath,
				Command:     command,
				Args:        args,
				StreamStdio: !quietBuild,
				Env:         envs,
			}

			res, err := task.Execute(ctx)
			if err != nil {
				return err
			}

			if res.ExitCode == -1 && errors.Is(ctx.Err(), context.Canceled) {
				return ctx.Err()
			}

			if res.ExitCode != 0 {
				return fmt.Errorf("[%s] received non-zero exit code %d from build, error: %s", functionName, res.ExitCode, res.Stderr)
			}

			fmt.Printf("Image: %s built.\n", imageName)
		}
	} else {
		return fmt.Errorf("language template: %s not supported, build a custom Dockerfile", language)
	}

	return nil
}

type FunctionMetadataSource interface {
	Get(tagType schema.BuildFormat, contextPath string) (branch, version string, err error)
}

type FunctionMetadataSourceLive struct {
}

func (FunctionMetadataSourceLive) Get(tagType schema.BuildFormat, contextPath string) (branch, version string, err error) {
	return GetImageTagValues(tagType, contextPath)
}

func NewFunctionMetadataSourceLive() FunctionMetadataSource {
	return FunctionMetadataSourceLive{}
}

// GetImageTagValues returns the image tag format and component information determined via GIT
func GetImageTagValues(tagType schema.BuildFormat, contextPath string) (branch, version string, err error) {
	switch tagType {
	case schema.SHAFormat:
		version = vcs.GetGitSHA()
		if len(version) == 0 {
			err = fmt.Errorf("cannot tag image with Git SHA as this is not a Git repository")
			return
		}
	case schema.BranchAndSHAFormat:
		branch = vcs.GetGitBranch()
		if len(branch) == 0 {
			err = fmt.Errorf("cannot tag image with Git branch and SHA as this is not a Git repository")
			return
		}

		version = vcs.GetGitSHA()
		if len(version) == 0 {
			err = fmt.Errorf("cannot tag image with Git SHA as this is not a Git repository")
			return

		}
	case schema.DescribeFormat:
		version = vcs.GetGitDescribe()
		if len(version) == 0 {
			err = fmt.Errorf("cannot tag image with Git Tag and SHA as this is not a Git repository")
			return
		}
	case schema.DigestFormat:
		hash, err := hashFolder(contextPath)
		if err != nil {
			return "", "", fmt.Errorf("unable to get hash of path: %q, %w", contextPath, err)
		}
		version = hash
	}

	return branch, version, nil
}

func hashFolder(contextPath string) (string, error) {
	m := make(map[string]string)

	if err := filepath.Walk(contextPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		hash := md5.Sum(data)
		m[path] = hex.EncodeToString(hash[:])

		return nil
	}); err != nil {
		return "", err
	}

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	hashes := ""
	for _, k := range keys {
		hashes += m[k]
	}

	hash := md5.Sum([]byte(hashes))

	return hex.EncodeToString(hash[:]), nil

}

func getDockerBuildCommand(build dockerBuild) (string, []string) {
	flagSlice := buildFlagSlice(build.NoCache, build.Squash, build.HTTPProxy, build.HTTPSProxy, build.BuildArgMap, build.BuildOptPackages, build.BuildLabelMap)
	args := []string{"build"}
	args = append(args, flagSlice...)

	args = append(args, "--tag", build.Image, ".")

	command := "docker"

	return command, args
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

	// Platforms for use with buildx and publish command
	Platforms string

	// ExtraTags for published images like :latest
	ExtraTags []string
}

var defaultDirPermissions os.FileMode = 0700

const defaultHandlerFolder string = "function"

// isRunningInCI checks the ENV var CI and returns true if it's set to true or 1
func isRunningInCI() bool {
	if env, ok := os.LookupEnv("CI"); ok {
		if env == "true" || env == "1" {
			return true
		}
	}
	return false
}

// createBuildContext creates temporary build folder to perform a Docker build with language template
func createBuildContext(functionName string, handler string, language string, useFunction bool, handlerFolder string, copyExtraPaths []string) (string, error) {
	tempPath := fmt.Sprintf("./build/%s/", functionName)

	if err := os.RemoveAll(tempPath); err != nil {
		return tempPath, fmt.Errorf("unable to clear temporary build folder: %s", tempPath)
	}

	functionPath := tempPath

	if useFunction {
		if handlerFolder == "" {
			functionPath = path.Join(functionPath, defaultHandlerFolder)
		} else {
			functionPath = path.Join(functionPath, handlerFolder)
		}
	}

	// fmt.Printf("Preparing: %s %s\n", handler+"/", functionPath)

	if isRunningInCI() {
		defaultDirPermissions = 0777
	}

	mkdirErr := os.MkdirAll(functionPath, defaultDirPermissions)
	if mkdirErr != nil {
		fmt.Printf("Error creating path: %s - %s.\n", functionPath, mkdirErr.Error())
		return tempPath, mkdirErr
	}

	if useFunction {
		if err := CopyFiles(path.Join("./template/", language), tempPath); err != nil {
			fmt.Printf("Error copying template directory: %s.\n", err.Error())
			return tempPath, err
		}
	}

	// Overlay in user-function
	// CopyFiles(handler, functionPath)
	infos, err := ioutil.ReadDir(handler)
	if err != nil {
		fmt.Printf("Error reading the handler: %s - %s.\n", handler, err.Error())
		return tempPath, err
	}

	for _, info := range infos {
		switch info.Name() {
		case "build", "template":
			fmt.Printf("Skipping \"%s\" folder\n", info.Name())
			continue
		default:
			if err := CopyFiles(
				filepath.Clean(path.Join(handler, info.Name())),
				filepath.Clean(path.Join(functionPath, info.Name())),
			); err != nil {
				return tempPath, err
			}
		}
	}

	for _, extraPath := range copyExtraPaths {
		extraPathAbs, err := pathInScope(extraPath, ".")
		if err != nil {
			return tempPath, err
		}
		// Note that if useFunction is false, ie is a `dockerfile` template, then
		// functionPath == tempPath, the docker build context, not the `function` handler folder
		// inside the docker build context
		copyErr := CopyFiles(
			extraPathAbs,
			filepath.Clean(path.Join(functionPath, extraPath)),
		)

		if copyErr != nil {
			return tempPath, copyErr
		}
	}

	return tempPath, nil
}

// pathInScope returns the absolute path to `path` and ensures that it is located within the
// provided scope. An error will be returned, if the path is outside of the provided scope.
func pathInScope(path string, scope string) (string, error) {
	scope, err := filepath.Abs(filepath.FromSlash(scope))
	if err != nil {
		return "", err
	}

	abs, err := filepath.Abs(filepath.FromSlash(path))
	if err != nil {
		return "", err
	}

	if abs == scope {
		return "", fmt.Errorf("forbidden path appears to equal the entire project: %s (%s)", path, abs)
	}

	if strings.HasPrefix(abs, scope) {
		return abs, nil
	}

	// default return is an error
	return "", fmt.Errorf("forbidden path appears to be outside of the build context: %s (%s)", path, abs)
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

func getBuildOptionPackages(requestedBuildOptions []string, language string, availableBuildOptions []stack.BuildOption) ([]string, error) {

	var buildPackages []string

	if len(requestedBuildOptions) > 0 {

		var allFound bool

		buildPackages, allFound = getPackages(availableBuildOptions, requestedBuildOptions)

		if !allFound {

			return nil, fmt.Errorf(
				`Error: You're using a build option unavailable for %s.
Please check /template/%s/template.yml for supported build options`, language, language)
		}

	}
	return buildPackages, nil
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

func isLanguageTemplate(language string) bool {
	return strings.ToLower(language) != "dockerfile"
}
