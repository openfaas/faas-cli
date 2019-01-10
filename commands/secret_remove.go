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

var secretRemoveCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm"},
	Short:   "remove a secret",
	Long:    `Remove a secret by name`,
	Example: `faas-cli secret remove NAME
faas-cli secret remove NAME --gateway=http://127.0.0.1:8080`,
	RunE:    runSecretRemove,
	PreRunE: preRunSecretRemoveCmd,
}

func init() {
	secretRemoveCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	secretRemoveCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")

	secretCmd.AddCommand(secretRemoveCmd)
}

func preRunSecretRemoveCmd(cmd *cobra.Command, args []string) error {

	if len(args) == 0 {
		return fmt.Errorf("give a name of a secret")
	}

	if len(args) > 1 {
		return fmt.Errorf("give ONLY the name of a single secret")
	}
	return nil
}

func runSecretRemove(cmd *cobra.Command, args []string) error {
	var gatewayAddress string
	gatewayAddress = getGatewayURL(gateway, defaultGateway, "", os.Getenv(openFaaSURLEnvironment))

	secret := schema.Secret{
		Name: args[0],
	}
	err := proxy.RemoveSecret(gatewayAddress, secret, tlsInsecure)
	if err != nil {
		return err
	}

	fmt.Print("Removed.. OK.\n")

	return nil
}
