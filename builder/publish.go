// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package builder

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	v2execute "github.com/alexellis/go-execute/v2"
	"github.com/openfaas/faas-cli/schema"
	"github.com/openfaas/faas-cli/stack"

	hmac "github.com/alexellis/hmac/v2"
)

type buildConfig struct {
	Image     string            `json:"image"`
	Frontend  string            `json:"frontend,omitempty"`
	BuildArgs map[string]string `json:"buildArgs,omitempty"`
	Platforms []string          `json:"platforms,omitempty"`
}

type builderResult struct {
	Log    []string `json:"log"`
	Image  string   `json:"image"`
	Status string   `json:"status"`
}

const BuilderConfigFilename = "com.openfaas.docker.config"

// PublishImage will publish images as multi-arch
// TODO: refactor signature to a struct to simplify the length of the method header
func PublishImage(image string, handler string, functionName string, language string, nocache bool, squash bool, shrinkwrap bool, buildArgMap map[string]string,
	buildOptions []string, tagMode schema.BuildFormat, buildLabelMap map[string]string, quietBuild bool, copyExtraPaths []string, platforms string, extraTags []string, remoteBuilder, payloadSecretPath string) error {

	if stack.IsValidTemplate(language) {
		pathToTemplateYAML := fmt.Sprintf("./template/%s/template.yml", language)
		if _, err := os.Stat(pathToTemplateYAML); os.IsNotExist(err) {
			return err
		}

		langTemplate, err := stack.ParseYAMLForLanguageTemplate(pathToTemplateYAML)
		if err != nil {
			return fmt.Errorf("error reading language template: %s", err.Error())
		}

		if err := ensureHandlerPath(handler); err != nil {
			return fmt.Errorf("building %s, %s is an invalid path", functionName, handler)
		}

		// To avoid breaking the CLI for custom templates that do not set the language attribute
		// we ensure it is always set.
		//
		// While templates are expected to have the language in `template.yaml` set to the same name as the template folder
		// this was never enforced.
		langTemplate.Language = language

		if isDockerfileTemplate(langTemplate.Language) {
			langTemplate = nil
		}

		tempPath, err := CreateBuildContext(functionName, handler, langTemplate, copyExtraPaths)
		if err != nil {
			return err
		}

		if shrinkwrap {
			fmt.Printf("%s shrink-wrapped to %s\n", functionName, tempPath)
			return nil
		}

		branch, version, err := GetImageTagValues(tagMode, handler)
		if err != nil {
			return err
		}

		imageName := schema.BuildImageName(tagMode, image, version, branch)

		fmt.Printf("Building: %s with %s template. Please wait..\n", imageName, language)

		if remoteBuilder != "" {
			tempDir, err := os.MkdirTemp(os.TempDir(), "builder-*")
			if err != nil {
				return fmt.Errorf("failed to create temporary directory for %s, error: %w", functionName, err)
			}
			defer os.RemoveAll(tempDir)

			tarPath := path.Join(tempDir, "req.tar")

			builderPlatforms := strings.Split(platforms, ",")

			if err := makeTar(buildConfig{Image: imageName, BuildArgs: buildArgMap, Platforms: builderPlatforms}, path.Join("build", functionName), tarPath); err != nil {
				return fmt.Errorf("failed to create tar file for %s, error: %w", functionName, err)
			}

			res, err := callBuilder(tarPath, remoteBuilder, functionName, payloadSecretPath)
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
			buildOptPackages, buildPackageErr := getBuildOptionPackages(buildOptions, language, langTemplate.BuildOptions)

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
				Platforms:        platforms,
				ExtraTags:        extraTags,
			}

			command, args := getDockerBuildxCommand(dockerBuildVal)
			fmt.Printf("Publishing with command: %v %v\n", command, args)

			task := v2execute.ExecTask{
				Cwd:         tempPath,
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
		build.BuildOptPackages, build.BuildLabelMap)

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

func makeTar(buildConfig buildConfig, base, tarPath string) error {
	configBytes, _ := json.Marshal(buildConfig)
	if err := os.WriteFile(path.Join(base, BuilderConfigFilename), configBytes, 0664); err != nil {
		return err
	}

	tarFile, err := os.Create(tarPath)
	if err != nil {
		return err
	}

	tarWriter := tar.NewWriter(tarFile)
	defer tarWriter.Close()

	err = filepath.Walk(base, func(path string, f os.FileInfo, pathErr error) error {
		if pathErr != nil {
			return pathErr
		}

		targetFile, err := os.Open(path)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(f, f.Name())
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(path, base)
		if header.Name != fmt.Sprintf("/%s", BuilderConfigFilename) {
			header.Name = filepath.Join("context", header.Name)
		}

		header.Name = strings.TrimPrefix(header.Name, "/")

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if f.Mode().IsDir() {
			return nil
		}

		_, err = io.Copy(tarWriter, targetFile)
		return err
	})

	return err
}

func callBuilder(tarPath, builderAddress, functionName, payloadSecretPath string) (*http.Response, error) {

	payloadSecret, err := os.ReadFile(payloadSecretPath)
	if err != nil {
		return nil, err
	}

	tarFile, err := os.Open(tarPath)
	if err != nil {
		return nil, err
	}
	defer tarFile.Close()

	tarFileBytes, err := io.ReadAll(tarFile)
	if err != nil {
		return nil, err
	}

	digest := hmac.Sign(tarFileBytes, bytes.TrimSpace(payloadSecret), sha256.New)
	fmt.Println(hex.EncodeToString(digest))

	r, err := http.NewRequest(http.MethodPost, builderAddress, bytes.NewReader(tarFileBytes))
	if err != nil {
		return nil, err
	}

	r.Header.Set("X-Build-Signature", "sha256="+hex.EncodeToString(digest))
	r.Header.Set("Content-Type", "application/octet-stream")

	log.Printf("%s invoking the API for build at %s", functionName, builderAddress)
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}

	return res, nil
}
