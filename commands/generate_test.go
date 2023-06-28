// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package commands

import (
	"testing"

	v2 "github.com/openfaas/faas-cli/schema/store/v2"

	"github.com/openfaas/faas-cli/schema"

	"github.com/openfaas/faas-cli/stack"
)

var generateTestcases = []struct {
	Name       string
	Input      string
	Output     []string
	Format     schema.BuildFormat
	APIVersion string
	Namespace  string
	Branch     string
	Version    string
}{
	{
		Name: "Specified Namespace and API Version",
		Input: `
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
  network: "func_functions"      
functions:
 url-ping:
   lang: python
   handler: ./sample/url-ping
   image: alexellis/faas-url-ping:0.2`,
		Output: []string{`---
apiVersion: openfaas.com/v1
kind: Function
metadata:
  name: url-ping
  namespace: openfaas-fn
spec:
  name: url-ping
  image: alexellis/faas-url-ping:0.2
`},
		Format:     schema.DefaultFormat,
		APIVersion: "openfaas.com/v1",
		Namespace:  "openfaas-fn",
		Branch:     "",
		Version:    "",
	},
	{
		Name: "Annotation present",
		Input: `
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
functions:
 url-ping:
   lang: python
   handler: ./sample/url-ping
   image: alexellis/faas-url-ping:0.2
   annotations:
     com.scale.zero: 1
`,
		Output: []string{`---
apiVersion: openfaas.com/v1
kind: Function
metadata:
  name: url-ping
  namespace: openfaas-fn
spec:
  name: url-ping
  image: alexellis/faas-url-ping:0.2
  annotations:
    com.scale.zero: "1"
`},
		Format:     schema.DefaultFormat,
		APIVersion: "openfaas.com/v1",
		Namespace:  "openfaas-fn",
		Branch:     "",
		Version:    "",
	},
	{
		Name: "Blank namespace",
		Input: `
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
  network: "func_functions"
functions:
 url-ping:
  lang: python
  handler: ./sample/url-ping
  image: alexellis/faas-url-ping:0.2`,
		Output: []string{`---
apiVersion: openfaas.com/v1
kind: Function
metadata:
  name: url-ping
spec:
  name: url-ping
  image: alexellis/faas-url-ping:0.2
`},
		Format:     schema.DefaultFormat,
		APIVersion: "openfaas.com/v1",
		Namespace:  "",
		Branch:     "",
		Version:    "",
	},
	{
		Name: "BranchAndSHA Image format",
		Input: `
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
  network: "func_functions"
functions:
 url-ping:
  lang: python
  handler: ./sample/url-ping
  image: alexellis/faas-url-ping:0.2`,
		Output: []string{`---
apiVersion: openfaas.com/v1
kind: Function
metadata:
  name: url-ping
  namespace: openfaas-function
spec:
  name: url-ping
  image: alexellis/faas-url-ping:0.2-master-6bgf36qd
`},
		Format:     schema.BranchAndSHAFormat,
		APIVersion: "openfaas.com/v1",
		Namespace:  "openfaas-function",
		Branch:     "master",
		Version:    "6bgf36qd",
	},
	{
		Name: "Multiple functions",
		Input: `
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080  
  network: "func_functions"
functions:
 url-ping:
  lang: python
  handler: ./sample/url-ping
  image: alexellis/faas-url-ping:0.2
 astronaut-finder:
  lang: python3
  handler: ./astronaut-finder
  image: astronaut-finder
  environment:
   write_debug: true`,
		Output: []string{`---
apiVersion: openfaas.com/v2alpha2
kind: Function
metadata:
  name: url-ping
  namespace: openfaas-fn
spec:
  name: url-ping
  image: alexellis/faas-url-ping:0.2
---
apiVersion: openfaas.com/v2alpha2
kind: Function
metadata:
  name: astronaut-finder
  namespace: openfaas-fn
spec:
  name: astronaut-finder
  image: astronaut-finder:latest
  environment:
    write_debug: "true"
`, `---
apiVersion: openfaas.com/v2alpha2
kind: Function
metadata:
  name: astronaut-finder
  namespace: openfaas-fn
spec:
  name: astronaut-finder
  image: astronaut-finder:latest
  environment:
    write_debug: "true"
---
apiVersion: openfaas.com/v2alpha2
kind: Function
metadata:
  name: url-ping
  namespace: openfaas-fn
spec:
  name: url-ping
  image: alexellis/faas-url-ping:0.2
`},
		Format:     schema.DefaultFormat,
		APIVersion: "openfaas.com/v2alpha2",
		Namespace:  "openfaas-fn",
		Branch:     "",
		Version:    "",
	},
	{
		Name: "Read-only root filesystem",
		Input: `
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
functions:
 url-ping:
  lang: python
  handler: ./sample/url-ping
  image: alexellis/faas-url-ping:0.2
  readonly_root_filesystem: true`,
		Output: []string{`---
apiVersion: openfaas.com/v1
kind: Function
metadata:
  name: url-ping
  namespace: openfaas-fn
spec:
  name: url-ping
  image: alexellis/faas-url-ping:0.2
  readOnlyRootFilesystem: true
`},
		Format:     schema.DefaultFormat,
		APIVersion: "openfaas.com/v1",
		Namespace:  "openfaas-fn",
		Branch:     "",
		Version:    "",
	},
}

