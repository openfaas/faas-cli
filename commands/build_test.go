// Copyright (c) OpenFaaS Project 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"testing"
)

func Test_build(t *testing.T) {

	aTests := [][]string{
		{},
		{"--image=my_image"},
		{"--image=my_image", "--handler=/path/to/fn/"},
	}

	for _, aTest := range aTests {
		buildCmd := newBuildCmd()
		buildCmd.SetArgs(aTest)
		err := buildCmd.Execute()
		if err == nil {
			t.Fatalf("No error found while testing \n%v", err)
		}
	}
}
