// Copyright (c) OpenFaaS Project 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"os"
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

func Test_validateBuildOption(t *testing.T) {

	buildOptions := []struct {
		buildOption           string
		expectedBuildArgValue string
	}{
		{
			buildOption:           "dev",
			expectedBuildArgValue: "ADDITIONAL_PACKAGE=make automake gcc g++ subversion python3-dev musl-dev libffi-dev",
		},

		{
			buildOption:           "undefined",
			expectedBuildArgValue: "",
		},
	}

	os.MkdirAll("template/python3", os.ModePerm)
	python3_template_yml, err := os.Create("template/python3/template.yml")
	if err != nil {
		t.Errorf("Error creating template/python3/template.yml file")
	}

	_, err = python3_template_yml.WriteString("language: python3\n" +
		"fprocess: python3 index.py\n" +
		"build_options: \n" +
		"  - name: dev\n" +
		"    packages: \n" +
		"      - make\n" +
		"      - automake\n" +
		"      - gcc\n" +
		"      - g++\n" +
		"      - subversion\n" +
		"      - python3-dev\n" +
		"      - musl-dev\n" +
		"      - libffi-dev\n")

	if err != nil {
		t.Errorf("Error writing to template/python3/template.yml file")
	}

	os.MkdirAll("template/unsupported", os.ModePerm)
	unsupported_template_yml, err := os.Create("template/unsupported/template.yml")
	if err != nil {
		t.Errorf("Error creating template/unsupported/template.yml file")
	}

	_, err = unsupported_template_yml.WriteString("language: python3\n" +
		"fprocess: python3 index.py\n")

	if err != nil {
		t.Errorf("Error writing to template/pythunsupportedon3/template.yml file")
	}

	for _, test := range buildOptions {
		t.Run(test.buildOption, func(t *testing.T) {
			res, _, _ := validateBuildOption(test.buildOption, "python3")
			_, isValid, _ := validateBuildOption(test.buildOption, "unsupported")

			if res != test.expectedBuildArgValue {
				t.Errorf("validateBuildOption failed for build-option %s. Expected to return %s, but returned %s",
					test.buildOption, test.expectedBuildArgValue, res)
			}

			if isValid && test.buildOption == "dev" {
				t.Errorf("validateBuildOption failed for build-option %s and unsupported language. Expected validation to fail, but it was successful",
					test.buildOption)
			}
		})
	}

	os.RemoveAll("template")
}
