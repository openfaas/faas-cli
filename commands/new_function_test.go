// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"os"
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
- dockerfile
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
		title:       "lowercase-dockerfile",
		funcName:    "lowercase-dockerfile",
		funcLang:    "dockerfile",
		expectedMsg: SuccessMsg,
	},
	{
		title:       "uppercase-dockerfile",
		funcName:    "uppercase-dockerfile",
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
	execErr := faasCmd.Execute()

	if nft.expectedMsg == SuccessMsg {

		// Make sure that the folder and file was created:
		if _, err := os.Stat("./" + funcName); os.IsNotExist(err) {
			t.Fatalf("%s/ directory was not created", funcName)
		}

		// Check that the Dockerfile was created
		if funcLang == "Dockerfile" || funcLang == "dockerfile" {
			if _, err := os.Stat("./" + funcName + "/Dockerfile"); os.IsNotExist(err) {
				t.Fatalf("Dockerfile language should create a Dockerfile for you", funcName)
			}
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
		if found, err := regexp.MatchString(nft.expectedMsg, execErr.Error()); err != nil || !found {
			t.Fatalf("Output is not as expected: %s\n", execErr)
		}
	}

}

func Test_newFunctionTests(t *testing.T) {
	// Download templates
	templatePullLocalTemplateRepo(t)
	defer tearDownFetchTemplates(t)
	defer tearDownNewFunction(t)

	for _, testcase := range NewFunctionTests {
		t.Run(testcase.title, func(t *testing.T) {
			runNewFunctionTest(t, testcase)
		})
	}
}

func Test_newFunctionListCmds(t *testing.T) {
	// Download templates
	templatePullLocalTemplateRepo(t)
	defer tearDownFetchTemplates(t)
	defer tearDownNewFunction(t)

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
}

func Test_languageNotExists(t *testing.T) {
	// Download templates
	templatePullLocalTemplateRepo(t)
	defer tearDownFetchTemplates(t)
	defer tearDownNewFunction(t)

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
}

func tearDownNewFunction(t *testing.T) {
	// Remove existing archive file if it exists
	if _, err := os.Stat(".gitignore"); err == nil {
		err := os.Remove(".gitignore")
		if err != nil {
			t.Log(err)
		}
	}
}
