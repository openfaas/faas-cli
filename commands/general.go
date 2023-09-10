package commands

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/openfaas/faas-cli/proxy"
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
	gatewayAddress := getGatewayURL(gateway, defaultGateway, "", os.Getenv(openFaaSURLEnvironment))
	gatewayURL, err := url.Parse(gatewayAddress)
	if err != nil {
		return nil, err
	}

	transport := GetDefaultCLITransport(tlsInsecure, &commandTimeout)

	httpClient := &http.Client{}
	httpClient.Timeout = commandTimeout

	if transport != nil {
		httpClient.Transport = transport
	}

	clientAuth, err := proxy.NewCLIAuth(token, gateway)
	if err != nil {
		return nil, err
	}

	client := sdk.NewClient(gatewayURL, clientAuth, http.DefaultClient)

	return client, nil
}
