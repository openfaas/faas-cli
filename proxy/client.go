package proxy

import (
	"io"
	"net/http"
)

type Client struct {
	ClientAuth ClientAuth
}

type ClientAuth interface {
	Set(req *http.Request) error
}

func NewClient(auth ClientAuth) *Client {
	return &Client{
		ClientAuth: auth,
	}
}

type Auth struct {
	Username string
	Password string
	Token    string
}

func (c *Client) NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	c.ClientAuth.Set(req)

	return req, err
}
