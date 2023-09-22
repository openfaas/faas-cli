// Copyright (c) OpenFaaS Author(s) 2023. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	namespaceCmd.PersistentFlags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	namespaceCmd.PersistentFlags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")
	namespaceCmd.PersistentFlags().StringVarP(&token, "token", "k", "", "Pass a JWT token to use instead of basic auth")

	faasCmd.AddCommand(namespaceCmd)
}

var namespaceCmd = &cobra.Command{
	Use:     `namespace [--gateway GATEWAY_URL] [--tls-no-verify] [--token JWT_TOKEN]`,
	Aliases: []string{"ns"},
	Short:   "Manage OpenFaaS namespaces",
	Long:    "Query, create, update, and delete OpenFaaS namespaces",
}
