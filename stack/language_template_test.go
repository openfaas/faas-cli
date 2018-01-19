// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package stack

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func Test_ParseYAMLDataForLanguageTemplate(t *testing.T) {
	langTemplateTest := []struct {
		Input    string
		Expected *LanguageTemplate
	}{
		{
			`
language: python
fprocess: python index.py
`,
			&LanguageTemplate{
				Language: "python",
				FProcess: "python index.py",
			},
		},
		{
			`
language: python
`,
			&LanguageTemplate{
				Language: "python",
			},
		},
		{
			`
fprocess: python index.py
`,
			&LanguageTemplate{
				FProcess: "python index.py",
			},
		},
	}

	for k, i := range langTemplateTest {
		t.Run(fmt.Sprintf("%d", k), func(t *testing.T) {
			if actual, err := ParseYAMLDataForLanguageTemplate([]byte(i.Input)); err != nil {
				t.Errorf("test failed, %s", err)
			} else {
				if !reflect.DeepEqual(actual, i.Expected) {
					t.Errorf("does not match expected result;\n  parsedYAML:   [%+v]\n  expected: [%+v]",
						actual,
						i.Expected,
					)
				}
			}
		})
	}
}

func Test_IsValidTemplate(t *testing.T) {
	if IsValidTemplate("unknown-language") {
		t.Fatalf("unknown-language must be invalid")
	}

	os.MkdirAll("template/python", 0600)
	defer func() {
		os.RemoveAll("template")
	}()
	if IsValidTemplate("python") {
		t.Fatalf("python must is not valid because it does not contain template.yml")
	}
}
