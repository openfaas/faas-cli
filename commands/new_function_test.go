// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"io/ioutil"
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
- dockerfile
- ruby`

const LangNotExistsOutput = `(?m:is unavailable or not supported)`
const FunctionExistsOutput = `(Function (.+)? already exists in (.+)? file)`

type NewFunctionTest struct {
	title         string
	prefix        string
	funcName      string
	funcLang      string
	stackFile     string
	expectedImage string
	expectedMsg   string
}

var NewFunctionTests = []NewFunctionTest{
	{
		title:         "new_1",
		funcName:      "new-test-1",
		funcLang:      "ruby",
		expectedImage: "new-test-1",
		expectedMsg:   SuccessMsg,
	},
	{
		title:         "lowercase-dockerfile",
		funcName:      "lowercase-dockerfile",
		funcLang:      "dockerfile",
		expectedImage: "lowercase-dockerfile",
		expectedMsg:   SuccessMsg,
	},
	{
		title:         "uppercase-dockerfile",
		funcName:      "uppercase-dockerfile",
		funcLang:      "dockerfile",
		expectedImage: "uppercase-dockerfile",
		expectedMsg:   SuccessMsg,
	},
	{
		title:         "func-with-prefix",
		funcName:      "func-with-prefix",
		prefix:        " username ",
		funcLang:      "dockerfile",
		expectedImage: "username/func-with-prefix",
		expectedMsg:   SuccessMsg,
	},
	{
		title:         "func-with-whitespace-only-prefix",
		funcName:      "func-with-whitespace-only-prefix",
		prefix:        " ",
		funcLang:      "dockerfile",
		expectedImage: "func-with-whitespace-only-prefix",
		expectedMsg:   SuccessMsg,
	},
	{
		title:       "invalid_1",
		funcName:    "new-test-invalid-1",
		funcLang:    "dockerfilee",
		expectedMsg: LangNotExistsOutput,
	},
	{
		title:         "new_with_stack",
		funcName:      "new_with_stack",
		funcLang:      "ruby",
		expectedImage: "new_with_stack",
		stackFile:     "new_with_stack_provided.yml",
		expectedMsg:   SuccessMsg,
	},
}

func runNewFunctionTest(t *testing.T, nft NewFunctionTest) {
	funcName := nft.funcName
	funcLang := nft.funcLang
	imagePrefix := nft.prefix
	stackFile := nft.stackFile
	var funcYAML string

	// Cleanup the created directory
	defer func() {
		os.RemoveAll(funcName)
		os.Remove(funcYAML)
	}()

	cmdParameters := []string{}
	if stackFile != "" {
		cmdParameters = []string{
			"new",
			funcName,
			"--lang=" + funcLang,
			"--gateway=" + defaultGateway,
			"--prefix=" + imagePrefix,
			"--stack=" + stackFile,
		}
		funcYAML = stackFile
	} else {
		cmdParameters = []string{
			"new",
			funcName,
			"--lang=" + funcLang,
			"--gateway=" + defaultGateway,
			"--prefix=" + imagePrefix,
		}
		funcYAML = funcName + ".yml"
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
				t.Fatalf("Dockerfile language should create a Dockerfile for you as ./%s/Dockerfile", funcName)
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
		testServices.Functions[funcName] = stack.Function{Language: funcLang, Image: nft.expectedImage, Handler: "./" + funcName}
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

func Test_duplicateFunctionName(t *testing.T) {
	templatePullLocalTemplateRepo(t)
	defer tearDownFetchTemplates(t)
	defer tearDownNewFunction(t)

	const functionName = "sampleFunc"
	const functionLang = "ruby"

	defer func() {
		if _, err := os.Stat(functionName + ".yml"); err == nil {
			err := os.Remove(functionName + ".yml")
			if err != nil {
				t.Log(err)
			}
		}
	}()

	// Create a yml file with the same function name and language
	writeFunctionYmlFile(functionName, functionLang, t)

	appendParameters := []string{
		"new",
		functionName,
		"--lang=" + functionLang,
		"--stack=" + functionName + ".yml",
	}

	faasCmd.SetArgs(appendParameters)
	stdOut := faasCmd.Execute().Error()

	if found, err := regexp.MatchString(FunctionExistsOutput, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected: %s\n", stdOut)
	}
}

func writeFunctionYmlFile(functionName string, lang string, t *testing.T) {
	var stackYaml string

	stackYaml +=
		`provider:
  name: faas
  gateway: ` + defaultGateway + `

functions:
`

	stackYaml +=
		`  ` + functionName + `:
    lang: ` + lang + `
    handler: ./` + functionName + `
    image: ` + functionName + `
`

	stackWriteErr := ioutil.WriteFile("./"+functionName+".yml", []byte(stackYaml), 0600)
	if stackWriteErr != nil {
		t.Log(fmt.Errorf("error writing stack file %s", stackWriteErr))
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
