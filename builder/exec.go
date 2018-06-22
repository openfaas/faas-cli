// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package builder

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/morikuni/aec"
	"github.com/openfaas/faas-cli/util"
)

var dockerClient *client.Client

func init() {
	// Init Docker client
	if c, err := client.NewEnvClient(); err != nil {
		log.Fatalf(aec.RedF.Apply(err.Error()))
	} else {
		dockerClient = c
	}
}

// ExecCommand run a system command
func ExecCommand(tempPath string, builder []string) {
	targetCmd := exec.Command(builder[0], builder[1:]...)
	targetCmd.Dir = tempPath
	targetCmd.Stdout = os.Stdout
	targetCmd.Stderr = os.Stderr
	targetCmd.Start()
	err := targetCmd.Wait()
	if err != nil {
		errString := fmt.Sprintf("ERROR - Could not execute command: %s", builder)
		log.Fatalf(aec.RedF.Apply(errString))
	}
}

func Build(tempPath string, image string, nocache bool, squash bool, buildArgs map[string]*string) error {
	buf := new(bytes.Buffer)
	util.CreateTar(tempPath, buf)
	buildContext := bytes.NewReader(buf.Bytes())

	options := types.ImageBuildOptions{
		Tags:      []string{image},
		BuildArgs: buildArgs,
		Squash:    squash,
		NoCache:   nocache,
	}
	if response, err := dockerClient.ImageBuild(context.Background(), buildContext, options); err != nil {
		return err
	} else {
		if err := util.DockerImageBuildProcess(response); err != nil {
			return err
		}
	}

	return nil
}

func Push(image string) error {
	credential, err := util.DockerGetCredentialsForImage(image)
	if err != nil {
		return err
	}

	options := types.ImagePushOptions{
		RegistryAuth: credential.Auth,
	}
	if response, err := dockerClient.ImagePush(context.Background(), image, options); err != nil {
		return err
	} else {
		util.DockerPrintResponse(response)
	}

	return nil
}
