package commands

import (
	"github.com/openfaas/faas-cli/config"
	"github.com/openfaas/faas-cli/proxy"
)

func GetProxyAuth(token string) proxy.Auth {
	username, password, _ := config.LookupAuthConfig(gateway)

	auth := proxy.Auth{
		Username: username,
		Password: password,
		Token:    token,
	}
	return auth
}
