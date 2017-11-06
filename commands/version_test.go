// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"regexp"
	"testing"

	"github.com/openfaas/faas-cli/test"
	"github.com/openfaas/faas-cli/version"
)

func Test_addVersionDev(t *testing.T) {
	version.GitCommit = "sha-test"

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{"version"})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:Commit: sha-test)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}

	if found, err := regexp.MatchString(`(?m:Version: dev)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_addVersion(t *testing.T) {
	version.GitCommit = "sha-test"
	version.Version = "version.tag"

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{"version"})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString(`(?m:Commit: sha-test)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}

	if found, err := regexp.MatchString(`(?m:Version: version.tag)`, stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}

func Test_addVersion_short_version(t *testing.T) {
	version.Version = "version.tag"

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{"version", "--short-version"})
		faasCmd.Execute()
	})

	if found, err := regexp.MatchString("^version\\.tag", stdOut); err != nil || !found {
		t.Fatalf("Output is not as expected:\n%s", stdOut)
	}
}
