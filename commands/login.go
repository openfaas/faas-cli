// Copyright (c) OpenFaaS Author(s) 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/openfaas/faas-cli/proxy"

	"github.com/openfaas/faas-cli/config"
	"github.com/spf13/cobra"
)

var (
	username      string
	password      string
	passwordStdin bool
)

func init() {
	loginCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	loginCmd.Flags().StringVarP(&username, "username", "u", "admin", "Gateway username")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "Gateway password")
	loginCmd.Flags().BoolVarP(&passwordStdin, "password-stdin", "s", false, "Reads the gateway password from stdin")
	loginCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")

	faasCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   `login [--username admin|USERNAME] [--password PASSWORD] [--gateway GATEWAY_URL] [--tls-no-verify]`,
	Short: "Log in to OpenFaaS gateway",
	Long:  "Log in to OpenFaaS gateway.\nIf no gateway is specified, the default value will be used.",
	Example: `  cat ~/faas_pass.txt | faas-cli login -u user --password-stdin
  echo $PASSWORD | faas-cli login -s  --gateway https://openfaas.mydomain.com
  faas-cli login -u user -p password`,
	RunE: runLogin,
}

func runLogin(cmd *cobra.Command, args []string) error {

	if len(username) == 0 {
		return fmt.Errorf("must provide --username or -u")
	}

	if len(password) > 0 {
		fmt.Println("WARNING! Using --password is insecure, consider using: cat ~/faas_pass.txt | faas-cli login -u user --password-stdin")
		if passwordStdin {
			return fmt.Errorf("--password and --password-stdin are mutually exclusive")
		}

		if len(username) == 0 {
			return fmt.Errorf("must provide --username with --password")
		}
	}

	if passwordStdin {
		if len(username) == 0 {
			return fmt.Errorf("must provide --username with --password-stdin")
		}

		passwordStdin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		password = strings.TrimSpace(string(passwordStdin))
	}

	password = strings.TrimSpace(password)
	if len(password) == 0 {
		return fmt.Errorf("must provide a non-empty password via --password or --password-stdin")
	}

	fmt.Println("Calling the OpenFaaS server to validate the credentials...")

	gateway = getGatewayURL(gateway, defaultGateway, "", os.Getenv(openFaaSURLEnvironment))

	if err := validateLogin(gateway, username, password); err != nil {
		return err
	}

	token := config.EncodeAuth(username, password)
	if err := config.UpdateAuthConfig(gateway, token, config.BasicAuthType); err != nil {
		return err
	}

	authConfig, err := config.LookupAuthConfig(gateway)
	if err != nil {
		return err
	}

	user, _, err := config.DecodeAuth(authConfig.Token)
	if err != nil {
		return err
	}
	fmt.Println("credentials saved for", user, gateway)

	return nil
}

func validateLogin(gatewayURL string, user string, pass string) error {
	timeout := time.Duration(5 * time.Second)
	client := proxy.MakeHTTPClient(&timeout, tlsInsecure)

	req, err := http.NewRequest("GET", gatewayURL+"/system/functions", nil)
	if err != nil {
		return fmt.Errorf("invalid URL: %s", gatewayURL)
	}

	req.SetBasicAuth(user, pass)
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot connect to OpenFaaS on URL: %s. %v", gatewayURL, err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.TLS == nil {
		fmt.Println("WARNING! Communication is not secure, please consider using HTTPS. Letsencrypt.org offers free SSL/TLS certificates.")
	}

	switch res.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("unable to login, either username or password is incorrect")
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return fmt.Errorf("server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}

	return nil
}
