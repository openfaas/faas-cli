package sdk

import (
	"net/http"
)

// BasicAuth basic authentication for the the OpenFaaS client
type BasicAuth struct {
	Username string
	Password string
}

// Set Authorization Basic header on request
func (auth *BasicAuth) Set(req *http.Request) error {
	req.SetBasicAuth(auth.Username, auth.Password)
	return nil
}
