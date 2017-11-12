// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package analytics

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"

	"github.com/google/go-querystring/query"
	"github.com/openfaas/faas-cli/version"
)

const gaHost = "www.google-analytics.com"
const trackingID = "UA-107707760-3"
const applicationName = "faas-cli3"

// disableEnvvar will prevent submission of analytics events if set
const disableEnvvar = "OPEN_FAAS_TELEMETRY"

// Event posts an analytics event to GA
func Event(action string, language string, ch chan int) {
	if Disabled() {
		return
	}
	u, err := NewSession(language)
	if err != nil {
		return
	}
	u.EventAction = action
	go u.PostEvent(ch)
}

// NewSession provides a setup UserSession struct with sane defaults
func NewSession(language string) (*UserSession, error) {
	if len(language) == 0 {
		language = "unset"
	}
	userSession := &UserSession{
		HTTPClient:         http.DefaultClient,
		ProtocolVersion:    1,
		Type:               "event",
		TrackingID:         trackingID,
		ApplicationName:    applicationName,
		ApplicationVersion: version.BuildVersion(),
		AnonymizeIP:        1,
		Language:           language,
		OS:                 runtime.GOOS,
		ARCH:               runtime.GOARCH,
		EventCategory:      "cli-success",
	}

	uuid, err := getUUID()
	if err != nil {
		uuid, err = setUUID()
		if err != nil {
			return nil, fmt.Errorf("Unable to get or set Analytics Client ID")
		}
	}
	userSession.ClientID = uuid

	return userSession, nil
}

// PostEvent submits the generated event to Google Analytics, it is a
// fire and forget process and ignores the response
func (u UserSession) PostEvent(ch chan int) {
	v, _ := query.Values(u)
	req := &http.Request{
		Method: "POST",
		Host:   gaHost,
		URL: &url.URL{
			Host:     gaHost,
			Scheme:   "https",
			Path:     "/collect",
			RawQuery: v.Encode(),
		},
	}
	u.HTTPClient.Do(req)
	ch <- 1
}

// Disabled returns true if the opt-out envvar has been set
// and false if it has not been set
func Disabled() bool {
	val, ok := os.LookupEnv(disableEnvvar)
	if ok && val == "0" {
		return true
	}
	return false
}
