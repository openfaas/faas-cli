// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/openfaas/faas-cli/test"
	"github.com/openfaas/go-sdk/stack"
)

const (
	SuccessMsg       = `(?m:Function created in folder)`
	InvalidYAMLMsg   = `is not valid YAML`
	InvalidYAMLMap   = `map is empty`
	IncludeUpperCase = "function name can only contain a-z, 0-9 and dashes"
	NoFunctionName   = "please provide a name for the function"
	NoLanguage       = "you must supply a function language with the --lang flag"

	InvalidFile      = "unable to find file: (.+)? - (.+)?"
	ListOptionOutput = `Languages available as templates:
- dockerfile
- ruby`

	LangNotExistsOutput  = `(template: \"([0-9A-Za-z-])*\" was not found in the templates folder or in the store)`
	FunctionExistsOutput = `(Function (.+)? already exists in (.+)? file)`
	NoTemplates          = `no language templates were found.

Download templates:
  faas-cli template pull                download the default templates
  faas-cli template store list          view the template store
  faas-cli template store pull NAME     download the default templates
  faas-cli new --lang NAME              Attempt to download NAME from the template store`
	InvalidFileSuffix = "when appending to a stack the suffix should be .yml or .yaml"
)

type NewFunctionTest struct {
	title         string
	prefix        string
	funcName      string
	funcLang      string
	dirName       string
	expectedImage string
	expectedMsg   string

	stackYaml     string
	wantStackYaml string
}

var newFunctionTests = []NewFunctionTest{
	{
		title:         "new-test-1",
		funcName:      "new-test-1",
		funcLang:      "dockerfile",
		expectedImage: "new-test-1:latest",
		expectedMsg:   SuccessMsg,
		stackYaml:     "new-test-1.yml",
		wantStackYaml: "new-test-1.yml",
	},
	{
		title:         "lowercase-dockerfile",
		funcName:      "lowercase-dockerfile",
		funcLang:      "dockerfile",
		expectedImage: "lowercase-dockerfile:latest",
		expectedMsg:   SuccessMsg,
		stackYaml:     "lowercase-dockerfile.yml",
		wantStackYaml: "lowercase-dockerfile.yml",
	},
	{
		title:         "stack.yaml as default",
		funcName:      "default-stack",
		funcLang:      "dockerfile",
		expectedImage: "default-stack:latest",
		expectedMsg:   SuccessMsg,
		stackYaml:     "",
		wantStackYaml: "stack.yaml",
	},
	{
		title:         "func-with-prefix",
		funcName:      "func-with-prefix",
		prefix:        " username ",
		funcLang:      "dockerfile",
		expectedImage: "username/func-with-prefix:latest",
		expectedMsg:   SuccessMsg,
		stackYaml:     "func-with-prefix.yml",
		wantStackYaml: "func-with-prefix.yml",
	},
	{
		title:         "func-with-whitespace-only-prefix",
		funcName:      "func-with-whitespace-only-prefix",
		prefix:        " ",
		funcLang:      "dockerfile",
		expectedImage: "func-with-whitespace-only-prefix:latest",
		expectedMsg:   SuccessMsg,
		stackYaml:     "func-with-whitespace-only-prefix.yml",
		wantStackYaml: "func-with-whitespace-only-prefix.yml",
	},
	{
		title:         "long-name-with-hyphens",
		funcName:      "long-name-with-hyphens",
		dirName:       "customname",
		prefix:        " ",
		funcLang:      "dockerfile",
		expectedImage: "long-name-with-hyphens:latest",
		expectedMsg:   SuccessMsg,
		stackYaml:     "long-name-with-hyphens.yml",
		wantStackYaml: "long-name-with-hyphens.yml",
	},
	{
		title:         "template_not_found",
		funcName:      "template_not_found",
		funcLang:      "docker",
		expectedMsg:   LangNotExistsOutput,
		stackYaml:     "template_not_found.yml",
		wantStackYaml: "template_not_found.yml",
	},
	{
		title:         "test_Uppercase",
		funcName:      "test_Uppercase",
		funcLang:      "dockerfile",
		expectedMsg:   IncludeUpperCase,
		stackYaml:     "test_Uppercase.yml",
		wantStackYaml: "test_Uppercase.yml",
	},
	{
		title:         "no-function-name",
		funcName:      "",
		funcLang:      "--lang node",
		expectedMsg:   NoFunctionName,
		stackYaml:     "",
		wantStackYaml: "no-function-name.yml",
	},
	{
		title:         "no-language",
		funcName:      "no-language",
		funcLang:      "",
		expectedMsg:   NoLanguage,
		stackYaml:     "no-language.yml",
		wantStackYaml: "no-language.yml",
	},
}

