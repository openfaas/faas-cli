package commands

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	generateLength int
	generateOutput string
)

var secretGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a random secret value",
	Long:  "Generate a cryptographically random secret suitable for HMAC payload signing or other shared secrets",
	Example: `  # Print a 32-byte base64-encoded secret to stdout
  faas-cli secret generate

  # Write to a file
  faas-cli secret generate -o payload.txt

  # Custom length in bytes
  faas-cli secret generate --length 64
`,
	RunE: runSecretGenerate,
}

func init() {
	secretGenerateCmd.Flags().IntVar(&generateLength, "length", 32, "Number of random bytes")
	secretGenerateCmd.Flags().StringVarP(&generateOutput, "output", "o", "", "Write to file instead of stdout")

	secretCmd.AddCommand(secretGenerateCmd)
}

func runSecretGenerate(cmd *cobra.Command, args []string) error {
	buf := make([]byte, generateLength)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Errorf("generating random bytes: %w", err)
	}

	secret := base64.StdEncoding.EncodeToString(buf)

	if generateOutput != "" {
		dir := filepath.Dir(generateOutput)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0700); err != nil {
				return fmt.Errorf("creating directory: %w", err)
			}
		}

		if err := os.WriteFile(generateOutput, []byte(secret), 0600); err != nil {
			return fmt.Errorf("writing secret: %w", err)
		}
		fmt.Printf("Wrote %d-byte secret to %s\n", generateLength, generateOutput)
	} else {
		fmt.Println(secret)
	}

	return nil
}
