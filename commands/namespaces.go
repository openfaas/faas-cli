package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/spf13/cobra"
)

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	namespacesCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	namespacesCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")
	namespacesCmd.Flags().StringVarP(&token, "token", "k", "", "Pass a JWT token to use instead of basic auth")

	faasCmd.AddCommand(namespacesCmd)
}

var namespacesCmd = &cobra.Command{
	Use:     `namespaces [--gateway GATEWAY_URL] [--tls-no-verify] [--token JWT_TOKEN]`,
	Aliases: []string{"ns"},
	Short:   "List OpenFaaS namespaces",
	Long:    `Lists OpenFaaS namespaces either on a local or remote gateway`,
	Example: `  faas-cli namespaces
  faas-cli namespaces --gateway https://127.0.0.1:8080`,
	RunE: runNamespaces,
}

func runNamespaces(cmd *cobra.Command, args []string) error {
	gatewayAddress := getGatewayURL(gateway, defaultGateway, "", os.Getenv(openFaaSURLEnvironment))
	cliAuth := NewCLIAuth(token, gatewayAddress)
	transport := GetDefaultCLITransport(tlsInsecure, &commandTimeout)
	client := proxy.NewClient(cliAuth, gatewayAddress, transport, &commandTimeout)
	namespaces, err := client.ListNamespaces(context.Background())
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
