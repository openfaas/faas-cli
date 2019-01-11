// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/schema"
	"github.com/spf13/cobra"
)

var (
	literalSecret string
	secretFile    string
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
	secretCreateCmd.Flags().StringVar(&literalSecret, "from-literal", "", "Value of the secret")
	secretCreateCmd.Flags().StringVar(&secretFile, "from-file", "", "Path to the secret file")
	secretCreateCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")
	secretCreateCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	secretCmd.AddCommand(secretCreateCmd)
}

func preRunSecretCreate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("secret name required")
	}

	if len(args) > 1 {
		return fmt.Errorf("too many values for secret name")
	}

	if len(secretFile) > 0 && len(literalSecret) > 0 {
		return fmt.Errorf("please provide secret using only one option from --from-literal, --from-file and STDIN")
	}

	return nil
}

func runSecretCreate(cmd *cobra.Command, args []string) error {
	secret := schema.Secret{
		Name: args[0],
	}

	switch {
	case len(literalSecret) > 0:
		secret.Value = literalSecret

	case len(secretFile) > 0:
		var err error
		secret.Value, err = readSecretFromFile(secretFile)
		if err != nil {
			return err
		}

	default:
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			fmt.Fprintf(os.Stderr, "Reading from STDIN - hit (Control + D) to stop.\n")
		}

		secretStdin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		secret.Value = string(secretStdin)
	}

	secret.Value = strings.TrimSpace(secret.Value)

	if len(secret.Value) == 0 {
		return fmt.Errorf("must provide a non empty secret via --from-literal, --from-file or STDIN")
	}

	gatewayAddress := getGatewayURL(gateway, defaultGateway, "", os.Getenv(openFaaSURLEnvironment))

	fmt.Println("Creating secret: " + secret.Name)
	_, output := proxy.CreateSecret(gatewayAddress, secret, tlsInsecure)
	fmt.Printf(output)

	return nil
}

func readSecretFromFile(secretFile string) (string, error) {
	fileData, err := ioutil.ReadFile(secretFile)
	return string(fileData), err
}
