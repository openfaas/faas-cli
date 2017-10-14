package proxy

import (
	"net"
	"net/http"
	"time"
)

// MakeHTTPClient makes a HTTP client with good defaults for timeouts.
func MakeHTTPClient() http.Client {
	return http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout: 3 * time.Second,
				// KeepAlive: 0,
			}).DialContext,
			// MaxIdleConns:          1,
			// DisableKeepAlives:     true,
			IdleConnTimeout:       120 * time.Millisecond,
			ExpectContinueTimeout: 1500 * time.Millisecond,
		},
	}
}