func Test_generateCRDYAML(t *testing.T) {

	for _, testcase := range generateTestcases {
		parsedServices, err := stack.ParseYAMLData([]byte(testcase.Input), "", "", true)

		if err != nil {
			t.Fatalf("%s failed: error while parsing the input data", testcase.Name)
		}

		if parsedServices == nil {
			t.Fatalf("%s failed: empty input file", testcase.Name)
		}
		services := *parsedServices

		generatedYAML, err := generateCRDYAML(services, testcase.Format, testcase.APIVersion, testcase.Namespace,
			NewFunctionMetadataSourceStub(testcase.Branch, testcase.Version))
		if err != nil {
			t.Fatalf("%s failed: error while generating CRD YAML", testcase.Name)
		}

		if !stringInSlice(generatedYAML, testcase.Output) {
			t.Fatalf("%s failed: want:\n%q, but got:\n%q", testcase.Name, testcase.Output, generatedYAML)
		}
	}
}

type FunctionMetadataSourceStub struct {
	version string
	branch  string
}

func NewFunctionMetadataSourceStub(branch, version string) FunctionMetadataSourceStub {
	return FunctionMetadataSourceStub{
		version: version,
		branch:  branch,
	}
}

func (f FunctionMetadataSourceStub) Get(tagType schema.BuildFormat, contextPath string) (branch, version string, err error) {
	return f.branch, f.version, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func Test_filterStoreItem_Found(t *testing.T) {

	items := []v2.StoreFunction{
		{
			Name: "figlet",
		},
	}
	wantName := "figlet"
	got, gotErr := filterStoreItem(items, wantName)

	if gotErr != nil {
		t.Errorf("got error %s", gotErr)
		t.Fail()
	}

	if got.Name != wantName {
		t.Errorf("name got: %s, but want: %s", got.Name, wantName)
		t.Fail()
	}

}

func Test_filterStoreItem_NotFound(t *testing.T) {

	items := []v2.StoreFunction{
		{
			Name: "figlets",
		},
	}
	wantName := "figlet"
	got, gotErr := filterStoreItem(items, wantName)

	if got != nil {
		t.Errorf("want nil, but got item %s", got.Name)
		t.Fail()
	}

	if gotErr == nil {
		t.Errorf("want error, got nil")
		t.Fail()
	}
	wantError := "unable to find 'figlet' in store"
	if gotErr.Error() != wantError {
		t.Errorf("want error %s, got %s", wantError, gotErr.Error())
		t.Fail()
	}

}

var generateOrderedTestcases = []struct {
	Name      string
	Input     string
	Output    []string
	ExpectErr bool
}{
	{
		Name: "Smae order in and out",
		Input: `
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
  network: "func_functions"    
functions:
 fn1:
  lang: python3
  handler: ./fn1
  image: fn1:latest
 fn2:
  lang: python3
  handler: ./fn2
  image: fn2:latest
 fn3:
  lang: python3
  handler: ./fn3
  image: fn3:latest
 fn4:
  lang: python3
  handler: ./fn4
  image: fn4:latest
 fn5:
  lang: python3
  handler: ./fn5
  image: fn5:latest
 fn6:
  lang: python3
  handler: ./fn6
  image: fn6:latest
 fn7:
  lang: python3
  handler: ./fn7
  image: fn7:latest
 fn8:
  lang: python3
  handler: ./fn8
  image: fn8:latest
 fn9:
  lang: python3
  handler: ./fn9
  image: fn9:latest
 fn10:
  lang: python3
  handler: ./fn10
  image: fn10:latest`,
		Output: []string{
			"fn1", "fn10",
			"fn2", "fn3",
			"fn4", "fn5",
			"fn6", "fn7",
			"fn8", "fn9",
		},
		ExpectErr: false,
	},
	{
		Name: "Different input order",
		Input: `
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
  network: "func_functions"    
functions:
 fn3:
  lang: python3
  handler: ./fn3
  image: fn3:latest
 fn7:
  lang: python3
  handler: ./fn7
  image: fn7:latest
 fn2:
  lang: python3
  handler: ./fn2
  image: fn2:latest
 fn10:
  lang: python3
  handler: ./fn10
  image: fn10:latest
 fn5:
  lang: python3
  handler: ./fn5
  image: fn5:latest
 fn1:
  lang: python3
  handler: ./fn1
  image: fn1:latest
 fn6:
  lang: python3
  handler: ./fn6
  image: fn6:latest
 fn9:
  lang: python3
  handler: ./fn9
  image: fn9:latest
 fn4:
  lang: python3
  handler: ./fn4
  image: fn4:latest
 fn8:
  lang: python3
  handler: ./fn8
  image: fn8:latest`,
		Output: []string{
			"fn1", "fn10",
			"fn2", "fn3",
			"fn4", "fn5",
			"fn6", "fn7",
			"fn8", "fn9",
		},
		ExpectErr: false,
	},
	{
		Name: "Different input order wrong output order - expect error",
		Input: `
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
  network: "func_functions"    
functions:
 fn3:
  lang: python3
  handler: ./fn3
  image: fn3:latest
 fn7:
  lang: python3
  handler: ./fn7
  image: fn7:latest
 fn2:
  lang: python3
  handler: ./fn2
  image: fn2:latest
 fn10:
  lang: python3
  handler: ./fn10
  image: fn10:latest
 fn5:
  lang: python3
  handler: ./fn5
  image: fn5:latest
 fn1:
  lang: python3
  handler: ./fn1
  image: fn1:latest
 fn6:
  lang: python3
  handler: ./fn6
  image: fn6:latest
 fn9:
  lang: python3
  handler: ./fn9
  image: fn9:latest
 fn4:
  lang: python3
  handler: ./fn4
  image: fn4:latest
 fn8:
  lang: python3
  handler: ./fn8
  image: fn8:latest`,
		Output: []string{
			"fn1",
			"fn2", "fn3",
			"fn4", "fn5",
			"fn6", "fn7",
			"fn8", "fn9",
			"fn10",
		},
		ExpectErr: true,
	},
}

func Test_generateFunctionOrder(t *testing.T) {
	for _, testcase := range generateOrderedTestcases {
		parsedServices, err := stack.ParseYAMLData([]byte(testcase.Input), "", "", true)
		if err != nil {
			t.Fatalf("%s failed: error while parsing the input data.", testcase.Name)
		}

		if parsedServices == nil {
			t.Fatalf("%s failed: empty input file", testcase.Name)
		}
		services := *parsedServices
		orderedSlice := generateFunctionOrder(services.Functions)

		if len(orderedSlice) != len(testcase.Output) {
			t.Errorf("Slice sizes do not match: %s", testcase.Name)
			t.Fail()
		}
		for i, v := range testcase.Output {
			if v != orderedSlice[i] && !testcase.ExpectErr {
				t.Errorf("Exected %s got %s: %s", v, orderedSlice[i], testcase.Name)
				t.Fail()
			}
		}

	}
}
