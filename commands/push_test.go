// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"testing"

	"github.com/openfaas/faas-cli/stack"
)

func Test_PushValidation(t *testing.T) {
	testCases := []struct {
		name     string
		scenario string
		image    string
		isValid  bool
	}{
		{scenario: "Valid image with username", name: "cli", image: "alexellis/faas-cli", isValid: true},
		{scenario: "Valid image with remote repo", name: "cli", image: "10.1.95.201:5000/faas-cli", isValid: true},
		{scenario: "Invalid image - missing prefix", name: "cli", image: "faas-cli", isValid: false},
	}

	for _, testCase := range testCases {
		functions := map[string]stack.Function{
			"cli": {
				Name:  testCase.name,
				Image: testCase.image,
			},
		}
		invalidImages := validateImages(functions)
		if len(invalidImages) > 0 && testCase.isValid == true {
			t.Logf("scenario: %s want %s to be valid, but was invalid", testCase.scenario, testCase.image)
			t.Fail()
		}
		if len(invalidImages) == 0 && testCase.isValid == false {
			t.Logf("scenario: %s want %s to be invalid, but was valid", testCase.scenario, testCase.image)
			t.Fail()
		}

	}
}
