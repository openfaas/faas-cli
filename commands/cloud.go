// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/openfaas/faas-cli/schema"
	"github.com/spf13/cobra"
)

var (
	namespace  string
	name       string
	literal    *[]string
	outputFile string
	fromFile   *[]string
	certFile   string
)

func init() {
	faasCmd.AddCommand(cloudCmd)

	cloudSealCmd.Flags().StringVar(&name, "name", "", "Secret name")

	cloudSealCmd.Flags().StringVarP(&namespace, "namespace", "n", "openfaas-fn", "Secret name")
	cloudSealCmd.Flags().StringVarP(&certFile, "cert", "c", "pub-cert.pem", "Filename of public certificate")

	cloudSealCmd.Flags().StringVarP(&outputFile, "output-file", "o", "secrets.yml", "Output file for secrets")

	literal = cloudSealCmd.Flags().StringArrayP("literal", "l", []string{}, "Secret literal key-value data")

	fromFile = cloudSealCmd.Flags().StringArrayP("from-file", "i", []string{}, "Read a secret from a from file")

	cloudCmd.AddCommand(cloudSealCmd)
}

var cloudCmd = &cobra.Command{
	Use:   `cloud`,
	Short: "OpenFaaS Cloud commands",
	Long:  "Commands for operating with OpenFaaS Cloud",
}

var cloudSealCmd = &cobra.Command{
	Use:     `seal [--name secret-name] [--literal k=v] [--namespace openfaas-fn]`,
	Short:   "Seal a secret for usage with OpenFaaS Cloud",
	Example: `  faas-cli cloud seal --name alexellis-github --literal hmac-secret=c4488af0c158e8c`,
	RunE:    runCloudSeal,
}

func runCloudSeal(cmd *cobra.Command, args []string) error {
	if len(name) == 0 {
		return fmt.Errorf("--name is required")
	}

	fmt.Printf("Sealing secret: %s in namespace: %s\n", name, namespace)

	fmt.Println("")

	enc := base64.StdEncoding

	secret := schema.KubernetesSecret{
		ApiVersion: "v1",
		Kind:       "Secret",
		Metadata: schema.KubernetesSecretMetadata{
			Name:      name,
			Namespace: namespace,
		},
		Data: make(map[string]string),
	}

	if literal != nil {
		args, _ := parseBuildArgs(*literal)

		for k, v := range args {
			secret.Data[k] = enc.EncodeToString([]byte(v))
		}
	}

	if fromFile != nil {
		for _, file := range *fromFile {
			bytesOut, err := ioutil.ReadFile(file)
			if err != nil {
				return err
			}

			key := filepath.Base(file)
			secret.Data[key] = enc.EncodeToString(bytesOut)
		}
	}

	sec, err := json.Marshal(secret)
	if err != nil {
		panic(err)
	}

	if _, err := os.Stat(certFile); err != nil {
		return fmt.Errorf("unable to load public certificate %s", certFile)
	}

	kubeseal := exec.Command("kubeseal", "--format=yaml", "--cert="+certFile)

	stdin, stdinErr := kubeseal.StdinPipe()
	if stdinErr != nil {
		panic(stdinErr)
	}

	stdin.Write(sec)
	stdin.Close()

	out, err := kubeseal.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unable to start \"kubeseal\", check it is installed, error: %s", err.Error())
	}

	writeErr := ioutil.WriteFile(outputFile, out, 0755)

	if writeErr != nil {
		return fmt.Errorf("unable to write secret: %s to %s", name, outputFile)
	}

	fmt.Printf("%s written.\n", outputFile)

	return nil

}
