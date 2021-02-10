// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/openfaas/faas/gateway/types"
)

//GetSystemInfo get system information from /system/info endpoint
func (c *Client) GetSystemInfo(ctx context.Context) (types.GatewayInfo, error) {
	infoEndPoint := "/system/info"
	var info types.GatewayInfo

	req, err := c.newRequest(http.MethodGet, infoEndPoint, nil)
	if err != nil {
		return info, fmt.Errorf("invalid HTTP method or invalid URL")
	}

	response, err := c.doRequest(ctx, req)
	if err != nil {
		return info, fmt.Errorf("cannot connect to OpenFaaS on URL: %s", c.GatewayURL.String())
	}

	if response.Body != nil {
		defer response.Body.Close()
	}

	switch response.StatusCode {
	case http.StatusOK:
		bytesOut, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return info, fmt.Errorf("cannot read result from OpenFaaS on URL: %s", c.GatewayURL.String())
		}
		err = json.Unmarshal(bytesOut, &info)
		if err != nil {
			return info, fmt.Errorf("cannot parse result from OpenFaaS on URL: %s\n%s", c.GatewayURL.String(), err.Error())
		}

	case http.StatusUnauthorized:
		return info, fmt.Errorf("unauthorized access, run \"faas-cli login\" to setup authentication for this server")
	default:
		bytesOut, err := ioutil.ReadAll(response.Body)
		if err == nil {
			return info, fmt.Errorf("server returned unexpected status code: %d - %s", response.StatusCode, string(bytesOut))
		}
	}

	return info, nil
}
