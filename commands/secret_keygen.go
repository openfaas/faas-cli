package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openfaas/go-sdk/seal"
	"github.com/spf13/cobra"
)

var keygenOutput string

var secretKeygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate a keypair for sealing build secrets",
	Long:  "Generate a Curve25519 keypair for use with faas-cli secret seal and the pro-builder",
	Example: `  # Generate key and key.pub in the current directory
  faas-cli secret keygen

  # Generate mykey and mykey.pub in a specific directory
  faas-cli secret keygen -o ./keys/mykey
`,
	RunE: runSecretKeygen,
}

func init() {
	secretKeygenCmd.Flags().StringVarP(&keygenOutput, "output", "o", "key", "Output path for the private key (public key gets .pub appended)")

	secretCmd.AddCommand(secretKeygenCmd)
}

func runSecretKeygen(cmd *cobra.Command, args []string) error {
	pub, priv, err := seal.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("generating keypair: %w", err)
	}

	privPath := keygenOutput
	pubPath := keygenOutput + ".pub"

	dir := filepath.Dir(privPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	if err := os.WriteFile(privPath, priv, 0600); err != nil {
		return fmt.Errorf("writing private key: %w", err)
	}

	if err := os.WriteFile(pubPath, pub, 0644); err != nil {
		return fmt.Errorf("writing public key: %w", err)
	}

	keyID, err := seal.DeriveKeyID(pub)
	if err != nil {
		return fmt.Errorf("deriving key ID: %w", err)
	}

	fmt.Printf("Wrote private key: %s\n", privPath)
	fmt.Printf("Wrote public key:  %s\n", pubPath)
	fmt.Printf("Key ID:            %s\n", keyID)

	return nil
}
