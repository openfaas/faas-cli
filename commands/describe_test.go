package commands

import "testing"

func Test_getFunctionURLs(t *testing.T) {
	cases := []struct {
		name              string
		gateway           string
		functionName      string
		functionNamespace string
		expectedURL       string
		expectedAsyncURL  string
	}{
		{"localhost", "http://127.0.0.1:8080", "figlet", "alpha", "http://127.0.0.1:8080/function/figlet.alpha", "http://127.0.0.1:8080/async-function/figlet.alpha"},
		{"secure site", "https://example.com", "nodeinfo", "beta", "https://example.com/function/nodeinfo.beta", "https://example.com/async-function/nodeinfo.beta"},
		{"no namespace", "https://example.com:31112", "nodeinfo", "", "https://example.com:31112/function/nodeinfo", "https://example.com:31112/async-function/nodeinfo"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			url, asyncURL := getFunctionURLs(tc.gateway, tc.functionName, tc.functionNamespace)

			if url != tc.expectedURL || asyncURL != tc.expectedAsyncURL {
				t.Fatalf("incorrect URL(s), want: %q and %q, got: %q and %q", tc.expectedURL, tc.expectedAsyncURL, url, asyncURL)
			}
		})
	}
}
