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
	secretValue string
	secretFile  string
)

var secretUpdateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"replace"},
	Short:   "Update a secret",
	Long:    `Update a secret by name`,
	Example: `faas-cli secret update NAME
faas-cli secret update NAME --from-literal=secret-value
faas-cli secret update NAME --from-file=/path/to/secret/file
faas-cli secret update NAME --from-literal=secret-value --gateway=http://127.0.0.1:8080
cat /path/to/secret/file | faas-cli secret update NAME`,
	RunE:    runSecretUpdate,
	PreRunE: preRunSecretUpdateCmd,
}

func init() {
	secretUpdateCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	secretUpdateCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")
	secretUpdateCmd.Flags().StringVar(&secretValue, "from-literal", "", "Value of the secret")
	secretUpdateCmd.Flags().StringVar(&secretFile, "from-file", "", "Path to the secret file")

	secretCmd.AddCommand(secretUpdateCmd)
}

func preRunSecretUpdateCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("give a name of a secret")
	}

	if len(args) > 1 {
		return fmt.Errorf("give ONLY the name of a single secret")
	}

	return nil
}

func runSecretUpdate(cmd *cobra.Command, args []string) error {
	gatewayAddress := getGatewayURL(gateway, defaultGateway, "", os.Getenv(openFaaSURLEnvironment))

	secret := schema.Secret{
		Name: args[0],
	}

	if len(secretValue) == 0 {
		if len(secretFile) == 0 {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				fmt.Fprintf(os.Stderr, "Reading from STDIN - hit (Control + D) to stop.\n")
			}
			input, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("unable to read standard input: %s", err.Error())
			}
			secretValue = string(input)
		} else {
			input, err := ioutil.ReadFile(secretFile)
			if err != nil {
				return fmt.Errorf("unable to read secret file: %s", err.Error())
			}
			secretValue = string(input)
		}
	}
	secretValue = strings.TrimSpace(secretValue)

	err := proxy.UpdateSecret(gatewayAddress, secret, secretValue, tlsInsecure)
	if err != nil {
		return err
	}

	fmt.Printf("Secret %q updated.\n", secret.Name)

	return nil
}
