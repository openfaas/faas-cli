// Copyright (c) OpenFaaS Project 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"strings"

	"github.com/openfaas/faas-cli/config"
	"github.com/spf13/cobra"
)

func init() {
	logoutCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")

	faasCmd.AddCommand(logoutCmd)
}

var logoutCmd = &cobra.Command{
	Use:     `logout [--gateway GATEWAY_URL]`,
	Short:   "Log out from OpenFaaS gateway",
	Long:    "Log out from OpenFaaS gateway.\nIf no gateway is specified, the default local one will be used.",
	Example: `  faas-cli logout --gateway https://openfaas.mydomain.com`,
	RunE:    runLogout,
}

func runLogout(cmd *cobra.Command, args []string) error {
	if len(gateway) == 0 {
		return fmt.Errorf("gateway cannot be an empty string")
	}

	gateway = strings.TrimRight(strings.TrimSpace(gateway), "/")
	err := config.RemoveAuthConfig(gateway)
	if err != nil {
		return err
	}
	fmt.Println("credentials removed for", gateway)

	return nil
}
