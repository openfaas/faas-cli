// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package analytics

import "net/http"

// UserSession contains common Google Analytics event values and
// the client used to submit the event.
type UserSession struct {
	HTTPClient         *http.Client `url:"-"`
	ProtocolVersion    int          `url:"v"`
	Type               string       `url:"t"`
	ClientID           string       `url:"cid"`
	TrackingID         string       `url:"tid"`
	ApplicationName    string       `url:"an"`
	ApplicationVersion string       `url:"av"`
	AnonymizeIP        int          `url:"aip"`
	Language           string       `url:"cd1"`
	OS                 string       `url:"cd2"`
	ARCH               string       `url:"cd3"`
	EventCategory      string       `url:"ec,omitempty"`
	EventAction        string       `url:"ea,omitempty"`
}
