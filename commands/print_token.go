// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	faasCmd.AddCommand(printToken)
}

var printToken = &cobra.Command{
	Use:   `print-token ./token.txt`,
	Short: "Pretty-print the contents of a JWT token",
	Example: `  # Print the contents of a JWT token
  faas-cli print-token ./token.txt
`,
	RunE:   runPrintTokenE,
	Hidden: true,
}

func runPrintTokenE(cmd *cobra.Command, args []string) error {

	if len(args) < 1 {
		return fmt.Errorf("provide the filename as an argument i.e. faas-cli print-token ./token.txt")
	}

	tokenFile := args[0]

	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return err
	}

	token := string(data)

	jwtToken, err := unmarshalJwt(token)
	if err != nil {
		return err
	}

	j, err := json.MarshalIndent(jwtToken, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(j))

	return nil
}

type JwtToken struct {
	Header  map[string]interface{} `json:"header"`
	Payload map[string]interface{} `json:"payload"`
}

func unmarshalJwt(token string) (JwtToken, error) {

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return JwtToken{}, fmt.Errorf("token should have 3 parts, got %d", len(parts))
	}

	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return JwtToken{}, err
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return JwtToken{}, err
	}

	var jwt JwtToken

	err = json.Unmarshal(header, &jwt.Header)
	if err != nil {
		return JwtToken{}, err
	}

	err = json.Unmarshal(payload, &jwt.Payload)
	if err != nil {
		return JwtToken{}, err
	}

	return jwt, nil
}
