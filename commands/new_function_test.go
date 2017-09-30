// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/faas-cli/test"
)

const SuccessMsg = `(?m:Function created in folder)`
const InvalidYAMLMsg = `is not valid YAML`
const InvalidYAMLMap = `map is empty`

type NewFunctionTest struct {
	title       string
	funcName    string
	funcLang    string
	file        string
	expectedMsg string
}

var NewFunctionTests = []NewFunctionTest{
	{
		title:       "new_1",
		funcName:    "new-test-1",
		funcLang:    "ruby",
		file:        "",
		expectedMsg: SuccessMsg,
	},
	{
		title:       "new_2",
		funcName:    "new-test-2",
		funcLang:    "dockerfile",
		file:        "",
		expectedMsg: SuccessMsg,
	},
	{
		title:       "new_append_1",
		funcName:    "new-test-append-1",
		funcLang:    "python",
		file:        "new_test.yml",
		expectedMsg: SuccessMsg,
	},
	{
		title:       "new_append_2",
		funcName:    "new-test-append-2",
		funcLang:    "Dockerfile",
		file:        "new_test.yml",
		expectedMsg: SuccessMsg,
	},
	{
		title:       "new_append_invalid_1",
		funcName:    "new-test-append-invalid-1",
		funcLang:    "Dockerfile",
		file:        "invalid1.yml",
		expectedMsg: InvalidYAMLMsg,
	},
	{
		title:       "new_append_invalid_2",
		funcName:    "new-test-append-invalid-2",
		funcLang:    "csharp",
		file:        "invalid2.yml",
		expectedMsg: InvalidYAMLMap,
	},
	{
		title:       "new_append_invalid_3",
		funcName:    "new-test-append-invalid-3",
		funcLang:    "python3",
		file:        "invalid3.yml",
		expectedMsg: InvalidYAMLMsg,
	},
}

func runNewFunctionTest(t *testing.T, nft NewFunctionTest) {
	funcName := nft.funcName
	funcLang := nft.funcLang
	var funcYAML string

	originalYAMLFile := "new_test.orig_yaml"
	if len(nft.file) == 0 {
		funcYAML = funcName + ".yml"
	} else {
		funcYAML = nft.file
		test.Copy(funcYAML, originalYAMLFile)
	}

	// Cleanup the created directory
	defer func() {
		os.RemoveAll(funcName)
		if len(nft.file) == 0 {
			os.Remove(funcYAML)
		} else {
			test.Copy(originalYAMLFile, funcYAML)
			os.Remove(originalYAMLFile)
		}
	}()

	cmdParameters := []string{
		"new",
		"--name=" + funcName,
		"--lang=" + funcLang,
	}
	if len(nft.file) > 0 {
		cmdParameters = append(cmdParameters, "--yaml="+nft.file)
	}

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs(cmdParameters)
		faasCmd.Execute()
	})

	// Validate new function output
	if found, err := regexp.MatchString(nft.expectedMsg, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}

	if nft.expectedMsg == SuccessMsg {

		// Make sure that the folder and file was created:
		if _, err := os.Stat("./" + funcName); os.IsNotExist(err) {
			t.Fatalf("%s/ directory was not created", funcName)
		}

		if _, err := os.Stat(funcYAML); os.IsNotExist(err) {
			t.Fatalf("\"%s\" yaml file was not created", funcYAML)
		}

		// Make sure that the information in the YAML file is correct:
		parsedServices, err := stack.ParseYAMLFile(funcYAML, "", "")
		if err != nil {
			t.Fatalf("Couldn't open modified YAML file \"%s\" due to error: %v", funcYAML, err)
		}
		services := *parsedServices

		var testServices stack.Services
		testServices.Provider = stack.Provider{Name: "faas", GatewayURL: "http://localhost:8080"}
		if !reflect.DeepEqual(services.Provider, testServices.Provider) {
			t.Fatalf("YAML `provider` section was not created correctly for file %s: got %v", funcYAML, services.Provider)
		}

		testServices.Functions = make(map[string]stack.Function)
		testServices.Functions[funcName] = stack.Function{Language: funcLang, Image: funcName, Handler: "./" + funcName}
		if !reflect.DeepEqual(services.Functions[funcName], testServices.Functions[funcName]) {
			t.Fatalf("YAML `functions` section was not created correctly for file %s, got %v", funcYAML, services.Functions[funcName])
		}
	}

}

func Test_newFunctionTests(t *testing.T) {
	defer func() {
		os.Remove("template")
	}()

	// Change directory to testdata
	if err := os.Chdir("testdata/new_function"); err != nil {
		t.Fatalf("Error on cd to testdata dir: %v", err)
	}

	// Symlink template directory at root to current directory to avoid re-downloading templates
	os.Symlink("../../../template", "template")

	for _, test := range NewFunctionTests {
		t.Run(test.title, func(t *testing.T) {
			runNewFunctionTest(t, test)
		})
	}
}
