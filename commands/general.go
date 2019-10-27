package commands

import (
	"net/http"

	"github.com/openfaas/faas-cli/config"
	"github.com/openfaas/faas-cli/proxy"
)

type CLIAuth struct {
	Username string
	Password string
	Token    string
}

func NewCLIAuth(token string, gateway string) proxy.ClientAuth {
	username, password, _ := config.LookupAuthConfig(gateway)

	return &CLIAuth{
		Username: username,
		Password: password,
		Token:    token,
	}
}

func (c *CLIAuth) Set(req *http.Request) error {
	if len(c.Token) > 0 {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	} else {
		req.SetBasicAuth(c.Username, c.Password)
	}

	return nil
}
