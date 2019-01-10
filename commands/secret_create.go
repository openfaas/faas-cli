// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

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
	Use:   "create SECRET_NAME [--from-literal=SECRET_VALUE] [--from-file=/path/to/secret/file] [STDIN]",
	Short: "Create a new secret",
	Long:  `The create command creates a new secret from file, literal or STDIN`,
	Example: `faas-cli secret create secret-name --from-literal=secret-value
faas-cli secret create secret-name --from-literal=secret-value --gateway=http://127.0.0.1:8080
faas-cli secret create secret-name --from-file=/path/to/secret/file --gateway=http://127.0.0.1:8080
cat /path/to/secret/file | faas-cli secret create secret-name`,
	RunE:    runSecretCreate,
	PreRunE: preRunSecretCreate,
}

func init() {
	secretCreateCmd.Flags().StringVar(&secretValue, "from-literal", "", "Value of the secret")
	secretCreateCmd.Flags().StringVar(&secretFile, "from-file", "", "Path to the secret file")
	secretCreateCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")
	secretCreateCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	secretCmd.AddCommand(secretCreateCmd)
}

func preRunSecretCreate(cmd *cobra.Command, args []string) error {

	return nil
}

func runSecretCreate(cmd *cobra.Command, args []string) error {
	var gatewayAddress string
	var err error

	if len(args) < 1 {
		return fmt.Errorf("please provide secret name")
	}
	secretName = args[0]

	if len(secretValue) == 0 && len(secretFile) > 0 {
		secretValue, err = readSecretFromFile(secretFile)
		if err != nil {
			return err
		}

	}

	if len(secretValue) == 0 && len(secretFile) == 0 {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			fmt.Fprintf(os.Stderr, "Reading from STDIN - hit (Control + D) to stop.\n")
		}

		secretStdin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		secretValue = string(secretStdin)
	}
	secretValue = strings.TrimSpace(secretValue)

	if len(secretValue) == 0 {
		return fmt.Errorf("must provide a non empty secret via --from-literal, --from-file or STDIN")
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
