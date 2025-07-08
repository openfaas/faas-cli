// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package builder

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	v2execute "github.com/alexellis/go-execute/v2"
	"github.com/openfaas/faas-cli/schema"
	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/go-sdk/builder"
)

// PublishImage will publish images as multi-arch
// TODO: refactor signature to a struct to simplify the length of the method header
func PublishImage(image string, handler string, functionName string, language string, nocache bool, squash bool, shrinkwrap bool, buildArgMap map[string]string,
	buildOptions []string, tagMode schema.BuildFormat, buildLabelMap map[string]string, quietBuild bool, copyExtraPaths []string, platforms string, extraTags []string, remoteBuilder, payloadSecretPath string, forcePull bool) error {

	if stack.IsValidTemplate(language) {
		pathToTemplateYAML := fmt.Sprintf("./template/%s/template.yml", language)
		if _, err := os.Stat(pathToTemplateYAML); err != nil && os.IsNotExist(err) {
			return err
		}

		langTemplate, err := stack.ParseYAMLForLanguageTemplate(pathToTemplateYAML)
		if err != nil {
			return fmt.Errorf("error reading language template: %s", err.Error())
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

		branch, version, err := GetImageTagValues(tagMode, handler)
		if err != nil {
			return err
		}

		imageName := schema.BuildImageName(tagMode, image, version, branch)

		buildOptPackages, err := getBuildOptionPackages(buildOptions, language, langTemplate.BuildOptions)
		if err != nil {
			return err
		}
		buildArgMap = appendAdditionalPackages(buildArgMap, buildOptPackages)

		fmt.Printf("Building: %s with %s template. Please wait..\n", imageName, language)

		if remoteBuilder != "" {

			if forcePull {
				return fmt.Errorf("--pull is not supported with --remote-builder")
			}

			tempDir, err := os.MkdirTemp(os.TempDir(), "openfaas-build-*")
			if err != nil {
				return fmt.Errorf("failed to create temporary directory: %w", err)
			}
			defer os.RemoveAll(tempDir)

			tarPath := path.Join(tempDir, "req.tar")

			builderPlatforms := strings.Split(platforms, ",")
			buildConfig := builder.BuildConfig{
				Image:     imageName,
				BuildArgs: buildArgMap,
				Platforms: builderPlatforms,
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
				return fmt.Errorf("failed to invoke builder:: %w", err)
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
				Platforms:     platforms,
				ExtraTags:     extraTags,
				ForcePull:     forcePull,
			}

			command, args := getDockerBuildxCommand(dockerBuildVal)
			fmt.Printf("Publishing with command: %v %v\n", command, args)

			task := v2execute.ExecTask{
				Cwd:         buildContext,
				Command:     command,
				Args:        args,
				StreamStdio: !quietBuild,
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

func getDockerBuildxCommand(build dockerBuild) (string, []string) {
	flagSlice := buildFlagSlice(build.NoCache, build.Squash, build.HTTPProxy, build.HTTPSProxy, build.BuildArgMap,
		build.BuildLabelMap, build.ForcePull)

	// pushOnly defined at https://github.com/docker/buildx
	const pushOnly = "--output=type=registry,push=true"

	args := []string{"buildx", "build", "--progress=plain", "--platform=" + build.Platforms, pushOnly}

	args = append(args, flagSlice...)

	args = append(args, "--tag", build.Image, ".")

	for _, t := range build.ExtraTags {

		var tag string
		if i := strings.LastIndex(build.Image, ":"); i > -1 {
			tag = applyTag(i, build.Image, t)
		} else {
			tag = applyTag(len(build.Image)-1, build.Image, t)
		}
		args = append(args, "--tag", tag)
	}

	command := "docker"

	return command, args
}

func applyTag(index int, baseImage, tag string) string {
	return fmt.Sprintf("%s:%s", baseImage[:index], tag)
}
