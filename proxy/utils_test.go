package proxy

import (
	"testing"
)

func Test_createSystemEndpoint(t *testing.T) {
	tests := []struct {
		title            string
		gateway          string
		namespace        string
		exptectedErr     bool
		expectedEndpoint string
	}{
		{
			title:            "Namespace is set",
			gateway:          "http://127.0.0.1:8080",
			namespace:        "production",
			exptectedErr:     false,
			expectedEndpoint: "http://127.0.0.1:8080/system/functions?namespace=production",
		},
		{
			title:            "Namespace is not set",
			gateway:          "http://127.0.0.1:8080",
			namespace:        "",
			exptectedErr:     false,
			expectedEndpoint: "http://127.0.0.1:8080/system/functions",
		},
		{
			title:            "Bad gateway formatting",
			gateway:          "127.0.0.1:8080",
			namespace:        "production",
			exptectedErr:     true,
			expectedEndpoint: "",
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			output, err := createSystemEndpoint(test.gateway, test.namespace)
			if output != test.expectedEndpoint {
				t.Errorf("Expected format of the gateway: %s found: %s",
					test.expectedEndpoint,
					output)
			}
			if err != nil && test.exptectedErr == false {
				t.Errorf("Expected nil error, got: %s",
					err.Error())
			}
		})
	}
}

func Test_createFunctionEndpoint(t *testing.T) {
	tests := []struct {
		title            string
		gateway          string
		namespace        string
		functionName     string
		exptectedErr     bool
		expectedEndpoint string
	}{
		{
			title:            "Namespace is set",
			gateway:          "http://127.0.0.1:8080",
			namespace:        "production",
			functionName:     "cows",
			exptectedErr:     false,
			expectedEndpoint: "http://127.0.0.1:8080/system/function/cows?namespace=production",
		},
		{
			title:            "Namespace is not set",
			gateway:          "http://127.0.0.1:8080",
			functionName:     "cows",
			namespace:        "",
			exptectedErr:     false,
			expectedEndpoint: "http://127.0.0.1:8080/system/function/cows",
		},
		{
			title:            "Bad gateway formatting",
			gateway:          "127.0.0.1:8080",
			namespace:        "production",
			exptectedErr:     true,
			expectedEndpoint: "",
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			output, err := createFunctionEndpoint(test.gateway, test.functionName, test.namespace)
			if output != test.expectedEndpoint {
				t.Errorf("Expected format of the gateway: %s found: %s",
					test.expectedEndpoint,
					output)
			}
			if err != nil && test.exptectedErr == false {
				t.Errorf("Expected nil error, got: %s",
					err.Error())
			}
		})
	}
}
