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
		{name: "HTTPS gateway",
			args: args{gateway: "https://192.168.0.101:8080", tlsInsecure: false},
			want: ""},
		{name: "HTTPS gateway with TLSInsecure",
			args: args{gateway: "https://192.168.0.101:8080", tlsInsecure: true},
			want: ""},
		{name: "HTTP gateway without TLSInsecure",
			args: args{gateway: "http://192.168.0.101:8080", tlsInsecure: false},
			want: "WARNING! You are not using an encrypted connection to the gateway, consider using HTTPS."},
		{name: "HTTP gateway to 127.0.0.1 without TLSInsecure",
			args: args{gateway: "http://127.0.0.1:8080", tlsInsecure: false},
			want: ""},
		{name: "HTTP gateway to localhost without TLSInsecure",
			args: args{gateway: "http://localhost:8080", tlsInsecure: false},
			want: ""},
		{name: "HTTP gateway to remote host with TLSInsecure",
			args: args{gateway: "http://192.168.0.101:8080", tlsInsecure: true},
			want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkTLSInsecure(tt.args.gateway, tt.args.tlsInsecure)

			if got != tt.want {
				t.Errorf("[%s] want: %v, but got: %v", tt.name, tt.want, got)
			}
		})
	}
}
