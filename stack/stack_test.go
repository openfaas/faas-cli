// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package stack

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

const TestData_1 string = `provider:
  name: faas
  gateway: http://localhost:8080
  network: "func_functions"

functions:
  url-ping:
    lang: python
    handler: ./sample/url-ping
    image: alexellis/faas-url-ping

  nodejs-echo:
    lang: node
    handler: ./sample/nodejs-echo
    image: alexellis/faas-nodejs-echo

  imagemagick:
    lang: Dockerfile
    handler: ./sample/imagemagick
    image: functions/resizer
    fprocess: "convert - -resize 50% fd:1"

  ruby-echo:
    lang: ruby
    handler: ./sample/ruby-echo
    image: alexellis/ruby-echo

  abcd-eeee:
    lang: node
    handler: ./sample/abcd-eeee
    image: stuff2/stuff23423
`

const TestData_2 string = `provider:
  name: faas
  gateway: http://localhost:8080
  network: "func_functions"

`

const noMatchesErrorMsg string = "No functions matching --filter/--regex were found in the YAML file"
const invalidRegexErrorMsg string = "error parsing regexp"

var ParseYAMLTests_Regex = []struct {
	title         string
	searchTerm    string
	functions     []string
	file          string
	expectedError string
}{
	{
		title:         "Regex search for functions only containing 'node'",
		searchTerm:    "node",
		functions:     []string{"nodejs-echo"},
		file:          TestData_1,
		expectedError: "",
	},
	{
		title:         "Regex search for functions only containing 'echo'",
		searchTerm:    "echo",
		functions:     []string{"nodejs-echo", "ruby-echo"},
		file:          TestData_1,
		expectedError: "",
	},
	{
		title:         "Regex search for functions only containing '.+-.+'",
		searchTerm:    ".+-.+",
		functions:     []string{"abcd-eeee", "nodejs-echo", "ruby-echo", "url-ping"},
		file:          TestData_1,
		expectedError: "",
	},
	{
		title:         "Regex search for all functions: '.*'",
		searchTerm:    ".*",
		functions:     []string{"abcd-eeee", "imagemagick", "nodejs-echo", "ruby-echo", "url-ping"},
		file:          TestData_1,
		expectedError: "",
	},
	{
		title:         "Regex search for no functions: '----'",
		searchTerm:    "----",
		functions:     []string{},
		file:          TestData_1,
		expectedError: noMatchesErrorMsg,
	},
	{
		title:         "Regex search for functions without dashes: '^[^-]+$'",
		searchTerm:    "^[^-]+$",
		functions:     []string{"imagemagick"},
		file:          TestData_1,
		expectedError: "",
	},
	{
		title:         "Regex search for functions with 8 characters: '^.{8}$'",
		searchTerm:    "^.{8}$",
		functions:     []string{"url-ping"},
		file:          TestData_1,
		expectedError: "",
	},
	{
		title:         "Regex search for function with repeated 'e': 'e{2}'",
		searchTerm:    "e{2}",
		functions:     []string{"abcd-eeee"},
		file:          TestData_1,
		expectedError: "",
	},
	{
		title:         "Regex empty search term: ''",
		searchTerm:    "",
		functions:     []string{"abcd-eeee", "imagemagick", "nodejs-echo", "ruby-echo", "url-ping"},
		file:          TestData_1,
		expectedError: "",
	},
	{
		title:         "Regex invalid regex 1: '['",
		searchTerm:    "[",
		functions:     []string{},
		file:          TestData_1,
		expectedError: invalidRegexErrorMsg,
	},
	{
		title:         "Regex invalid regex 2: '*'",
		searchTerm:    "*",
		functions:     []string{},
		file:          TestData_1,
		expectedError: invalidRegexErrorMsg,
	},
	{
		title:         "Regex invalid regex 3: '(\\w)\\1'",
		searchTerm:    `(\w)\1`,
		functions:     []string{},
		file:          TestData_1,
		expectedError: invalidRegexErrorMsg,
	},
	{
		title:         "Regex that finds no matches: 'RANDOMREGEX'",
		searchTerm:    "RANDOMREGEX",
		functions:     []string{},
		file:          TestData_1,
		expectedError: noMatchesErrorMsg,
	},
	{
		title:         "Regex empty search term in empty YAML file: ",
		searchTerm:    "",
		functions:     []string{},
		file:          TestData_2,
		expectedError: "",
	},
}

