// Copyright (c) OpenFaaS project 2018. All rights reserved.
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
		{scenario: "Valid image username", name: "cli", image: "alexellis/faas-cli", isValid: true},
	}

	for _, testCase := range testCases {
		functions := map[string]stack.Function{
			"cli": stack.Function{
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
