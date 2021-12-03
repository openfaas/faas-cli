// Copyright (c) OpenFaaS Ltd 2021. All rights reserved.
//
// Licensed for use with OpenFaaS Pro only
// See EULA: https://github.com/openfaas/faas/blob/master/pro/EULA.md

package commands

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/openfaas/faas-cli/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	scope    string
	authURL  string
	tokenURL string

	clientID      string
	audience      string
	listenPort    int
	launchBrowser bool
	eula          bool
	grant         string
	clientSecret  string
	redirectHost  string
)

func init() {
	authCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	authCmd.Flags().StringVar(&authURL, "auth-url", "", "OAuth2 Authorize URL i.e. http://idp/oauth/authorize")
	authCmd.Flags().StringVar(&tokenURL, "token-url", "", "OAuth2 Token URL i.e. http://idp/oauth/token")

	authCmd.Flags().StringVar(&clientID, "client-id", "", "OAuth2 client_id")
	authCmd.Flags().IntVar(&listenPort, "listen-port", 31111, "OAuth2 local port for receiving cookie")
	authCmd.Flags().StringVar(&audience, "audience", "", "OAuth2 audience")
	authCmd.Flags().BoolVar(&launchBrowser, "launch-browser", true, "Launch browser for OAuth2 redirect")
	authCmd.Flags().StringVar(&redirectHost, "redirect-host", "http://127.0.0.1", "Host for OAuth2 redirection in the implicit flow including URL scheme")
	authCmd.Flags().BoolVar(&eula, "eula", false, "Agree to the EULA, for use with OpenFaaS Pro only")

	authCmd.Flags().StringVar(&scope, "scope", "openid profile", "scope for OAuth2 flow - i.e. \"openid profile\"")
	authCmd.Flags().StringVar(&grant, "grant", "implicit", "grant for OAuth2 flow - either implicit, implicit-id or client_credentials")
	authCmd.Flags().StringVar(&clientSecret, "client-secret", "", "OAuth2 client_secret, for use with client_credentials grant")

	faasCmd.AddCommand(authCmd)
}

var authCmd = &cobra.Command{
	Use: `auth --auth-url AUTH_URL | --client-id CLIENT_ID --scope SCOPE
  [--audience AUDIENCE]
  [--launch-browser LAUNCH_BROWSER]
  [--client-secret]
  [--grant GRANT]`,
	Short: "Obtain a token for your OpenFaaS gateway",
	Long: `Authenticate to an OpenFaaS gateway using OIDC.

Only licensed for use by OpenFaaS Pro customers.`,
	Example: `  faas-cli auth \
    --grant code \
    --client-id my-id \
    --auth-url https://tenant.auth0.com/authorize \
    --token-url https://tenant.auth0.com/oauth/token \
    --scope "oidc profile email"

  faas-cli auth --grant=client_credentials \
    --client-id=id \
    --client-secret=secret \
    --auth-url=https://tenant.auth0.com/oauth/token`,
	RunE:    runAuth,
	PreRunE: preRunAuth,
}

func preRunAuth(cmd *cobra.Command, args []string) error {
	return checkValues(authURL,
		clientID,
		eula,
	)
}

func checkValues(authURL, clientID string, eula bool) error {

	if len(authURL) == 0 {
		return fmt.Errorf("--auth-url is required and must be a valid OIDC URL")
	}

	u, uErr := url.Parse(authURL)
	if uErr != nil {
		return fmt.Errorf("--auth-url is an invalid URL: %s", uErr.Error())
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("--auth-url is an invalid URL: %s", u.String())
	}

	if len(clientID) == 0 {
		return fmt.Errorf("--client-id is required")
	}

	if !eula {
		return fmt.Errorf("the auth command is only licensed for OpenFaaS Pro customers, see: https://github.com/openfaas/faas/blob/master/pro/EULA.md")
	}

	return nil
}

func runAuth(cmd *cobra.Command, args []string) error {
	if grant == "implicit" {
		return authImplicit("token")
	} else if grant == "implicit-id" {
		return authImplicit("id_token")
	} else if grant == "client_credentials" {
		return authClientCredentials()
	}
	if grant == "code" {
		if len(tokenURL) == 0 {
			return fmt.Errorf("--token-url is required for PKCE")
		}
		return authPkce("id_token")
	}

	return nil
}

