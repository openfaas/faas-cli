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

func Test_parseBuildArgs_ValidParts(t *testing.T) {
	mapped, err := parseMap([]string{"k=v"}, "build-arg")

	if err != nil {
		t.Errorf("err was supposed to be nil but was: %s", err.Error())
		t.Fail()
	}

	if mapped["k"] != "v" {
		t.Errorf("value for 'k', want: %s got: %s", "v", mapped["k"])
		t.Fail()
	}
}

func Test_parseBuildArgs_NoSeparator(t *testing.T) {
	_, err := parseMap([]string{"kv"}, "build-arg")

	want := "label format is not correct, needs key=value"
	if err != nil && err.Error() != want {
		t.Errorf("Expected an error due to missing seperator")
		t.Fail()
	}
}

func Test_parseBuildArgs_EmptyKey(t *testing.T) {
	_, err := parseMap([]string{"=v"}, "build-arg")

	want := "empty build-arg name: [=v]"
	if err == nil {
		t.Errorf("Expected an error due to missing key")
		t.Fail()
	} else if err.Error() != want {
		t.Errorf("missing key error want: %s, got: %s", want, err.Error())
		t.Fail()
	}
}

func Test_parseBuildArgs_MultipleSeparators(t *testing.T) {
	mapped, err := parseMap([]string{"k=v=z"}, "build-arg")

	if err != nil {
		t.Errorf("Expected second separator to be included in value")
		t.Fail()
	}

	if mapped["k"] != "v=z" {
		t.Errorf("value for 'k', want: %s got: %s", "v=z", mapped["k"])
		t.Fail()
	}
}
