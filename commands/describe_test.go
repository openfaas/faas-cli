package commands

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/openfaas/faas-cli/schema"
	"github.com/openfaas/faas-provider/types"
)

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

func TestDescribeOuput(t *testing.T) {
	spaces := regexp.MustCompile(`[ ]{2,}`)
	cases := []struct {
		name           string
		function       schema.FunctionDescription
		verbose        bool
		expectedOutput string
	}{
		{
			name: "minimal output, non-verbose",
			function: schema.FunctionDescription{
				FunctionStatus: types.FunctionStatus{
					Name:        "figlet",
					Image:       "openfaas/figlet:latest",
					Labels:      &map[string]string{},
					Annotations: &map[string]string{},
				},
				Status: "Ready",
			},
			verbose:        false,
			expectedOutput: "Name:\tfiglet\nStatus:\tReady\nReplicas:\t0\nAvailable Replicas: 0\nInvocations:\t0\nImage:\topenfaas/figlet:latest\nFunction Process:\t<default>\n",
		},
		{
			name: "minimal output, verbose",
			function: schema.FunctionDescription{
				FunctionStatus: types.FunctionStatus{
					Name:        "figlet",
					Image:       "openfaas/figlet:latest",
					Labels:      &map[string]string{},
					Annotations: &map[string]string{},
				},
				Status: "Ready",
			},
			verbose:        true,
			expectedOutput: "Name:\tfiglet\nStatus:\tReady\nReplicas:\t0\nAvailable Replicas: 0\nInvocations:\t0\nImage:\topenfaas/figlet:latest\nFunction Process:\t<default>\nURL:\t<none>\nAsync URL:\t<none>\nLabels:\t<none>\nAnnotations:\t<none>\nConstraints:\t<none>\nEnvironment:\t<none>\nSecrets:\t<none>\nRequests:\t<none>\nLimits:\t<none>\nNo usage information available\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var dst bytes.Buffer
			printFunctionDescription(&dst, tc.function, tc.verbose)
			result := spaces.ReplaceAllString(dst.String(), "\t")
			if result != tc.expectedOutput {
				t.Fatalf("incorrect output,\nwant: %q\nnorm: %q\ngot: %q", tc.expectedOutput, result, dst.String())
			}
		})
	}
}