func authPkce(grant string) error {

	context, cancel := context.WithCancel(context.TODO())
	defer cancel()

	verifier := make([]byte, 32)

	_, err := rand.Read(verifier)
	if err != nil {
		return err
	}

	verifierEncoded := base64.RawURLEncoding.EncodeToString(verifier[:])

	challenge := sha256.Sum256([]byte(verifierEncoded))
	challengeEncoded := base64.RawURLEncoding.EncodeToString(challenge[:])

	q := url.Values{}
	q.Add("client_id", clientID)

	q.Add("state", fmt.Sprintf("%d", time.Now().UnixNano()))
	q.Add("nonce", fmt.Sprintf("%d", time.Now().UnixNano()))
	q.Add("scope", scope)
	q.Add("response_type", "code")
	q.Add("audience", audience)
	q.Add("code_challenge", challengeEncoded)
	q.Add("code_challenge_method", "S256")

	uri, err := makeRedirectURI(redirectHost, listenPort)
	if err != nil {
		return err
	}

	q.Add("redirect_uri", uri.String())

	authURLVal, _ := url.Parse(authURL)
	authURLVal.RawQuery = q.Encode()

	browserBase := authURLVal

	errCh := make(chan error, 1)

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", listenPort),
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
		Handler:        http.HandlerFunc(makeCodeCallbackHandler(cancel, errCh, verifierEncoded, clientID, uri.String())),
	}

	go func() {
		fmt.Printf("Starting local token server on port %d\n", listenPort)
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				panic(err)
			}
		}
	}()

	defer server.Shutdown(context)

	fmt.Printf("Launching browser: %s\n", browserBase)
	if launchBrowser {
		err := launchURL(browserBase.String())
		if err != nil {
			return errors.Wrap(err, "unable to launch browser")
		}
	}

	select {
	case <-context.Done():
		server.Shutdown(context)
	case serverErr := <-errCh:
		if serverErr != nil {
			return serverErr
		}
	}

	return nil
}

func authImplicit(grant string) error {

	context, cancel := context.WithCancel(context.TODO())
	defer cancel()

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", listenPort),
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
		Handler:        http.HandlerFunc(makeCallbackHandler(cancel)),
	}

	go func() {
		fmt.Printf("Starting local token server on port %d\n", listenPort)
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}

		select {
		case <-context.Done():
			break
		}
	}()

	defer server.Shutdown(context)

	q := url.Values{}
	q.Add("client_id", clientID)

	q.Add("state", fmt.Sprintf("%d", time.Now().UnixNano()))
	q.Add("nonce", fmt.Sprintf("%d", time.Now().UnixNano()))
	q.Add("response_type", grant)
	q.Add("scope", scope)
	q.Add("&response_mode", "fragment")
	q.Add("audience", audience)

	uri, err := makeRedirectURI(redirectHost, listenPort)
	if err != nil {
		return err
	}

	q.Add("redirect_uri", uri.String())

	authURLVal, _ := url.Parse(authURL)
	authURLVal.RawQuery = q.Encode()

	browserBase := authURLVal

	fmt.Printf("Launching browser: %s\n", browserBase)
	if launchBrowser {
		err := launchURL(browserBase.String())
		if err != nil {
			return errors.Wrap(err, "unable to launch browser")
		}
	}

	<-context.Done()

	return nil
}

func makeRedirectURI(host string, port int) (*url.URL, error) {
	val := fmt.Sprintf("%s/oauth/callback", fmt.Sprintf("%s:%d", host, port))
	res, err := url.Parse(val)

	if err != nil {
		return res, err
	}

	if st := res.String(); !(strings.HasPrefix(st, "http://") || strings.HasPrefix(st, "https://")) {
		return res, fmt.Errorf("a scheme is required for the URL, i.e. http://")
	}
	return res, err
}

func authClientCredentials() error {

	body := ClientCredentialsReq{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Audience:     audience,
		GrantType:    grant,
	}

	bodyBytes, marshalErr := json.Marshal(body)
	if marshalErr != nil {
		return errors.Wrapf(marshalErr, "unable to unmarshal %s", string(bodyBytes))
	}

	buf := bytes.NewBuffer(bodyBytes)
	req, _ := http.NewRequest(http.MethodPost, authURL, buf)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("cannot POST to %s", authURL))
	}
	if res.Body != nil {
		defer res.Body.Close()

		tokenData, _ := ioutil.ReadAll(res.Body)

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("cannot authenticate, code: %d.\nResponse: %s", res.StatusCode, string(tokenData))
		}
		token := AuthToken{}
		tokenErr := json.Unmarshal(tokenData, &token)
		if tokenErr != nil {
			return errors.Wrapf(tokenErr, "unable to unmarshal token: %s", string(tokenData))
		}

		if err := config.UpdateAuthConfig(gateway, token.AccessToken, config.Oauth2AuthType); err != nil {
			return err
		}
		fmt.Println("credentials saved for", gateway)
		printExampleTokenUsage(gateway, token.AccessToken)
	}

	return nil
}

