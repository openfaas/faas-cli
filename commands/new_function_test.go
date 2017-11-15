// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/faas-cli/test"
)

const SuccessMsg = `(?m:Function created in folder)`
const InvalidYAMLMsg = `is not valid YAML`
const InvalidYAMLMap = `map is empty`
const ListOptionOutput = `Languages available as templates:
- csharp
- go
- go-armhf
- node
- node-arm64
- node-armhf
- python
- python-armhf
- python3
- ruby`

const LangNotExistsOutput = `(?m:is unavailable or not supported)`

type NewFunctionTest struct {
	title       string
	funcName    string
	funcLang    string
	expectedMsg string
}

var NewFunctionTests = []NewFunctionTest{
	{
		title:       "new_1",
		funcName:    "new-test-1",
		funcLang:    "ruby",
		expectedMsg: SuccessMsg,
	},
	{
		title:       "new_2",
		funcName:    "new-test-2",
		funcLang:    "dockerfile",
		expectedMsg: SuccessMsg,
	},
	{
		title:       "invalid_1",
		funcName:    "new-test-invalid-1",
		funcLang:    "dockerfilee",
		expectedMsg: LangNotExistsOutput,
	},
}

func runNewFunctionTest(t *testing.T, nft NewFunctionTest) {
	funcName := nft.funcName
	funcLang := nft.funcLang
	var funcYAML string
	funcYAML = funcName + ".yml"

	// Cleanup the created directory
	defer func() {
		os.RemoveAll(funcName)
		os.Remove(funcYAML)
	}()

	cmdParameters := []string{
		"new",
		funcName,
		"--lang=" + funcLang,
		"--gateway=" + defaultGateway,
	}

	faasCmd.SetArgs(cmdParameters)
	fmt.Println("Executing command")
	stdOut := faasCmd.Execute()

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
		testServices.Provider = stack.Provider{Name: "faas", GatewayURL: defaultGateway}
		if !reflect.DeepEqual(services.Provider, testServices.Provider) {
			t.Fatalf("YAML `provider` section was not created correctly for file %s: got %v", funcYAML, services.Provider)
		}

		testServices.Functions = make(map[string]stack.Function)
		testServices.Functions[funcName] = stack.Function{Language: funcLang, Image: funcName, Handler: "./" + funcName}
		if !reflect.DeepEqual(services.Functions[funcName], testServices.Functions[funcName]) {
			t.Fatalf("YAML `functions` section was not created correctly for file %s, got %v", funcYAML, services.Functions[funcName])
		}
	} else {
		// Validate new function output
		if found, err := regexp.MatchString(nft.expectedMsg, stdOut.Error()); err != nil || !found {
			t.Fatalf("Output is not as expected: %s\n", stdOut)
		}
	}

}

func Test_newFunctionTests(t *testing.T) {

	homeDir, _ := filepath.Abs(".")
	if err := os.Chdir("testdata/new_function"); err != nil {
		t.Fatalf("Error on cd to testdata dir: %v", err)
	}

	for _, test := range NewFunctionTests {
		t.Run(test.title, func(t *testing.T) {
			runNewFunctionTest(t, test)
		})
	}

	if err := os.Chdir(homeDir); err != nil {
		t.Fatalf("Error on cd back to commands/ directory: %v", err)
	}
}

func Test_newFunctionListCmds(t *testing.T) {

	homeDir, _ := filepath.Abs(".")
	if err := os.Chdir("testdata/new_function"); err != nil {
		t.Fatalf("Error on cd to testdata dir: %v", err)
	}

	cmdParameters := []string{
		"new",
		"--list",
	}

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs(cmdParameters)
		faasCmd.Execute()
	})

	// Validate new function output
	if !strings.HasPrefix(stdOut, ListOptionOutput) {
		t.Fatalf("Output is not as expected: %s\n", stdOut)
	}

	if err := os.Chdir(homeDir); err != nil {
		t.Fatalf("Error on cd back to commands/ directory: %v", err)
	}
}

func Test_languageNotExists(t *testing.T) {

	homeDir, _ := filepath.Abs(".")
	if err := os.Chdir("testdata/new_function"); err != nil {
		t.Fatalf("Error on cd to testdata dir: %v", err)
	}

	// Attempt to create a function with a non-existing language
	cmdParameters := []string{
		"new",
		"sampleName",
		"--lang=bash",
		"--gateway=" + defaultGateway,
		"--list=false",
	}

	faasCmd.SetArgs(cmdParameters)
	stdOut := faasCmd.Execute().Error()

	// Validate new function output
	if found, err := regexp.MatchString(LangNotExistsOutput, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected: %s\n", stdOut)
	}

	if err := os.Chdir(homeDir); err != nil {
		t.Fatalf("Error on cd back to commands/ directory: %v", err)
	}
}
