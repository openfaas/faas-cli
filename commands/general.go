package commands

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/openfaas/faas-cli/config"
	"github.com/openfaas/faas-cli/proxy"
)

var (
	commandTimeout = 60 * time.Second
)

//CLIAuth auth struct for the CLI
type CLIAuth struct {
	Username string
	Password string
	Token    string
}

//BasicAuth basic authentication type
type BasicAuth struct {
	username string
	password string
}

func (auth *BasicAuth) Set(req *http.Request) error {
	req.SetBasicAuth(auth.username, auth.password)
	return nil
}

//BearerToken bearer token
type BearerToken struct {
	token string
}

func (c *BearerToken) Set(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+c.token)
	return nil
}

//NewCLIAuth returns a new CLI Auth
func NewCLIAuth(token string, gateway string) proxy.ClientAuth {
	authConfig, _ := config.LookupAuthConfig(gateway)

	var (
		username    string
		password    string
		bearerToken string
	)

	if authConfig.Auth == config.BasicAuthType {
		username, password, _ = config.DecodeAuth(authConfig.Token)

		return &BasicAuth{
			username: username,
			password: password,
		}

	}

	// User specified token gets priority
	if len(token) > 0 {
		bearerToken = token
	} else {
		bearerToken = authConfig.Token
	}

	return &BearerToken{
		token: bearerToken,
	}
}

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
