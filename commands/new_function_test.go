// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"io/ioutil"
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

type WriteYAMLTest struct {
	title        string
	services     stack.Services
	expectedText string
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

var WriteYAMLTests = []WriteYAMLTest{
	{
		title: "writeYAML-test1",
		services: stack.Services{
			Functions: map[string]stack.Function{
				"test1": {
					Language: "testLang1",
					Handler:  "testHandler1",
					Image:    "testImage1",
				},
			},
			Provider: stack.Provider{
				Name:       "testProvider1",
				GatewayURL: "testGate1",
				Network:    "testNet1",
			},
		},
		expectedText: `provider:
  name: testProvider1
  gateway: testGate1
  network: testNet1
functions:
  test1:
    lang: testLang1
    handler: testHandler1
    image: testImage1
`,
	},
	{
		title: "writeYAML-test2",
		services: stack.Services{
			Functions: map[string]stack.Function{
				"test2": {
					Language: "testLang2",
					Handler:  "testHandler2",
					Image:    "testImage2",
				},
			},
		},
		expectedText: `functions:
  test2:
    lang: testLang2
    handler: testHandler2
    image: testImage2
`,
	},
	{
		title: "writeYAML-test3",
		services: stack.Services{
			Functions: map[string]stack.Function{
				"test3": {
					Name:     "testName3",
					Language: "testLang3",
					Handler:  "testHandler3",
					Image:    "testImage3",
					FProcess: "testFProcess3",
					Environment: map[string]string{
						"envKey1": "envVal1",
						"envKey2": "envVal2",
						"envKey3": "envVal3",
					},
					SkipBuild:       true,
					Constraints:     &[]string{"Constraint1", "Constraint2"},
					EnvironmentFile: []string{"EnvFile1", "EnvFile2", "EnvFile3"},
				},
			},
		},
		expectedText: `functions:
  test3:
    '-': testName3
    lang: testLang3
    handler: testHandler3
    image: testImage3
    fprocess: testFProcess3
    environment:
      envKey1: envVal1
      envKey2: envVal2
      envKey3: envVal3
    skip_build: true
    constraints:
    - Constraint1
    - Constraint2
    environment_file:
    - EnvFile1
    - EnvFile2
    - EnvFile3
`,
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
		testServices.Provider = stack.Provider{Name: "faas", GatewayURL: defaultGateway}
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

func runWriteYAMLTest(t *testing.T, wyt WriteYAMLTest) {
	testFile := "WriteYAMLTestFile.yml"

	defer func() {
		os.Remove(testFile)
	}()

	err := stack.WriteYAMLData(&wyt.services, testFile)
	if err != nil {
		t.Fatalf("Error running WriteYAMLData to file %s: %v", testFile, err)
	}

	byteData, err := ioutil.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Error reading from file %s: %v", testFile, err)
	}

	if string(byteData[:]) != wyt.expectedText {
		t.Fatalf("Test %s failed, mismatch between expected and actual YAML", wyt.title)
	}
}

func Test_newFunctionTests(t *testing.T) {
	// Reset parameters which may have been modified by other tests
	defer func() {
		yamlFile = ""
	}()

	// Change directory to testdata
	if err := os.Chdir("testdata/new_function"); err != nil {
		t.Fatalf("Error on cd to testdata dir: %v", err)
	}

	for _, test := range NewFunctionTests {
		t.Run(test.title, func(t *testing.T) {
			runNewFunctionTest(t, test)
		})
	}

	// Run WriteYAMLData tests
	for _, test := range WriteYAMLTests {
		t.Run(test.title, func(t *testing.T) {
			runWriteYAMLTest(t, test)
		})
	}
}
