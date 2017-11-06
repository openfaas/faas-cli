// Copyright (c) OpenFaaS Project 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"testing"
)

func Test_build(t *testing.T) {

	aTests := [][]string{
		{"build"},
		{"build", "--image=my_image"},
		{"build", "--image=my_image", "--handler=/path/to/fn/"},
	}

	for _, aTest := range aTests {
		faasCmd.SetArgs(aTest)
		err := faasCmd.Execute()
		if err == nil {
			t.Fatalf("No error found while testing \n%v", err)
		}
	}
}
