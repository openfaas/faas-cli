// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"reflect"
	"testing"
)

var translateLegacyOptsTests = []struct {
	title        string
	inputArgs    []string
	expectedArgs []string
	expectError  bool
}{
	{
		title:        "legacy deploy action with all args, no =",
		inputArgs:    []string{"faas-cli", "-action", "deploy", "-image", "testimage", "-name", "fnname", "-fprocess", `"/usr/bin/faas-img2ansi"`, "-gateway", "https://url", "-handler", "/dir/", "-lang", "python", "-replace"},
		expectedArgs: []string{"faas-cli", "deploy", "--image", "testimage", "--name", "fnname", "--fprocess", `"/usr/bin/faas-img2ansi"`, "--gateway", "https://url", "--handler", "/dir/", "--lang", "python", "--replace"},
		expectError:  false,
	},
	{
		title:        "legacy deploy action with =",
		inputArgs:    []string{"faas-cli", "-action=deploy", "-image=testimage", "-name=fnname", `-fprocess="/usr/bin/faas-img2ansi"`},
		expectedArgs: []string{"faas-cli", "deploy", "--image=testimage", "--name=fnname", `--fprocess="/usr/bin/faas-img2ansi"`},
		expectError:  false,
	},
	{
		title:        "legacy deploy action with -f",
		inputArgs:    []string{"faas-cli", "-action=deploy", "-f", "/dir/file.yml"},
		expectedArgs: []string{"faas-cli", "deploy", "-f", "/dir/file.yml"},
		expectError:  false,
	},
	{
		title:        "legacy deploy action with -yaml",
		inputArgs:    []string{"faas-cli", "-action=deploy", "-yaml", "/dir/file.yml"},
		expectedArgs: []string{"faas-cli", "deploy", "--yaml", "/dir/file.yml"},
		expectError:  false,
	},
	{
		title:        "legacy build action with all args, no =",
		inputArgs:    []string{"faas-cli", "-action", "build", "-image", "testimage", "-name", "fnname", "-handler", "/dir/", "-lang", "python", "-no-cache", "-squash"},
		expectedArgs: []string{"faas-cli", "build", "--image", "testimage", "--name", "fnname", "--handler", "/dir/", "--lang", "python", "--no-cache", "--squash"},
		expectError:  false,
	},
	{
		title:        "legacy delete action (note delete->remove translation)",
		inputArgs:    []string{"faas-cli", "-action", "delete", "-name", "fnname"},
		expectedArgs: []string{"faas-cli", "remove", "fnname"},
		expectError:  false,
	},
	{
		title:        "legacy delete action with yaml",
		inputArgs:    []string{"faas-cli", "-action", "delete", "-f", "/dir/file.yml"},
		expectedArgs: []string{"faas-cli", "remove", "-f", "/dir/file.yml"},
		expectError:  false,
	},
	{
		title:        "legacy version flag",
		inputArgs:    []string{"faas-cli", "-version"},
		expectedArgs: []string{"faas-cli", "version"},
		expectError:  false,
	},
	{
		title:        "version command",
		inputArgs:    []string{"faas-cli", "version"},
		expectedArgs: []string{"faas-cli", "version"},
		expectError:  false,
	},
	{
		title:        "deploy command",
		inputArgs:    []string{"faas-cli", "deploy", "--image", "testimage", "--name", "fnname", "--fprocess", `"/usr/bin/faas-img2ansi"`, "--gateway", "https://url", "--handler", "/dir/", "--lang", "python", "--replace", "--env", "KEY1=VAL1", "--env", "KEY2=VAL2"},
		expectedArgs: []string{"faas-cli", "deploy", "--image", "testimage", "--name", "fnname", "--fprocess", `"/usr/bin/faas-img2ansi"`, "--gateway", "https://url", "--handler", "/dir/", "--lang", "python", "--replace", "--env", "KEY1=VAL1", "--env", "KEY2=VAL2"},
		expectError:  false,
	},
	{
		title:        "build command",
		inputArgs:    []string{"faas-cli", "build", "--image", "testimage", "--name", "fnname", "--handler", "/dir/", "--lang", "python", "--no-cache", "--squash"},
		expectedArgs: []string{"faas-cli", "build", "--image", "testimage", "--name", "fnname", "--handler", "/dir/", "--lang", "python", "--no-cache", "--squash"},
		expectError:  false,
	},
	{
		title:        "remove command",
		inputArgs:    []string{"faas-cli", "remove", "fnname"},
		expectedArgs: []string{"faas-cli", "remove", "fnname"},
		expectError:  false,
	},
	{
		title:        "remove command alias rm",
		inputArgs:    []string{"faas-cli", "rm", "fnname"},
		expectedArgs: []string{"faas-cli", "rm", "fnname"},
		expectError:  false,
	},
	{
		title:        "remove command alias delete",
		inputArgs:    []string{"faas-cli", "delete", "fnname"},
		expectedArgs: []string{"faas-cli", "delete", "fnname"},
		expectError:  false,
	},
	{
		title:        "push command",
		inputArgs:    []string{"faas-cli", "delete", "fnname"},
		expectedArgs: []string{"faas-cli", "delete", "fnname"},
		expectError:  false,
	},
	{
		title:        "bashcompletion command",
		inputArgs:    []string{"faas-cli", "bashcompletion", "/dir/file"},
		expectedArgs: []string{"faas-cli", "bashcompletion", "/dir/file"},
		expectError:  false,
	},
	{
		title:        "legacy flag as value without =",
		inputArgs:    []string{"faas-cli", "-action", "deploy", "-name", `"-name"`},
		expectedArgs: []string{"faas-cli", "deploy", "--name", `"-name"`},
		expectError:  false,
	},
	{
		title:        "legacy flag as value with =",
		inputArgs:    []string{"faas-cli", "-action", "deploy", "-name=-name"},
		expectedArgs: []string{"faas-cli", "deploy", "--name=-name"},
		expectError:  false,
	},
	{
		title:        "unknown legacy flag",
		inputArgs:    []string{"faas-cli", "-action", "deploy", "-fe"},
		expectedArgs: []string{"faas-cli", "deploy", "-fe"},
		expectError:  false,
	},
	{
		title:        "legacy -action missing value",
		inputArgs:    []string{"faas-cli", "-action"},
		expectedArgs: []string{""},
		expectError:  true,
	},
	{
		title:        "legacy -action= missing value",
		inputArgs:    []string{"faas-cli", "-action="},
		expectedArgs: []string{""},
		expectError:  true,
	},
	{
		title:        "legacy -action with unknown value",
		inputArgs:    []string{"faas-cli", "-action", "unknownaction"},
		expectedArgs: []string{""},
		expectError:  true,
	},
	{
		title:        "legacy -action= with unknown value",
		inputArgs:    []string{"faas-cli", "-action=unknownaction"},
		expectedArgs: []string{""},
		expectError:  true,
	},
}

func Test_translateLegacyOpts(t *testing.T) {
	for _, test := range translateLegacyOptsTests {
		t.Run(test.title, func(t *testing.T) {
			actual, err := translateLegacyOpts(test.inputArgs)
			if test.expectError {
				if err == nil {
					t.Errorf("TranslateLegacyOpts test [%s] test failed, expected error not thrown", test.title)
					return
				}
			} else {
				if err != nil {
					t.Errorf("TranslateLegacyOpts test [%s] test failed, unexpected error thrown", test.title)
					return
				}
			}
			if !reflect.DeepEqual(actual, test.expectedArgs) {
				t.Errorf("TranslateLegacyOpts test [%s] test failed, does not match expected result;\n  actual:   [%v]\n  expected: [%v]",
					test.title,
					actual,
					test.expectedArgs,
				)
			}
		})
	}
}
