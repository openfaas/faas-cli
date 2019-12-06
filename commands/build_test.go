// Copyright (c) OpenFaaS Author(s) 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"testing"
)

func Test_build(t *testing.T) {
	templatePullLocalTemplateRepo(t)
	defer tearDownFetchTemplates(t)

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

func Test_preRunBuild_ParallelOverZero(t *testing.T) {
	buildCmd.ParseFlags([]string{"--parallel=0"})
	got := buildCmd.PreRunE(buildCmd, nil)

	if got == nil {
		t.Error("Parallel should have errored about being over zero")
		t.Fail()
		return
	}
	want := "the --parallel flag must be great than 0"
	if got.Error() != want {
		t.Errorf("parsing error, want %s, got %s", want, got.Error())
	}
}

func Test_parseBuildArgs_ValidParts(t *testing.T) {
	mapped, err := parseBuildArgs([]string{"k=v"})

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
	_, err := parseBuildArgs([]string{"kv"})

	want := "each build-arg must take the form key=value"
	if err != nil && err.Error() != want {
		t.Errorf("Expected an error due to missing seperator")
		t.Fail()
	}
}

func Test_parseBuildArgs_EmptyKey(t *testing.T) {
	_, err := parseBuildArgs([]string{"=v"})

	want := "build-arg must have a non-empty key"
	if err == nil {
		t.Errorf("Expected an error due to missing key")
		t.Fail()
	} else if err.Error() != want {
		t.Errorf("missing key error want: %s, got: %s", want, err.Error())
		t.Fail()
	}
}

func Test_parseBuildArgs_MultipleSeparators(t *testing.T) {
	mapped, err := parseBuildArgs([]string{"k=v=z"})

	if err != nil {
		t.Errorf("Expected second separator to be included in value")
		t.Fail()
	}

	if mapped["k"] != "v=z" {
		t.Errorf("value for 'k', want: %s got: %s", "v=z", mapped["k"])
		t.Fail()
	}
}
