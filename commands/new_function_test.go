// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/faas-cli/test"
)

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
		expectedMsg: `(?m:Function created in folder)`,
	},
	{
		title:       "new_2",
		funcName:    "new-test-2",
		funcLang:    "dockerfile",
		file:        "",
		expectedMsg: `(?m:Function created in folder)`,
	},
}

var AppendFunctionTests = []NewFunctionTest{
	{
		title:       "new_append_1",
		funcName:    "new-test-append-1",
		funcLang:    "python",
		file:        "append1.yml",
		expectedMsg: `(?m:Function created in folder)`,
	},
	{
		title:       "new_append_1_dockerfile",
		funcName:    "new-test-append-1-dockerfile",
		funcLang:    "Dockerfile",
		file:        "append1.yml",
		expectedMsg: `(?m:Function created in folder)`,
	},
	{
		title:       "new_append_2",
		funcName:    "new-test-append-2",
		funcLang:    "csharp",
		file:        "append2.yml",
		expectedMsg: `(?m:Function created in folder)`,
	},
	{
		title:       "new_append_3",
		funcName:    "new-test-append-3",
		funcLang:    "python",
		file:        "append3.yml",
		expectedMsg: `(?m:Function created in folder)`,
	},
	{
		title:       "new_append_4",
		funcName:    "new-test-append-4",
		funcLang:    "python",
		file:        "append4.yml",
		expectedMsg: `(?m:Function created in folder)`,
	},
}

var InvalidNewFunctionTests = []NewFunctionTest{
	{
		title:       "new_append_invalid_1",
		funcName:    "new-test-append-invalid-1",
		funcLang:    "Dockerfile",
		file:        "invalid1.yml",
		expectedMsg: `is not valid YAML`,
	},
	{
		title:       "new_append_invalid_2",
		funcName:    "new-test-append-invalid-2",
		funcLang:    "python3",
		file:        "invalid2.yml",
		expectedMsg: `is not valid YAML`,
	},
	{
		title:       "new_append_invalid_3",
		funcName:    "new-test-append-invalid-3",
		funcLang:    "python",
		file:        "invalid3.yml",
		expectedMsg: `is not valid YAML`,
	},
	{
		title:       "new_append_invalid_4",
		funcName:    "new-test-append-invalid-4",
		funcLang:    "python",
		file:        "invalid4.yml",
		expectedMsg: `is not valid YAML`,
	},
	{
		title:       "invalid_4",
		funcName:    "new-test-invalid-1",
		funcLang:    "dockerfilee",
		file:        "",
		expectedMsg: `(?m:is unavailable or not supported)`,
	},
}

func parseYAMLFileForNewTest(t *testing.T, fileName string) *stack.Services {
	parsedService, err := stack.ParseYAMLFile(fileName, "", "")
	if err != nil {
		t.Fatalf("Error encountered in file \"%s\": %v", fileName, err)
	}
	return parsedService
}

func copyFileForNewTest(t *testing.T, srcFileName string, dstFileName string) {
	err := test.CopyFile(srcFileName, dstFileName)
	if err != nil {
		t.Fatalf("Error encountered while saving test file contents: %v", err)
	}
}

func runNewFunctionTest(t *testing.T, nft NewFunctionTest) {
	successMsg := `(?m:Function created in folder)`
	funcName := nft.funcName
	funcLang := nft.funcLang

	cmdParameters := []string{
		"new",
		funcName,
		"--lang=" + funcLang,
	}

	funcYAML := funcName + ".yml"

	// After the test is complete, reset the test YAML file (that has been
	// appended to), and clean up the newly created directory.
	defer func() {
		os.RemoveAll(funcName)
		os.Remove(funcYAML)
	}()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs(cmdParameters)
		faasCmd.Execute()
	})

	// Validate new function output
	if found, err := regexp.MatchString(nft.expectedMsg, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}

	if nft.expectedMsg == successMsg {

		// Make sure that the folder and file was created:
		if _, err := os.Stat("./" + funcName); os.IsNotExist(err) {
			t.Fatalf("%s/ directory was not created", funcName)
		}

		if _, err := os.Stat(funcYAML); os.IsNotExist(err) {
			t.Fatalf("\"%s\" yaml file was not created", funcYAML)
		}

		// Make sure that the information in the YAML file is correct:
		parsedServices := parseYAMLFileForNewTest(t, funcYAML)
		services := *parsedServices

		testProvider := stack.Provider{Name: "faas", GatewayURL: defaultGateway}
		if !reflect.DeepEqual(services.Provider, testProvider) {
			t.Fatalf("YAML `provider` section was not created correctly for file %s: got %v", funcYAML, services.Provider)
		}

		testFunction := stack.Function{Language: funcLang, Image: funcName, Handler: "./" + funcName}
		if !reflect.DeepEqual(services.Functions[funcName], testFunction) {
			t.Fatalf("YAML `functions` section was not created correctly for file %s, got %v", funcYAML, services.Functions[funcName])
		}
	}
}

