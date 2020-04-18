// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"fmt"
	v2 "github.com/openfaas/faas-cli/schema/store/v2"
	"github.com/openfaas/faas-cli/test"
	"net/http"
	"reflect"
	"testing"
)

const testStack = `
{
    "version": "0.2.0",
    "functions": [
    {
        "title": "NodeInfo",
        "name": "nodeinfo",
        "description": "Get info about the machine that you're deployed on. Tells CPU count, hostname, OS, and Uptime",
        "images": {
            "arm64": "functions/nodeinfo:arm64",
            "armhf": "functions/nodeinfo-http:latest-armhf",
            "x86_64": "functions/nodeinfo-http:latest"
        },
        "repo_url": "https://github.com/openfaas/faas/tree/master/sample-functions/NodeInfo"
    }]
}
`

func Test_Generate(t *testing.T) {
	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/functions",
			ResponseBody:       testStack,
			ResponseStatusCode: http.StatusOK,
		},
	})
	defer s.Close()
	u := fmt.Sprintf("%s%s", s.URL, "/functions")

	got, err := FunctionStoreList(u)
	if err != nil {
		t.Fatalf("err was not nill, %s", err.Error())
	}

	want := []v2.StoreFunction{{
		Title:                  "NodeInfo",
		Name:                   "nodeinfo",
		Description:            "Get info about the machine that you're deployed on. Tells CPU count, hostname, OS, and Uptime",
		Images:                 map[string]string{"arm64": "functions/nodeinfo:arm64", "armhf": "functions/nodeinfo-http:latest-armhf", "x86_64": "functions/nodeinfo-http:latest"},
		RepoURL:                "https://github.com/openfaas/faas/tree/master/sample-functions/NodeInfo",
		ReadOnlyRootFilesystem: false,
		Environment:            nil,
		Labels:                 nil,
		Annotations:            nil,
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %v, \nwant %v", got, want)
	}
}
