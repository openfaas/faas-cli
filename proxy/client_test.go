package proxy

import (
	"testing"
)

func Test_NewClient(t *testing.T) {
	auth := NewTestAuth(nil)
	testcases := []struct {
		Name   string
		Input  string
		Output string
	}{
		{
			Name:   "Without trailing slash",
			Input:  "http://127.0.0.1:8080",
			Output: "http://127.0.0.1:8080",
		},
		{
			Name:   "With trailing slash",
			Input:  "http://127.0.0.1:8080/",
			Output: "http://127.0.0.1:8080",
		},
	}

	for _, test := range testcases {
		newClient := NewClient(auth, test.Input, nil, &defaultCommandTimeout)
		clientURL := newClient.GatewayURL.String()
		if clientURL != test.Output {
			t.Fatalf("Testcase %s failed. Expected: %s, Got: %s", test.Name, test.Output, clientURL)
		}
	}
}

func Test_newRequest_URL(t *testing.T) {
	auth := NewTestAuth(nil)
	gatewayURL := "http://127.0.0.1:8080"
	client := NewClient(auth, gatewayURL, nil, &defaultCommandTimeout)

	testcases := []struct {
		Name        string
		Path        string
		ExpectedURL string
	}{
		{
			Name:        "A valid path",
			Path:        "/system/functions",
			ExpectedURL: "http://127.0.0.1:8080/system/functions",
		},
		{
			Name:        "Root Path",
			Path:        "/",
			ExpectedURL: "http://127.0.0.1:8080/",
		},
		{
			Name:        "Path without starting slash",
			Path:        "system/functions",
			ExpectedURL: "http://127.0.0.1:8080/system/functions",
		},
		{
			Name:        "Path with querystring",
			Path:        "system/functions?namespace=fn",
			ExpectedURL: "http://127.0.0.1:8080/system/functions?namespace=fn",
		},
	}

	for _, test := range testcases {
		request, err := client.newRequest("POST", test.Path, nil)
		if err != nil {
			t.Fatalf("Got Error! %s", err.Error())
		}

		url := request.URL.String()
		if url != test.ExpectedURL {
			t.Fatalf("Testcase %s failed. Expected: %s, Got: %s", test.Name, test.ExpectedURL, url)
		}
	}

}

func Test_addQueryParams(t *testing.T) {

	testcases := []struct {
		Name        string
		QueryParams map[string]string
		URL         string
		ExpectedURL string
	}{
		{
			Name:        "URL without hostname",
			QueryParams: map[string]string{"namespace": "openfaas-fn"},
			URL:         "/system/functions",
			ExpectedURL: "/system/functions?namespace=openfaas-fn",
		},
		{
			Name:        "URL hostname",
			QueryParams: map[string]string{"namespace": "openfaas-fn"},
			URL:         "http://127.0.0.1/system/functions",
			ExpectedURL: "http://127.0.0.1/system/functions?namespace=openfaas-fn",
		},
		{
			Name:        "A URL with simple hostname",
			QueryParams: map[string]string{"namespace": "openfaas-fn"},
			URL:         "example",
			ExpectedURL: "example?namespace=openfaas-fn",
		},
	}

	for _, test := range testcases {
		url, err := addQueryParams(test.URL, test.QueryParams)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if url != test.ExpectedURL {
			t.Fatalf("Testcase %s failed, Expected: %s, Got: %s", test.Name, test.ExpectedURL, url)
		}
	}
}
