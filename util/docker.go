// Copyright (c) OpenFaaS Project 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker-credential-helpers/client"
	"github.com/docker/docker/api/types"
	"github.com/mitchellh/go-homedir"
	"github.com/morikuni/aec"
	"github.com/pkg/errors"
)

const (
	dockerRegistryHost = "https://index.docker.io/v1/"
)

type DockerConfig struct {
	initialized      bool
	CredentialsStore string                      `json:"credsStore"`
	AuthConfig       map[string]types.AuthConfig `json:"auths"`
}

type DockerErrorDetailResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type DockerProgressDetailResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type DockerImageBuildResponse struct {
	Id             string                       `json:"id"`
	Status         string                       `json:"status"`
	Stream         string                       `json:"stream"`
	Progress       string                       `json:"progress"`
	ProgressDetail DockerProgressDetailResponse `json:"progressDetail"`
	Error          string                       `json:"error"`
	ErrorDetail    DockerErrorDetailResponse    `json:"errorDetail"`
}

var config DockerConfig

func init() {
	config = DockerConfig{}
}

func DockerGetConfig() *DockerConfig {
	return &config
}

func DockerGetServerHostFromImage(image string) string {
	if imageParts := strings.Split(image, "/"); len(imageParts) == 2 {
		// docker registry: <owner>/<name>
		return dockerRegistryHost
	} else {
		// other registries: <host>/<owner>/<name>
		// Get the registry host which is the first part of the image name
		return imageParts[0]
	}
}

func DockerGetCredentialsForImage(image string) (types.AuthConfig, error) {
	return DockerGetCredentials(DockerGetServerHostFromImage(image))
}

func DockerGetCredentials(server string) (types.AuthConfig, error) {
	return dockerGetCredentials(server, "")
}

// dockerGetCredentials retrieves the credential for a server
// configDir defines the directory in which the Docker's config.json is to be used, default to HOME/.docker
func dockerGetCredentials(server string, configDir string) (types.AuthConfig, error) {
	if !config.initialized {
		if "" == configDir {
			home, err := homedir.Dir()
			if err != nil {
				return types.AuthConfig{}, err
			}
			configDir = filepath.Join(home, ".docker")
		}

		file, err := os.Open(filepath.Join(configDir, "config.json"))
		if err != nil {
			return types.AuthConfig{}, err
		}
		defer file.Close()

		content, err := ioutil.ReadAll(file)
		if err != nil {
			return types.AuthConfig{}, err
		}
		if err = json.Unmarshal(content, &config); err != nil {
			return types.AuthConfig{}, err
		}
		config.initialized = true
	}

	// Get from config first
	if authConfig, exists := config.AuthConfig[server]; exists && len(authConfig.Auth) > 0 {
		DebugPrint("Retrieving credential for %s from config\n", server)

		return authConfig, nil
	}

	// If there is a credential store defined, try to retrieve from the store
	if config.CredentialsStore != "" {
		DebugPrint("Retrieving credential for %s from store %s\n", server, config.CredentialsStore)

		p := client.NewShellProgramFunc("docker-credential-" + config.CredentialsStore)

		if credentials, err := client.Get(p, server); err != nil {
			return types.AuthConfig{}, err
		} else {
			config.AuthConfig[server] = types.AuthConfig{
				Username:      credentials.Username,
				Password:      credentials.Secret,
				ServerAddress: credentials.ServerURL,
				Auth:          base64.StdEncoding.EncodeToString([]byte(credentials.Username + ":" + credentials.Secret)),
			}
		}
	}
	return config.AuthConfig[server], nil
}

func DockerImageBuildProcess(r types.ImageBuildResponse) error {
	fmt.Printf("OS Type : %s\n", r.OSType)

	d := json.NewDecoder(r.Body)
	for {
		var msg DockerImageBuildResponse
		if err := d.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(aec.RedF.Apply(fmt.Sprintf("Error while decoding Docker response: %s", err)))
		}

		if len(msg.ErrorDetail.Message) > 0 {
			return errors.New(msg.ErrorDetail.Message)
		} else if len(msg.Stream) > 0 {
			DebugPrint(msg.Stream)
		}
	}

	return nil
}

func DockerPrintResponse(r io.Reader) {
	d := json.NewDecoder(r)
	for {
		var msg struct {
			Error          string                       `json:"error"`
			Status         string                       `json:"status"`
			Progress       string                       `json:"progress"`
			ProgressDetail DockerProgressDetailResponse `json:"progressDetail"`
			Id             string                       `json:"id"`
			Aux            struct {
				Tag    string `json:"tag"`
				Digest string `json:"digest"`
				Size   int    `json:"size"`
			} `json:"aux"`
		}
		if err := d.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(aec.RedF.Apply(fmt.Sprintf("Error while decoding Docker response: %s", err)))
		}

		if len(msg.Status) > 0 {
			DebugPrint("%s: %s\n", msg.Id, msg.Status)
		}
	}
}
