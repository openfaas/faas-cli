// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package builder

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	v2execute "github.com/alexellis/go-execute/v2"
	"github.com/openfaas/faas-cli/schema"
	"github.com/openfaas/faas-cli/stack"
	vcs "github.com/openfaas/faas-cli/versioncontrol"
	"github.com/openfaas/go-sdk/builder"
)

// AdditionalPackageBuildArg holds the special build-arg keyname for use with build-opts.
// Can also be passed as a build arg hence needs to be accessed from commands
const AdditionalPackageBuildArg = "ADDITIONAL_PACKAGE"

// BuildImage construct Docker image from function parameters
// TODO: refactor signature to a struct to simplify the length of the method header
func BuildImage(image string, handler string, functionName string, language string, nocache bool, squash bool, shrinkwrap bool, buildArgMap map[string]string, buildOptions []string, tagFormat schema.BuildFormat, buildLabelMap map[string]string, quietBuild bool, copyExtraPaths []string, remoteBuilder, payloadSecretPath string, forcePull bool) error {

	if stack.IsValidTemplate(language) {
		pathToTemplateYAML := fmt.Sprintf("./template/%s/template.yml", language)
		if _, err := os.Stat(pathToTemplateYAML); err != nil && os.IsNotExist(err) {
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

		opts := []builder.BuildContextOption{}
		if len(langTemplate.HandlerFolder) > 0 {
			opts = append(opts, builder.WithHandlerOverlay(langTemplate.HandlerFolder))
		}

		buildContext, err := builder.CreateBuildContext(functionName, handler, language, copyExtraPaths, opts...)
		if err != nil {
			return err
		}

		if shrinkwrap {
			fmt.Printf("%s shrink-wrapped to %s\n", functionName, buildContext)
			return nil
		}

		branch, version, err := GetImageTagValues(tagFormat, handler)
		if err != nil {
			return err
		}

		imageName := schema.BuildImageName(tagFormat, image, version, branch)

		buildOptPackages, err := getBuildOptionPackages(buildOptions, language, langTemplate.BuildOptions)
		if err != nil {
			return err

		}
		buildArgMap = appendAdditionalPackages(buildArgMap, buildOptPackages)

		fmt.Printf("Building: %s with %s template. Please wait..\n", imageName, language)

		if remoteBuilder != "" {
			tempDir, err := os.MkdirTemp(os.TempDir(), "openfaas-build-*")
			if err != nil {
				return fmt.Errorf("failed to create temporary directory: %w", err)
			}
			defer os.RemoveAll(tempDir)

			tarPath := path.Join(tempDir, "req.tar")

			buildConfig := builder.BuildConfig{
				Image:     imageName,
				BuildArgs: buildArgMap,
			}

			// Prepare a tar archive that contains the build config and build context.
			if err := builder.MakeTar(tarPath, path.Join("build", functionName), &buildConfig); err != nil {
				return fmt.Errorf("failed to create tar file for %s, error: %w", functionName, err)
			}

			// Get the HMAC secret used for payload authentication with the builder API.
			payloadSecret, err := os.ReadFile(payloadSecretPath)
			if err != nil {
				return fmt.Errorf("failed to read payload secret: %w", err)
			}
			payloadSecret = bytes.TrimSpace(payloadSecret)

			// Initialize a new builder client.
			u, _ := url.Parse(remoteBuilder)
			builderURL := &url.URL{
				Scheme: u.Scheme,
				Host:   u.Host,
			}
			b := builder.NewFunctionBuilder(builderURL, http.DefaultClient, builder.WithHmacAuth(string(payloadSecret)))

			stream, err := b.BuildWithStream(tarPath)
			if err != nil {
				return fmt.Errorf("failed to invoke builder: %w", err)
			}
			defer stream.Close()

			for result := range stream.Results() {
				if !quietBuild {
					for _, logMsg := range result.Log {
						fmt.Printf("%s\n", logMsg)
					}
				}

				switch result.Status {
				case builder.BuildSuccess:
					log.Printf("%s success building and pushing image: %s", functionName, result.Image)
				case builder.BuildFailed:
					return fmt.Errorf("%s failure while building or pushing image %s: %s", functionName, imageName, result.Error)
				}
			}

		} else {
			dockerBuildVal := dockerBuild{
				Image:         imageName,
				NoCache:       nocache,
				Squash:        squash,
				HTTPProxy:     os.Getenv("http_proxy"),
				HTTPSProxy:    os.Getenv("https_proxy"),
				BuildArgMap:   buildArgMap,
				BuildLabelMap: buildLabelMap,
				ForcePull:     forcePull,
			}

			command, args := getDockerBuildCommand(dockerBuildVal)

			envs := os.Environ()
			if mountSSH {
				envs = append(envs, "DOCKER_BUILDKIT=1")
			}
			log.Printf("Build flags: %+v\n", args)

			task := v2execute.ExecTask{
				Cwd:         buildContext,
				Command:     command,
				Args:        args,
				StreamStdio: !quietBuild,
				Env:         envs,
			}

			res, err := task.Execute(context.TODO())

			if err != nil {
				return err
			}

			if res.ExitCode != 0 {
				return fmt.Errorf("[%s] received non-zero exit code from build, error: %s", functionName, res.Stderr)
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

		data, err := os.ReadFile(path)
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
	flagSlice := buildFlagSlice(build.NoCache, build.Squash, build.HTTPProxy, build.HTTPSProxy, build.BuildArgMap, build.BuildLabelMap, build.ForcePull)
	args := []string{"build"}
	args = append(args, flagSlice...)

	args = append(args, "--tag", build.Image, ".")

	command := "docker"

	return command, args
}

type dockerBuild struct {
	Image         string
	Version       string
	NoCache       bool
	Squash        bool
	HTTPProxy     string
	HTTPSProxy    string
	BuildArgMap   map[string]string
	BuildLabelMap map[string]string

	// Platforms for use with buildx and publish command
	Platforms string

	// ExtraTags for published images like :latest
	ExtraTags []string

	ForcePull bool
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

func buildFlagSlice(nocache bool, squash bool, httpProxy string, httpsProxy string, buildArgMap map[string]string, buildLabelMap map[string]string, forcePull bool) []string {

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

	for k, v := range buildLabelMap {
		spaceSafeBuildFlags = append(spaceSafeBuildFlags, "--label", fmt.Sprintf("%s=%s", k, v))
	}

	if forcePull {
		spaceSafeBuildFlags = append(spaceSafeBuildFlags, "--pull")
	}

	return spaceSafeBuildFlags
}

// appendAdditionalPackages appends additional packages  to the ADDITIONAL_PACKAGE build arg.
// If the ADDITIONAL_PACKAGE build arg is not present, it is created.
// If the ADDITIONAL_PACKAGE build arg is present, the packages are appended to the list.
func appendAdditionalPackages(buildArgs map[string]string, additionalPackages []string) map[string]string {
	for k, v := range buildArgs {
		if k == AdditionalPackageBuildArg {
			packages := strings.Split(v, " ")
			for i := range packages {
				packages[i] = strings.TrimSpace(packages[i])
			}

			// Remove empty strings
			packages = slices.DeleteFunc(packages, func(s string) bool {
				return len(s) == 0
			})

			additionalPackages = append(additionalPackages, packages...)
		}
	}
	additionalPackages = deDuplicate(additionalPackages)
	sort.Strings(additionalPackages)
	buildArgs[AdditionalPackageBuildArg] = strings.Join(additionalPackages, " ")

	return buildArgs
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
