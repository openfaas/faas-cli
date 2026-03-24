package commands

import (
	"fmt"
	"os"

	"github.com/openfaas/go-sdk/seal"
	"github.com/spf13/cobra"
)

var (
	unsealInput string
	unsealKey   string
)

var secretUnsealCmd = &cobra.Command{
	Use:   "unseal [private-key-file]",
	Short: "Unseal and inspect a sealed secrets file",
	Long:  "Decrypt a sealed secrets file using a private key and print the key/value pairs",
	Example: `  # Print all secrets
  faas-cli secret unseal key

  # Print a single secret value
  faas-cli secret unseal key --key pip_token

  # Specify a different sealed file
  faas-cli secret unseal key --in ./build/com.openfaas.secrets
`,
	Args: cobra.ExactArgs(1),
	RunE: runSecretUnseal,
}

func init() {
	secretUnsealCmd.Flags().StringVar(&unsealInput, "in", "com.openfaas.secrets", "Path to the sealed secrets file")
	secretUnsealCmd.Flags().StringVar(&unsealKey, "key", "", "Unseal a single key (omit to print all)")

	secretCmd.AddCommand(secretUnsealCmd)
}

func runSecretUnseal(cmd *cobra.Command, args []string) error {
	privKey, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("reading private key: %w", err)
	}

	envelope, err := os.ReadFile(unsealInput)
	if err != nil {
		return fmt.Errorf("reading sealed file: %w", err)
	}

	if unsealKey != "" {
		value, err := seal.UnsealKey(privKey, envelope, unsealKey)
		if err != nil {
			return err
		}
		fmt.Print(string(value))
		return nil
	}

	values, err := seal.Unseal(privKey, envelope)
	if err != nil {
		return err
	}

	for k, v := range values {
		fmt.Printf("%s=%s\n", k, string(v))
	}

	return nil
}
