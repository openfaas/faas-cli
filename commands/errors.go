// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"strings"
)

const (
	// NoTLSWarn Warning thrown when no SSL/TLS is used
	NoTLSWarn = "WARNING! You are not using an encrypted connection to the gateway, consider using HTTPS."
)

// checkTLSInsecure returns a warning message if the given gateway does not have https.
// Use tsInsecure to skip validations
func checkTLSInsecure(gateway string, tlsInsecure bool) string {
	if !tlsInsecure {
		if strings.HasPrefix(gateway, "https") == false &&
			strings.HasPrefix(gateway, "http://127.0.0.1") == false &&
			strings.HasPrefix(gateway, "http://localhost") == false {
			return NoTLSWarn
		}
	}
	return ""
}
