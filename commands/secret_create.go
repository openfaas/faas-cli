// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/spf13/cobra"
)

var (
	secretName  string
	secretValue string
	secretFile  string
)

// secretCreateCmd represents the secretCreate command
var secretCreateCmd = &cobra.Command{
	Use:   "create SECRET_NAME --value=SECRET_VALUE --from-file=/path/to/secret/file",
	Short: "Create a new secret",
	Long:  `The create command creates a new secret`,
	Example: `faas-cli secret create secret-name --value=secret-value
faas-cli secret create secret-name --value=secret-value --gateway=http://127.0.0.1:8080
faas-cli secret create secret-name --from-file=/path/to/secret/file --gateway=http://127.0.0.1:8080`,
	RunE:    runSecretCreate,
	PreRunE: preRunSecretCreate,
}

func init() {
	secretCreateCmd.Flags().StringVar(&secretValue, "value", "", "Value of the secret")
	secretCreateCmd.Flags().StringVar(&secretFile, "from-file", "", "Path to the secret file")
	secretCreateCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")
	secretCreateCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	secretCmd.AddCommand(secretCreateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// secretCreateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// secretCreateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func preRunSecretCreate(cmd *cobra.Command, args []string) error {

	if len(secretValue) == 0 && len(secretFile) == 0 {
		return fmt.Errorf("please provide a value or file path for the secret")
	}

	if len(secretValue) != 0 && len(secretFile) != 0 {
		return fmt.Errorf("please either provide value or file for the secret")
	}

	return nil
}

func runSecretCreate(cmd *cobra.Command, args []string) error {
	var gatewayAddress string

	if len(args) < 1 {
		return fmt.Errorf("please provide secret name")
	}
	secretName = args[0]

	if len(secretFile) > 0 {
		fileSecretValue, err := readSecretFromFile(secretFile)
		if err != nil {
			return err
		}
		secretValue = fileSecretValue
	}

	gatewayAddress = getGatewayURL(gateway, defaultGateway, "", os.Getenv(openFaaSURLEnvironment))

	fmt.Println("Creating secret: " + secretName + "\n")
	_, output := proxy.CreateSecret(gatewayAddress, secretName, secretValue, tlsInsecure)
	fmt.Printf(output)

	return nil
}

func readSecretFromFile(secretFile string) (string, error) {
	fileData, err := ioutil.ReadFile(secretFile)
	return string(fileData), err
}