func runNewFunctionTest(t *testing.T, nft NewFunctionTest) {
	funcName := nft.funcName
	funcLang := nft.funcLang
	dirName := nft.dirName
	imagePrefix := nft.prefix

	var funcYAML string

	if len(nft.stackYaml) > 0 {
		funcYAML = nft.stackYaml
	} else {
		funcYAML = defaultYAML
	}

	cmdParameters := []string{
		"new",
		"--lang=" + funcLang,
		"--gateway=" + defaultGateway,
	}
	if len(imagePrefix) > 0 {
		cmdParameters = append(cmdParameters, "--prefix="+imagePrefix)
	}

	if len(dirName) != 0 {
		cmdParameters = append(cmdParameters, "--handler="+dirName)
	} else {
		dirName = funcName
	}

	if len(funcYAML) > 0 && nft.stackYaml != defaultYAML {
		cmdParameters = append(cmdParameters, "--yaml="+funcYAML)
	}

	if len(funcName) != 0 {
		cmdParameters = append(cmdParameters, funcName)
	}

	if nft.stackYaml == "" {
		t.Logf("Cmd: %+v", cmdParameters)
	}

	faasCmd.SetArgs(cmdParameters)
	if err := faasCmd.Execute(); err != nil {

		if nft.expectedMsg != "" && nft.expectedMsg != SuccessMsg {
			// Validate new function output
			if found, err := regexp.MatchString(nft.expectedMsg, err.Error()); err != nil || !found {
				t.Logf("No match for:\n%s\nin\n%s\n", err, nft.expectedMsg)
			}
		} else {
			t.Fatalf("Error: %v", err)
		}
	}

	if nft.expectedMsg == SuccessMsg {
		cwd, _ := os.Getwd()

		handlerPath := path.Join(cwd, dirName)
		if len(nft.stackYaml) == 0 {
			t.Logf("Should have a function created in folder: %s", handlerPath)
			t.Logf("Should have a function created in folder: %s", nft.stackYaml)
		}

		// Make sure that the folder and file was created:
		if _, err := os.Stat(handlerPath); err != nil {
			if os.IsNotExist(err) {
				t.Fatalf("%s/ directory was not created", handlerPath)
			} else {
				t.Fatalf("Error: %v", err)
			}
		}

		// Check that the Dockerfile was created
		if funcLang == "Dockerfile" || funcLang == "dockerfile" {
			if _, err := os.Stat("./" + dirName + "/Dockerfile"); err != nil && os.IsNotExist(err) {
				t.Fatalf("Dockerfile language should create a Dockerfile for you: %s", funcName)
			}
		}

		if _, err := os.Stat(funcYAML); err != nil && os.IsNotExist(err) {
			t.Fatalf("%s was not created", funcYAML)
		}

		// Make sure that the information in the YAML file is correct:
		parsedServices, err := stack.ParseYAMLFile(funcYAML, "", "", false)
		if err != nil {
			t.Fatalf("Couldn't open modified YAML file \"%s\" due to error: %v", funcYAML, err)
		}
		services := *parsedServices

		var testServices stack.Services

		testServices.Version = defaultSchemaVersion
		if services.Version != testServices.Version {
			t.Fatalf("YAML `version` section was not created correctly for file %s: got %v", funcYAML, services.Version)
		}

		testServices.Provider = stack.Provider{Name: "openfaas", GatewayURL: defaultGateway}
		if !reflect.DeepEqual(services.Provider, testServices.Provider) {
			t.Fatalf("YAML `provider` section was not created correctly for file %s: got %v", funcYAML, services.Provider)
		}

		testServices.Functions = make(map[string]stack.Function)
		testServices.Functions[funcName] = stack.Function{Language: funcLang, Image: nft.expectedImage, Handler: "./" + dirName}
		if !reflect.DeepEqual(services.Functions[funcName], testServices.Functions[funcName]) {
			t.Fatalf("YAML `functions` section was not created correctly for file %s, got %v", funcYAML, services.Functions[funcName])
		}
	}
}

