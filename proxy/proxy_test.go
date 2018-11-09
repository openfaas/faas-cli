// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"net/http"
	"testing"
	"time"
)

func Test_MakeHTTPClientWithDisableKeepAlives_(t *testing.T) {
	cases := []struct {
		name              string
		timeout           *time.Duration
		tlsInsecure       bool
		disableKeepAlives bool
		match             func(http.Client, *http.Transport) bool
	}{
		{name: "no timeout, secure, keep-alive", timeout: nil, tlsInsecure: false, disableKeepAlives: false, match: func(client http.Client, transport *http.Transport) bool {
			return transport == nil
		}},
		{name: "no timeout, insecure, keep-alive", timeout: nil, tlsInsecure: true, disableKeepAlives: false, match: func(client http.Client, transport *http.Transport) bool {
			return transport != nil &&
				transport.TLSClientConfig.InsecureSkipVerify &&
				transport.DisableKeepAlives == false &&
				transport.Proxy != nil
		}},
		{name: "no timeout, insecure, disable keep-alive", timeout: nil, tlsInsecure: true, disableKeepAlives: true, match: func(client http.Client, transport *http.Transport) bool {
			return transport != nil &&
				transport.TLSClientConfig.InsecureSkipVerify &&
				transport.DisableKeepAlives == true &&
				transport.Proxy != nil
		}},
		{name: "timeout, secure, keep-alive", timeout: durationPtr(time.Second * 30), tlsInsecure: false, disableKeepAlives: false, match: func(client http.Client, transport *http.Transport) bool {
			return client.Timeout == time.Second*30 &&
				transport != nil &&
				transport.DialContext != nil &&
				transport.TLSClientConfig == nil &&
				transport.DisableKeepAlives == false &&
				transport.Proxy != nil
		}},
		{name: "timeout, secure, disable keep-alive", timeout: durationPtr(time.Second * 30), tlsInsecure: false, disableKeepAlives: true, match: func(client http.Client, transport *http.Transport) bool {
			return client.Timeout == time.Second*30 &&
				transport != nil &&
				transport.DialContext != nil &&
				transport.TLSClientConfig == nil &&
				transport.DisableKeepAlives == true &&
				transport.Proxy != nil
		}},
		{name: "timeout, insecure, disable keep-alive", timeout: durationPtr(time.Second * 30), tlsInsecure: true, disableKeepAlives: true, match: func(client http.Client, transport *http.Transport) bool {
			return client.Timeout == time.Second*30 &&
				transport != nil &&
				transport.DialContext != nil &&
				transport.TLSClientConfig.InsecureSkipVerify &&
				transport.DisableKeepAlives == true &&
				transport.Proxy != nil
		}},
	}

	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			client := makeHTTPClientWithDisableKeepAlives(v.timeout, v.tlsInsecure, v.disableKeepAlives)
			var transport *http.Transport
			if client.Transport != nil {
				transport = client.Transport.(*http.Transport)
			}
			if !v.match(client, transport) {
				t.Logf("%s did not match", v.name)
				t.Fail()
			}
		})
	}
}

func durationPtr(duration time.Duration) *time.Duration {
	return &duration
}
