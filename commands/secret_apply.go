// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/openfaas/faas-cli/proxy"
	types "github.com/openfaas/faas-provider/types"
	"github.com/spf13/cobra"
)

var secretApplyCmd = &cobra.Command{
	Use:   `apply [--tls-no-verify]`,
	Short: "Apply secrets from .secrets folder",
	Long:  `Apply all secrets from the .secrets folder to the gateway. Each file in .secrets/ will be synced to the gateway, replacing existing secrets with the same name.`,
	Example: `  # Apply all secrets from .secrets folder
  faas-cli secret apply
  
  # Apply secrets to a specific namespace
  faas-cli secret apply --namespace=my-namespace
  
  # Apply secrets with a custom gateway
  faas-cli secret apply --gateway=http://127.0.0.1:8080`,
	RunE: runSecretApply,
}

func init() {
	secretApplyCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	secretApplyCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")
	secretApplyCmd.Flags().StringVarP(&token, "token", "k", "", "Pass a JWT token to use instead of basic auth")
	secretApplyCmd.Flags().StringVarP(&functionNamespace, "namespace", "n", "", "Namespace of the function")
	secretApplyCmd.Flags().BoolVar(&trimSecret, "trim", true, "Trim whitespace from the start and end of the secret value")

	secretCmd.AddCommand(secretApplyCmd)
}

func runSecretApply(cmd *cobra.Command, args []string) error {
	gatewayAddress := getGatewayURL(gateway, defaultGateway, "", os.Getenv(openFaaSURLEnvironment))

	if msg := checkTLSInsecure(gatewayAddress, tlsInsecure); len(msg) > 0 {
		fmt.Println(msg)
	}

	cliAuth, err := proxy.NewCLIAuth(token, gatewayAddress)
	if err != nil {
		return err
	}
	transport := GetDefaultCLITransport(tlsInsecure, &commandTimeout)
	client, err := proxy.NewClient(cliAuth, gatewayAddress, transport, &commandTimeout)
	if err != nil {
		return err
	}

	// Get the absolute path to .secrets directory
	secretsPath, err := filepath.Abs(localSecretsDir)
	if err != nil {
		return fmt.Errorf("can't determine secrets folder: %w", err)
	}

	// Check if .secrets directory exists
	if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
		return fmt.Errorf("secrets directory does not exist: %s", secretsPath)
	}

	// Read all files from .secrets directory
	files, err := os.ReadDir(secretsPath)
	if err != nil {
		return fmt.Errorf("failed to read secrets directory: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No secrets found in .secrets directory")
		return nil
	}

	// Get list of existing secrets from gateway to check if they exist
	existingSecrets, err := client.GetSecretList(context.Background(), functionNamespace)
	if err != nil {
		return fmt.Errorf("failed to get secret list: %w", err)
	}

	// Create a map of existing secrets for quick lookup
	secretMap := make(map[string]bool)
	for _, secret := range existingSecrets {
		// Match by name and namespace
		if secret.Namespace == functionNamespace {
			secretMap[secret.Name] = true
		}
	}

	// Process each file in .secrets directory
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		secretName := file.Name()

		// Validate secret name
		isValid, err := validateSecretName(secretName)
		if !isValid {
			fmt.Printf("Skipping invalid secret name: %s - %v\n", secretName, err)
			continue
		}

		// Read secret file content
		secretFilePath := filepath.Join(secretsPath, secretName)
		fileData, err := os.ReadFile(secretFilePath)
		if err != nil {
			fmt.Printf("Failed to read secret file %s: %v\n", secretName, err)
			continue
		}

		secretValue := string(fileData)
		if trimSecret {
			secretValue = strings.TrimSpace(secretValue)
		}

		if len(secretValue) == 0 {
			fmt.Printf("Skipping empty secret: %s\n", secretName)
			continue
		}

		secret := types.Secret{
			Name:      secretName,
			Namespace: functionNamespace,
			Value:     secretValue,
			RawValue:  fileData,
		}

		// Check if secret exists and delete it if it does
		if secretMap[secretName] {
			fmt.Printf("Secret %s exists, deleting before recreating...\n", secretName)
			deleteSecret := types.Secret{
				Name:      secretName,
				Namespace: functionNamespace,
			}
			err = client.RemoveSecret(context.Background(), deleteSecret)
			if err != nil {
				fmt.Printf("Failed to remove existing secret %s: %v\n", secretName, err)
				// Continue anyway to try creating it
			}
		}

		// Create the secret
		fmt.Printf("Creating secret: %s.%s\n", secret.Name, functionNamespace)
		status, output := client.CreateSecret(context.Background(), secret)

		if status == http.StatusConflict {
			// If secret still exists (race condition), try update instead
			fmt.Printf("Secret %s still exists, updating...\n", secretName)
			_, output = client.UpdateSecret(context.Background(), secret)
		}

		fmt.Print(output)
	}

	return nil
}
