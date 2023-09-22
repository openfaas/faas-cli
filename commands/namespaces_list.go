package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	namespacesCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	namespacesCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")
	namespacesCmd.Flags().StringVarP(&token, "token", "k", "", "Pass a JWT token to use instead of basic auth")

	faasCmd.AddCommand(namespacesCmd)
	namespaceCmd.AddCommand(namespaceListCmd)
}

var namespacesCmd = &cobra.Command{
	Use:   `namespaces [--gateway GATEWAY_URL] [--tls-no-verify] [--token JWT_TOKEN]`,
	Short: "List OpenFaaS namespaces",
	Long:  `Lists OpenFaaS namespaces for the given gateway URL`,
	Example: `  faas-cli namespaces
  faas-cli namespaces --gateway https://127.0.0.1:8080`,
	RunE:       runNamespaces,
	Hidden:     true,
	Deprecated: "This has moved to \"faas-cli namespace list\".",
}

var namespaceListCmd = &cobra.Command{
	Use:     `list`,
	Aliases: []string{"ls"},
	Short:   "List OpenFaaS namespaces",
	Long:    `Lists OpenFaaS namespaces for the given gateway URL`,
	Example: `faas-cli namespace list`,
	RunE:    runNamespaces,
}

func runNamespaces(cmd *cobra.Command, args []string) error {
	client, err := GetDefaultSDKClient()
	if err != nil {
		return err
	}

	namespaces, err := client.GetNamespaces(context.Background())
	if err != nil {
		return err
	}
	printNamespaces(namespaces)
	return nil
}

func printNamespaces(namespaces []string) {
	fmt.Print("Namespaces:\n")
	for _, v := range namespaces {
		fmt.Printf(" - %s\n", v)
	}
}