func Test_newFunctionTests(t *testing.T) {
	// Download templates
	templatePullLocalTemplateRepo(t)
	defer tearDownFetchTemplates(t)

	cwd, _ := os.Getwd()
	for _, testcase := range newFunctionTests {
		t.Run(testcase.title, func(t *testing.T) {

			name := testcase.funcName
			wantStackYaml := testcase.wantStackYaml
			// Clean-up functions and stack.yaml files created by the tests
			// These should be removed at the end of the function's execution
			defer tearDownNewFunction(t, name, wantStackYaml)

			// Remove any left over stack.yaml files from a previous run
			os.Remove(path.Join(cwd, wantStackYaml))

			// Clean-up any existing handler folder before running
			// only do this for tests that are meant to pass and create
			// a valid folder
			handlerDir := testcase.dirName
			if len(handlerDir) == 0 {
				handlerDir = testcase.funcName
			}

			if testcase.expectedMsg == SuccessMsg {
				os.Remove(path.Join(cwd, handlerDir))
			}

			runNewFunctionTest(t, testcase)

			defer os.Remove(path.Join(cwd, testcase.wantStackYaml))
		})
	}
}

func Test_newFunctionListCmds(t *testing.T) {
	// Download templates
	templatePullLocalTemplateRepo(t)
	defer tearDownFetchTemplates(t)

	cmdParameters := []string{
		"new",
		"--list",
	}

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs(cmdParameters)
		faasCmd.Execute()
	})

	// Validate command output
	if !strings.HasPrefix(stdOut, ListOptionOutput) {
		t.Fatalf("Output is not as expected: %s\n", stdOut)
	}
}

func Test_newFunctionListNoTemplates(t *testing.T) {
	cmdParameters := []string{
		"new",
		"--list",
	}

	faasCmd.SetArgs(cmdParameters)
	stdOut := faasCmd.Execute().Error()

	// Validate command output
	if !strings.HasPrefix(stdOut, NoTemplates) {
		t.Fatalf("output of new --list incorrect, \nwant: %q, \ngot:  %q.\n", NoTemplates, stdOut)
	}
}