func runAppendFunctionTest(t *testing.T, nft NewFunctionTest) {

	funcName := nft.funcName
	funcLang := nft.funcLang
	funcYAML := nft.file
	var originalYAMLFile string
	var originalServices stack.Services

	cmdParameters := []string{
		"new",
		funcName,
		"--lang=" + funcLang,
	}

	// Copy the YAML file to a separate '.orig' file, so that we can re-set
	// the test YAML file back to its original state after the test completes
	originalYAMLFile = funcYAML + ".orig"
	copyFileForNewTest(t, funcYAML, originalYAMLFile)

	originalServiceYAMLData := parseYAMLFileForNewTest(t, funcYAML)
	originalServices = *originalServiceYAMLData

	cmdParameters = append(cmdParameters, "--yaml="+funcYAML)

	// Check if the file to append to actually exists
	if _, err := os.Stat(funcYAML); os.IsNotExist(err) {
		t.Fatalf("\"%s\" yaml file does not exist", funcYAML)
	}

	// After the test is complete, reset the test YAML file (that has been
	// appended to), and clean up the newly created directory.
	defer func() {
		os.RemoveAll(funcName)
		copyFileForNewTest(t, originalYAMLFile, funcYAML)
		os.Remove(originalYAMLFile)
	}()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs(cmdParameters)
		faasCmd.Execute()
	})

	// Validate new function output
	if found, err := regexp.MatchString(nft.expectedMsg, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}

	// Make sure that the folder and file was created:
	if _, err := os.Stat("./" + funcName); os.IsNotExist(err) {
		t.Fatalf("%s/ directory was not created", funcName)
	}

	// Make sure that the file still exists
	if _, err := os.Stat(funcYAML); os.IsNotExist(err) {
		t.Fatalf("\"%s\" yaml file no longer exists", funcYAML)
	}

	// Make sure that the information in the YAML file is correct:
	parsedServices := parseYAMLFileForNewTest(t, funcYAML)
	services := *parsedServices

	testProvider := stack.Provider{Name: "faas", GatewayURL: defaultGateway}
	if !reflect.DeepEqual(services.Provider, testProvider) {
		t.Fatalf("YAML `provider` section was not created correctly for file %s: got %v", funcYAML, services.Provider)
	}

	for key := range services.Functions {
		if _, ok := originalServices.Functions[key]; ok {
			if !reflect.DeepEqual(services.Functions[key], originalServices.Functions[key]) {
				t.Fatalf("YAML `functions` section was not created correctly for file %s, want: %+v, got %+v", funcYAML, originalServices.Functions[key], services.Functions[key])
			}

		} else {
			testFunction := stack.Function{Language: funcLang, Image: funcName, Handler: "./" + funcName}
			if !reflect.DeepEqual(services.Functions[key], testFunction) {
				t.Fatalf("YAML `functions` section was not created correctly for file %s, want: %+v, got %+v", funcYAML, testFunction, services.Functions[key])
			}
		}
	}
}

func runInvalidNewFuncTest(t *testing.T, nft NewFunctionTest) {
	yamlFile = ""

	funcName := nft.funcName
	funcLang := nft.funcLang
	funcYAML := nft.file
	originalYAMLFile := funcYAML + ".orig"

	cmdParameters := []string{
		"new",
		funcName,
		"--lang=" + funcLang,
	}

	if nft.file != "" {
		// Preserve original file state
		copyFileForNewTest(t, funcYAML, originalYAMLFile)

		cmdParameters = append(cmdParameters, "--yaml="+funcYAML)

		// After the test is complete, reset the test YAML file (that has been
		// appended to), and clean up the newly created directory.
		defer func() {
			os.RemoveAll(funcName)
			copyFileForNewTest(t, originalYAMLFile, funcYAML)
			os.Remove(originalYAMLFile)
		}()
	}

	faasCmd.SilenceErrors = false
	faasCmd.SetArgs(cmdParameters)

	var errMsg string
	if err := faasCmd.Execute(); err != nil {
		errMsg = err.Error()
	}

	// Validate new function output
	if found, err := regexp.MatchString(nft.expectedMsg, errMsg); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", errMsg)
	}

}

func Test_newFunctionTests(t *testing.T) {

	homeDir, _ := filepath.Abs(".")
	// Change directory to testdata
	if err := os.Chdir("testdata/new_function"); err != nil {
		t.Fatalf("Error on cd to testdata dir: %v", err)
	}

	for _, test := range NewFunctionTests {
		t.Run(test.title, func(t *testing.T) {
			runNewFunctionTest(t, test)
		})
	}

	for _, test := range AppendFunctionTests {
		t.Run(test.title, func(t *testing.T) {
			runAppendFunctionTest(t, test)
		})
	}

	for _, test := range InvalidNewFunctionTests {
		t.Run(test.title, func(t *testing.T) {
			runInvalidNewFuncTest(t, test)
		})
	}

	if err := os.Chdir(homeDir); err != nil {
		t.Fatalf("Error on cd back to commands/ directory: %v", err)
	}
}

func Test_newFunctionListCmds(t *testing.T) {
	listOptionOutput := `Languages available as templates:
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
	if !strings.HasPrefix(stdOut, listOptionOutput) {
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
	if found, err := regexp.MatchString(`(?m:is unavailable or not supported)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected: %s\n", stdOut)
	}

	if err := os.Chdir(homeDir); err != nil {
		t.Fatalf("Error on cd back to commands/ directory: %v", err)
	}
}
