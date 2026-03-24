package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/openfaas/go-sdk/seal"
	"github.com/spf13/cobra"
)

var (
	sealKeyID       string
	sealOutput      string
	sealFromLiteral []string
	sealFromFile    []string
)

var secretSealCmd = &cobra.Command{
	Use:   "seal [public-key-file]",
	Short: "Seal build secrets into an encrypted file",
	Long:  "Seal key/value pairs using a public key. The output file can be included in a build tar or committed to git.",
	Example: `  # Seal literal values
  faas-cli secret seal key.pub \
    --from-literal pip_token=s3cr3t \
    --from-literal npm_token=tok123

  # Seal from files (binary-safe)
  faas-cli secret seal key.pub \
    --from-file ca.crt=./certs/ca.crt \
    --from-literal api_key=sk-1234

  # Specify key ID and output path
  faas-cli secret seal key.pub \
    --key-id builder-key-1 \
    --from-literal token=s3cr3t \
    -o ./build/com.openfaas.secrets
`,
	Args:    cobra.ExactArgs(1),
	RunE:    runSecretSeal,
	PreRunE: preRunSecretSeal,
}

func init() {
	secretSealCmd.Flags().StringVar(&sealKeyID, "key-id", "", "Key ID for rotation tracking (optional)")
	secretSealCmd.Flags().StringVarP(&sealOutput, "output", "o", "com.openfaas.secrets", "Output file path")
	secretSealCmd.Flags().StringArrayVar(&sealFromLiteral, "from-literal", nil, "Literal secret in key=value format (can be repeated)")
	secretSealCmd.Flags().StringArrayVar(&sealFromFile, "from-file", nil, "Secret from file in key=path format (can be repeated)")

	secretCmd.AddCommand(secretSealCmd)
}

func preRunSecretSeal(cmd *cobra.Command, args []string) error {
	if len(sealFromLiteral) == 0 && len(sealFromFile) == 0 {
		return fmt.Errorf("provide at least one secret via --from-literal or --from-file")
	}

	return nil
}

func runSecretSeal(cmd *cobra.Command, args []string) error {
	pubKey, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("reading public key: %w", err)
	}

	values := make(map[string][]byte)

	for _, lit := range sealFromLiteral {
		k, v, ok := strings.Cut(lit, "=")
		if !ok || k == "" {
			return fmt.Errorf("invalid --from-literal format %q, expected key=value", lit)
		}
		values[k] = []byte(v)
	}

	for _, f := range sealFromFile {
		k, path, ok := strings.Cut(f, "=")
		if !ok || k == "" || path == "" {
			return fmt.Errorf("invalid --from-file format %q, expected key=path", f)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading file for key %q: %w", k, err)
		}
		values[k] = data
	}

	sealed, err := seal.Seal(pubKey, values, sealKeyID)
	if err != nil {
		return fmt.Errorf("sealing secrets: %w", err)
	}

	if err := os.WriteFile(sealOutput, sealed, 0600); err != nil {
		return fmt.Errorf("writing sealed file: %w", err)
	}

	fmt.Printf("Sealed %d secret(s) to %s\n", len(values), sealOutput)

	return nil
}
