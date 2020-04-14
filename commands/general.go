package commands

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
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
