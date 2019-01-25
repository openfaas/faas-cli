// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package commands

import "testing"

func Test_checkTLSInsecure(t *testing.T) {
	type args struct {
		gateway     string
		tlsInsecure bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "secure gateway and tls secure", args: args{gateway: "https://127.0.0.1:8080", tlsInsecure: false}, want: ""},
		{name: "secure gateway and tls insecure", args: args{gateway: "https://127.0.0.1:8080", tlsInsecure: true}, want: ""},
		{name: "insecure gateway and tls secure", args: args{gateway: "http://127.0.0.1:8080", tlsInsecure: false}, want: "WARNING! Communication is not secure, please consider using HTTPS. Letsencrypt.org offers free SSL/TLS certificates."},
		{name: "insecure gateway and tls insecure", args: args{gateway: "http://127.0.0.1:8080", tlsInsecure: true}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkTLSInsecure(tt.args.gateway, tt.args.tlsInsecure); got != tt.want {
				t.Errorf("checkTLSInsecure() = %v, want %v", got, tt.want)
			}
		})
	}
}
