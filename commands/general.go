package commands

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/openfaas/faas-cli/config"
	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/go-sdk"
)

var (
	commandTimeout = 60 * time.Second
)

func GetDefaultCLITransport(tlsInsecure bool, timeout *time.Duration) *http.Transport {
	if timeout != nil || tlsInsecure {
		tr := &http.Transport{
			Proxy:             http.ProxyFromEnvironment,
			DisableKeepAlives: false,
		}

		if timeout != nil {
			tr.DialContext = (&net.Dialer{
				Timeout: *timeout,
			}).DialContext

			tr.IdleConnTimeout = 120 * time.Millisecond
			tr.ExpectContinueTimeout = 1500 * time.Millisecond
		}

		if tlsInsecure {
			tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: tlsInsecure}
		}
		tr.DisableKeepAlives = false

		return tr
	}
	return nil
}

func GetDefaultSDKClient() (*sdk.Client, error) {
	gatewayAddress, err := getGatewayAddress()
	if err != nil {
		return nil, err
	}

	gatewayURL, err := url.Parse(gatewayAddress)
	if err != nil {
		return nil, err
	}

	authConfig, err := config.LookupAuthConfig(gatewayURL.String())
	if err != nil {
		fmt.Printf("Failed to lookup auth config: %s\n", err)
	}

	var clientAuth sdk.ClientAuth
	var functionTokenSource sdk.TokenSource
	if authConfig.Auth == config.BasicAuthType {
		username, password, err := config.DecodeAuth(authConfig.Token)
		if err != nil {
			return nil, err
		}

		clientAuth = &sdk.BasicAuth{
			Username: username,
			Password: password,
		}
	}

	if authConfig.Auth == config.Oauth2AuthType {
		tokenAuth := &StaticTokenAuth{
			token: authConfig.Token,
		}

		clientAuth = tokenAuth
		functionTokenSource = tokenAuth
	}

	// User specified token gets priority
	if len(token) > 0 {
		tokenAuth := &StaticTokenAuth{
			token: token,
		}

		clientAuth = tokenAuth
		functionTokenSource = tokenAuth
	}

	httpClient := &http.Client{}
	httpClient.Timeout = commandTimeout

	transport := GetDefaultCLITransport(tlsInsecure, &commandTimeout)
	if transport != nil {
		httpClient.Transport = transport
	}

	return sdk.NewClientWithOpts(gatewayURL, httpClient,
		sdk.WithAuthentication(clientAuth),
		sdk.WithFunctionTokenSource(functionTokenSource),
	), nil
}

func getGatewayAddress() (string, error) {
	var yamlUrl string
	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAMLFile(yamlFile, regex, filter, envsubst)
		if err != nil {
			return "", err
		}

		if parsedServices != nil {
			yamlUrl = parsedServices.Provider.GatewayURL
		}
	}

	return getGatewayURL(gateway, defaultGateway, yamlUrl, os.Getenv(openFaaSURLEnvironment)), nil
}

type StaticTokenAuth struct {
	token string
}

func (a *StaticTokenAuth) Set(req *http.Request) error {
	req.Header.Add("Authorization", "Bearer "+a.token)
	return nil
}

func (ts *StaticTokenAuth) Token() (string, error) {
	return ts.token, nil
}
