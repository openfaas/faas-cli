// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// MakeHTTPClientWithDisableKeepAlives makes a HTTP client with good defaults for timeouts.
func MakeHTTPClientWithDisableKeepAlives(timeout *time.Duration, tlsInsecure bool, disableKeepAlives *bool) http.Client {
	if timeout != nil {
		tr := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout: *timeout,
			}).DialContext,

			IdleConnTimeout:       120 * time.Millisecond,
			ExpectContinueTimeout: 1500 * time.Millisecond,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: tlsInsecure},
		}
		if disableKeepAlives != nil {
			tr.DisableKeepAlives = *disableKeepAlives
		}

		return http.Client{
			Timeout:   *timeout,
			Transport: tr,
		}
	}
	return http.Client{}
}

// MakeHTTPClient makes a HTTP client with good defaults for timeouts.
func MakeHTTPClient(timeout *time.Duration, tlsInsecure bool) http.Client {
	return MakeHTTPClientWithDisableKeepAlives(timeout, tlsInsecure, nil)
}
