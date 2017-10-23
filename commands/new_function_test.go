// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"os"
	"regexp"
	"testing"

	"github.com/openfaas/faas-cli/test"
)

func Test_new(t *testing.T) {
	//TODO activate the teardown when PR#87 is merged defer teardown()
	funcName := "test-function"

	// Symlink template directory at root to current directory to avoid re-downloading templates
	os.Symlink("../template", "template")

	// Cleanup the created directory
	defer func() {
		os.RemoveAll(funcName)
		os.Remove(funcName + ".yml")
		os.Remove("template")
	}()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"new",
			funcName,
			"--lang=python",
		})
		faasCmd.Execute()

	})

	if found, err := regexp.MatchString(`(?m:Function created in folder)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}
