// Copyright (c) OpenFaaS Author(s) 2020. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var registryLoginCommand = &cobra.Command{
	Use:          "registry-login",
	Short:        "Generate and save the registry authentication file",
	SilenceUsage: true,
	RunE:         generateRegistryAuthFile,
	PreRunE:      generateRegistryPreRun,
}

func init() {
	registryLoginCommand.Flags().String("server", "https://index.docker.io/v1/", "The server URL, it is defaulted to the docker registry")
	registryLoginCommand.Flags().StringP("username", "u", "", "The Registry Username")
	registryLoginCommand.Flags().String("password", "", "The registry password")
	registryLoginCommand.Flags().BoolP("password-stdin", "s", false, "Reads the docker password from stdin, either pipe to the command or remember to press ctrl+d when reading interactively")

	registryLoginCommand.Flags().Bool("ecr", false, "If we are using ECR we need a different set of flags, so if this is set, we need to set --username and --password")
	registryLoginCommand.Flags().String("account-id", "", "Your AWS Account id")
	registryLoginCommand.Flags().String("region", "", "Your AWS region")

	faasCmd.AddCommand(registryLoginCommand)
}

func generateRegistryPreRun(command *cobra.Command, args []string) error {
	server, err := command.Flags().GetString("server")
	if err != nil {
		return fmt.Errorf("error with --server usage: %s", err)
	}

	username, err = command.Flags().GetString("username")
	if err != nil {
		return fmt.Errorf("error with --username usage: %s", err)
	}

	password, err = command.Flags().GetString("password")
	if err != nil {
		return fmt.Errorf("error with --password usage: %s", err)
	}

	_, err = command.Flags().GetBool("password-stdin")
	if err != nil {
		return fmt.Errorf("error with --password-stdin usage: %s", err)
	}

	ecr, err := command.Flags().GetBool("ecr")
	if err != nil {
		return fmt.Errorf("error with --ecr usage: %s", err)
	}

	accountID, err := command.Flags().GetString("account-id")
	if err != nil {
		return fmt.Errorf("error with --account-id usage: %s", err)
	}

	region, err := command.Flags().GetString("region")
	if err != nil {
		return fmt.Errorf("error with --region usage: %s", err)
	}

	if ecr {
		if len(server) > 0 {
			return fmt.Errorf("the --server flag is not supported with ECR")
		}

		if len(password) > 0 {
			return fmt.Errorf("the --password flag is not supported with ECR")
		}

		if len(username) > 0 {
			return fmt.Errorf("the --username flag is not supported with ECR")
		}
		if len(accountID) == 0 {
			return fmt.Errorf("the --account-id flag is required with ECR")
		}
		if len(region) == 0 {
			return fmt.Errorf("the --region flag is required with ECR")
		}
	}

	return nil
}

func generateRegistryAuthFile(command *cobra.Command, _ []string) error {
	ecrEnabled, _ := command.Flags().GetBool("ecr")
	accountID, _ := command.Flags().GetString("account-id")
	region, _ := command.Flags().GetString("region")
	username, _ := command.Flags().GetString("username")
	password, _ := command.Flags().GetString("password")
	server, _ := command.Flags().GetString("server")
	passStdin, _ := command.Flags().GetBool("password-stdin")

	if len(username) == 0 {
		return fmt.Errorf("you must give --username (-u)")
	}

	if ecrEnabled {
		if err := generateECRFile(accountID, region); err != nil {
			return err
		}

	} else if passStdin {
		fmt.Printf("Enter your password, hit enter then type Ctrl+D\n\nPassword: ")
		passwordStdin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		if err := generateFile(username, strings.TrimSpace(string(passwordStdin)), server); err != nil {
			return err
		}
	} else {
		if err := generateFile(username, password, server); err != nil {
			return err
		}
	}

	fmt.Printf("\nWrote ./credentials/config.json..OK\n")

	return nil
}

func generateFile(username string, password string, server string) error {

	fileBytes, err := generateRegistryAuth(server, username, password)
	if err != nil {
		return err
	}
	return writeFileToFassCLITmp(fileBytes)
}

func generateECRFile(accountID string, region string) error {

	fileBytes, err := generateECRRegistryAuth(accountID, region)
	if err != nil {
		return err
	}

	return writeFileToFassCLITmp(fileBytes)
}

func generateRegistryAuth(server, username, password string) ([]byte, error) {
	if len(username) == 0 || len(password) == 0 || len(server) == 0 {
		return nil, errors.New("both --username and (--password-stdin or --password) are required")
	}

	encodedString := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	data := RegistryAuth{
		AuthConfigs: map[string]Auth{
			server: {Base64AuthString: encodedString},
		},
	}

	registryBytes, err := json.MarshalIndent(data, "", " ")

	return registryBytes, err
}

func generateECRRegistryAuth(accountID, region string) ([]byte, error) {
	if len(accountID) == 0 || len(region) == 0 {
		return nil, errors.New("you must provide an --account-id and --region when using --ecr")
	}

	data := ECRRegistryAuth{
		CredsStore: "ecr-login",
		CredHelpers: map[string]string{
			fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", accountID, region): "ecr-login",
		},
	}

	registryBytes, err := json.MarshalIndent(data, "", " ")

	return registryBytes, err
}

func writeFileToFassCLITmp(fileBytes []byte) error {
	path := "./credentials"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0744)
		if err != nil {
			return err
		}
	}

	return ioutil.WriteFile(filepath.Join(path, "config.json"), fileBytes, 0744)
}

type Auth struct {
	Base64AuthString string `json:"auth"`
}

type RegistryAuth struct {
	AuthConfigs map[string]Auth `json:"auths"`
}

type ECRRegistryAuth struct {
	CredsStore  string            `json:"credsStore"`
	CredHelpers map[string]string `json:"credHelpers"`
}
