// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"os"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/spf13/cobra"
)

// secretListCmd represents the secretCreate command
var secretListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secrets",
	Long:  `List all secrets names and metadata`,
	Example: `faas-cli secret list
faas-cli secret list --gateway=http://127.0.0.1:8080`,
	RunE:    runSecretList,
	PreRunE: preRunSecretListCmd,
}

func init() {
	secretListCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	// secretListCmd.Flags().BoolVarP(&verboseList, "verbose", "v", false, "Verbose output for the function list")
	secretListCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")

	secretCmd.AddCommand(secretListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// secretCreateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// secretCreateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func preRunSecretListCmd(cmd *cobra.Command, args []string) error {
	return nil
}

func runSecretList(cmd *cobra.Command, args []string) error {
	var gatewayAddress string
	gatewayAddress = getGatewayURL(gateway, defaultGateway, "", os.Getenv(openFaaSURLEnvironment))

	secrets, err := proxy.GetSecretList(gatewayAddress, tlsInsecure)
	if err != nil {
		return err
	}

	for _, secret := range secrets {
		fmt.Println(secret.Name)
	}

	return nil
}