func Test_languageNotExists(t *testing.T) {
	// Download templates
	templatePullLocalTemplateRepo(t)
	defer tearDownFetchTemplates(t)

	// Attempt to create a function with a non-existing language
	cmdParameters := []string{
		"new",
		"samplename",
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

func Test_appendInvalidSuffix(t *testing.T) {
	const functionName = "samplefunc"
	const functionLang = "dockerfile"

	templatePullLocalTemplateRepo(t)
	defer tearDownFetchTemplates(t)

	// Create function
	parameters := []string{
		"new",
		functionName,
		"--lang=" + functionLang,
		"--append=" + functionName + ".txt",
	}
	faasCmd.SetArgs(parameters)
	stdOut := faasCmd.Execute().Error()

	if found, err := regexp.MatchString(InvalidFileSuffix, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected: %s\n", stdOut)
	}
}

func Test_appendInvalidFile(t *testing.T) {
	const functionName = "samplefunc"
	const functionLang = "dockerfile"

	templatePullLocalTemplateRepo(t)
	defer tearDownFetchTemplates(t)

	// Create function
	parameters := []string{
		"new",
		functionName,
		"--lang=" + functionLang,
		"--append=" + functionLang + ".yml",
	}
	faasCmd.SetArgs(parameters)
	stdOut := faasCmd.Execute().Error()

	if found, err := regexp.MatchString(InvalidFile, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected: %s\n", stdOut)
	}
}

func Test_duplicateFunctionName(t *testing.T) {
	resetForTest()

	const functionName = "samplefunc"
	const functionLang = "dockerfile"
	const functionYaml = functionName + ".yml"

	templatePullLocalTemplateRepo(t)
	defer tearDownFetchTemplates(t)
	defer tearDownNewFunction(t, functionName, functionYaml)

	// Create function
	parameters := []string{
		"new",
		functionName,
		"--lang=" + functionLang,
		"--yaml=" + functionYaml,
	}
	faasCmd.SetArgs(parameters)
	faasCmd.Execute()

	// Attempt to create duplicate function
	parameters = append(parameters, "--append="+functionName+".yml")
	faasCmd.SetArgs(parameters)
	stdOut := faasCmd.Execute().Error()

	if found, err := regexp.MatchString(FunctionExistsOutput, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected: %s\n", stdOut)
	}
}

func Test_backfillTemplates(t *testing.T) {
	resetForTest()
	const functionName = "samplefunc"
	const functionLang = "dockerfile"
	const functionYaml = "samplefunc.yml"

	// Delete cached templates
	localTemplateRepository := setupLocalTemplateRepo(t)
	defer os.RemoveAll(localTemplateRepository)
	defer tearDownNewFunction(t, functionName, functionYaml)

	os.Setenv(templateURLEnvironment, localTemplateRepository)
	defer os.Unsetenv(templateURLEnvironment)

	parameters := []string{
		"new",
		functionName,
		"--lang=" + functionLang,
		"--yaml=" + functionYaml,
	}

	faasCmd.SetArgs(parameters)
	err := faasCmd.Execute()
	if err != nil {
		t.Fatalf("Failed to create function with custom template dir: %s\n", err.Error())
	}
}

func tearDownNewFunction(t *testing.T, functionName, functionStackYml string) {
	if _, err := os.Stat(".gitignore"); err == nil {
		if err := os.Remove(".gitignore"); err != nil {
			t.Log(err)
		}
	}
	hDir := handlerDir
	if len(hDir) == 0 {
		hDir = functionName
	}
	if _, err := os.Stat(hDir); err == nil {
		if err := os.RemoveAll(hDir); err != nil {
			t.Log(err)
		}
	}

	if len(functionStackYml) == 0 {
		functionYaml := functionName + ".yml"
		if _, err := os.Stat(functionYaml); err == nil {
			if err := os.Remove(functionYaml); err != nil {
				t.Log(err)
			}
		}
	} else {
		os.RemoveAll(functionStackYml)
	}

	handlerDir = ""
}

func Test_getPrefixValue_Default(t *testing.T) {
	os.Unsetenv("OPENFAAS_PREFIX")

	imagePrefix = ""

	val := getPrefixValue()
	want := ""
	if val != want {
		t.Errorf("want %s, got %s", want, val)
	}
}

func Test_getPrefixValue_Env(t *testing.T) {
	want := "alexellis"
	os.Setenv("OPENFAAS_PREFIX", want)
	imagePrefix = ""

	val := getPrefixValue()
	if val != want {
		t.Errorf("want %s, got %s", want, val)
	}
}

func Test_getPrefixValue_Flag(t *testing.T) {
	want := "other"
	os.Unsetenv("OPENFAAS_PREFIX")
	imagePrefix = "other"

	val := getPrefixValue()
	if val != want {
		t.Errorf("want %s, got %s", want, val)
	}
}
