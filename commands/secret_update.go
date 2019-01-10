// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"os"

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
	Short:   "update a secret",
	Long:    `Update a secret by name`,
	Example: `faas-cli secret update NAME --from-literal=secret-value
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
	var gatewayAddress string
	gatewayAddress = getGatewayURL(gateway, defaultGateway, "", os.Getenv(openFaaSURLEnvironment))

	secret := schema.Secret{
		Name: args[0],
	}

	// todo(leodido) > catch secret value from stdin or flags

	err := proxy.UpdateSecret(gatewayAddress, secret, "", tlsInsecure)
	if err != nil {
		return err
	}

	// todo(leodido) > what message/feedback to the user?

	return nil
}