var ParseYAMLTests_Filter = []struct {
	title         string
	searchTerm    string
	functions     []string
	file          string
	expectedError string
}{
	{
		title:         "Wildcard search for functions ending with 'echo'",
		searchTerm:    "*echo",
		functions:     []string{"nodejs-echo", "ruby-echo"},
		file:          TestData_1,
		expectedError: "",
	},
	{
		title:         "Wildcard search for functions with a - in between two words: '*-*'",
		searchTerm:    "*-*",
		functions:     []string{"abcd-eeee", "nodejs-echo", "ruby-echo", "url-ping"},
		file:          TestData_1,
		expectedError: "",
	},
	{
		title:         "Wildcard search for specific function: 'imagemagick'",
		searchTerm:    "imagemagick",
		functions:     []string{"imagemagick"},
		file:          TestData_1,
		expectedError: "",
	},
	{
		title:         "Wildcard search for all functions: '*'",
		searchTerm:    "*",
		functions:     []string{"abcd-eeee", "imagemagick", "nodejs-echo", "ruby-echo", "url-ping"},
		file:          TestData_1,
		expectedError: "",
	},
	{
		title:         "Wildcard empty search term: ''",
		searchTerm:    "",
		functions:     []string{"abcd-eeee", "imagemagick", "nodejs-echo", "ruby-echo", "url-ping"},
		file:          TestData_1,
		expectedError: "",
	},
	{
		title:         "Wildcard multiple wildcard characters: '**'",
		searchTerm:    "**",
		functions:     []string{"abcd-eeee", "imagemagick", "nodejs-echo", "ruby-echo", "url-ping"},
		file:          TestData_1,
		expectedError: "",
	},
	{
		title:         "Wildcard that finds no matches: 'RANDOMTEXT'",
		searchTerm:    "RANDOMTEXT",
		functions:     []string{},
		file:          TestData_1,
		expectedError: noMatchesErrorMsg,
	},
	{
		title:         "Wildcard empty search term in empty YAML file: ''",
		searchTerm:    "",
		functions:     []string{},
		file:          TestData_2,
		expectedError: "",
	},
}

func Test_ParseYAMLDataRegex(t *testing.T) {

	for _, test := range ParseYAMLTests_Regex {
		t.Run(test.title, func(t *testing.T) {

			parsedYAML, err := ParseYAMLData([]byte(test.file), test.searchTerm, "")

			if len(test.expectedError) > 0 {
				if err == nil {
					t.Errorf("Test_ParseYAMLDataRegex test [%s] test failed, expected error not thrown", test.title)
				}

				if !strings.Contains(err.Error(), test.expectedError) {
					t.Errorf("Test_ParseYAMLDataRegex test [%s] test failed, expected error message of '%s', got '%v'", test.title, test.expectedError, err)
				}

			} else {

				if err != nil {
					t.Errorf("Test_ParseYAMLDataRegex test [%s] test failed, unexpected error thrown: %v", test.title, err)
					return
				}

				keys := reflect.ValueOf(parsedYAML.Functions).MapKeys()
				strkeys := make([]string, len(keys))

				for i := 0; i < len(keys); i++ {
					strkeys[i] = keys[i].String()
				}

				sort.Strings(strkeys)
				t.Log(strkeys)

				if !reflect.DeepEqual(strkeys, test.functions) {
					t.Errorf("Test_ParseYAMLDataRegex test [%s] test failed, does not match expected result;\n  parsedYAML:   [%v]\n  expected: [%v]",
						test.title,
						strkeys,
						test.functions,
					)
				}
			}
		})
	}
}

func Test_ParseYAMLDataFilter(t *testing.T) {

	for _, test := range ParseYAMLTests_Filter {
		t.Run(test.title, func(t *testing.T) {

			parsedYAML, err := ParseYAMLData([]byte(test.file), "", test.searchTerm)

			if len(test.expectedError) > 0 {

				if err == nil {
					t.Errorf("Test_ParseYAMLDataFilter test [%s] test failed, expected error not thrown", test.title)
				}

				if !strings.Contains(err.Error(), test.expectedError) {
					t.Errorf("Test_ParseYAMLDataFilter test [%s] test failed, expected error message of '%s', got '%v'", test.title, test.expectedError, err)
				}

			} else {

				if err != nil {
					t.Errorf("Test_ParseYAMLDataFilter test [%s] test failed, unexpected error thrown: %v", test.title, err)
					return
				}

				keys := reflect.ValueOf(parsedYAML.Functions).MapKeys()
				strkeys := make([]string, len(keys))

				for i := 0; i < len(keys); i++ {
					strkeys[i] = keys[i].String()
				}

				sort.Strings(strkeys)
				t.Log(strkeys)

				if !reflect.DeepEqual(strkeys, test.functions) {
					t.Errorf("Test_ParseYAMLDataFilter test [%s] failed, does not match expected result;\n  parsedYAML:   [%v]\n  expected: [%v]",
						test.title,
						strkeys,
						test.functions,
					)
				}
			}
		})
	}
}

func Test_ParseYAMLDataFilterAndRegex(t *testing.T) {
	_, err := ParseYAMLData([]byte(TestData_1), ".*", "*")
	if err == nil {
		t.Errorf("Test_ParseYAMLDataFilterAndRegex test failed, expected error not thrown")
	}
}
