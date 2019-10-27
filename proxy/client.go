package proxy

import (
	"io"
	"net/http"
)

type Auth struct {
	Username string
	Password string
	Token    string
}

type Client struct {
	Auth *Auth
}

func NewClient(auth Auth) *Client {
	return &Client{
		Auth: &auth,
	}
}

func (c *Client) NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if len(c.Auth.Token) > 0 {
		req.Header.Set("Authorization", "Bearer "+c.Auth.Token)
	} else {
		req.SetBasicAuth(c.Auth.Username, c.Auth.Password)
	}

	return req, err
}