// launchURL opens a URL with the default browser for Linux, MacOS or Windows.
func launchURL(serverURL string) error {
	ctx := context.Background()
	var command *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		command = exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf(`xdg-open "%s"`, serverURL))
	case "darwin":
		command = exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf(`open "%s"`, serverURL))
	case "windows":
		escaped := strings.Replace(serverURL, "&", "^&", -1)
		command = exec.CommandContext(ctx, "cmd", "/c", fmt.Sprintf(`start %s`, escaped))
	}
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	return command.Run()
}

func printExampleTokenUsage(gateway, token string) {
	fmt.Printf(`Example usage:
  # Use an explicit token
  faas-cli list --gateway "%s" --token "%s"

  # Use the saved token
  faas-cli list --gateway "%s"
`, gateway, token, gateway)

}

func makeCodeCallbackHandler(cancel context.CancelFunc, errCh chan error, verifierEncoded, clientID, redirectURI string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL)

		if r.URL.Path == "/oauth/callback" {
			code := r.URL.Query().Get("code")
			state := r.URL.Query().Get("state")

			v := url.Values{}
			v.Add("code", code)
			v.Add("state", state)
			v.Add("code_verifier", verifierEncoded)
			v.Add("grant_type", "authorization_code")
			v.Add("client_id", clientID)
			v.Add("redirect_uri", redirectURI)

			u, err := url.Parse(tokenURL)
			if err != nil {
				errCh <- err
				return
			}

			buf := bytes.NewBufferString(v.Encode())
			req, err := http.NewRequest(http.MethodPost, u.String(), buf)
			if err != nil {
				errCh <- err
				return
			}

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				errCh <- err
				return
			}

			tokenData, _ := ioutil.ReadAll(res.Body)

			if res.StatusCode != http.StatusOK {
				errCh <- fmt.Errorf("cannot authenticate, code: %d.\nResponse: %s", res.StatusCode, string(tokenData))
				return
			}

			token := AuthToken{}
			tokenErr := json.Unmarshal(tokenData, &token)
			if tokenErr != nil {
				errCh <- errors.Wrapf(tokenErr, "unable to unmarshal token: %s", string(tokenData))
				return
			}

			if err := config.UpdateAuthConfig(gateway, token.IDToken, config.Oauth2AuthType); err != nil {
				errCh <- err
				return
			}
			fmt.Println("credentials saved for", gateway)
			printExampleTokenUsage(gateway, token.IDToken)

			errCh <- nil
		}

	}
}

func makeCallbackHandler(cancel context.CancelFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if v := r.URL.Query().Get("fragment"); len(v) > 0 {

			q, err := url.ParseQuery(v)
			if err != nil {
				panic(errors.Wrap(err, "unable to parse fragment response from browser redirect"))
			}

			key := "id_token"
			if token := q.Get(key); len(token) > 0 {

				if err := config.UpdateAuthConfig(gateway, token, config.Oauth2AuthType); err != nil {
					fmt.Printf("error while saving authentication token: %s", err.Error())
				}
				fmt.Println("credentials saved for", gateway)
				printExampleTokenUsage(gateway, token)
			} else {
				fmt.Printf("Unable to detect a valid %s in URL fragment. Check your credentials or contact your administrator.\n", key)
			}

			cancel()
			return
		}

		if r.Body != nil {
			defer r.Body.Close()
		}
		w.Write([]byte(buildCaptureFragment()))
	}
}

func buildCaptureFragment() string {
	return `
<html>
<head>
<title>OpenFaaS CLI Authorization flow</title>
<script>
	var xhttp = new XMLHttpRequest();
	xhttp.onreadystatechange = function() {
		if (this.readyState == 4 && this.status == 200) {
			console.log(xhttp.responseText)
		}
	};

	// Encode the fragment data which could contain data that is query-string formatted
	xhttp.open("GET", "/oauth2/callback?fragment="+encodeURIComponent(document.location.hash.slice(1)), true);
	xhttp.send();
</script>
</head>
<body>
 Authorization flow complete. Please close this browser window.
</body>
</html>`
}

type ClientCredentialsReq struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Audience     string `json:"audience"`
	GrantType    string `json:"grant_type"`
}

type AuthToken struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`

	Scope     string `json:"scope"`
	ExpiresIn int    `json:"expires_in"`
	TokenType string `json:"token_type"`
}
