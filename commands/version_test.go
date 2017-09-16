// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package commands

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"testing"
)

func Test_addVersionDev(t *testing.T) {
	GitCommit = "sha-test"

	output := captureStdout(func() {
		faasCmd.SetArgs([]string{"version"})
		faasCmd.Execute()
	})

	rCommit := regexp.MustCompile(`(?m:Commit: sha-test)`)
	if !rCommit.MatchString(output) {
		t.Fatal(output)
	}

	rVersion := regexp.MustCompile(`(?m:Version: dev)`)
	if !rVersion.MatchString(output) {
		t.Fatal(output)
	}
}

func Test_addVersion(t *testing.T) {
	GitCommit = "sha-test"
	Version = "version.tag"

	output := captureStdout(func() {
		faasCmd.SetArgs([]string{"version"})
		faasCmd.Execute()
	})

	rCommit := regexp.MustCompile(`(?m:Commit: sha-test)`)
	if !rCommit.MatchString(output) {
		t.Fatal(output)
	}

	rVersion := regexp.MustCompile(`(?m:Version: version.tag)`)
	if !rVersion.MatchString(output) {
		t.Fatal(output)
	}
}

func Test_addVersion_short_version(t *testing.T) {
	Version = "version.tag"

	output := captureStdout(func() {
		faasCmd.SetArgs([]string{"version", "--short-version"})
		faasCmd.Execute()
	})

	rVersion := regexp.MustCompile("^version\\.tag")
	if !rVersion.MatchString(output) {
		t.Fatal(output)
	}
}

func captureStdout(f func()) string {
	stdOut := os.Stdout
	r, w, _ := os.Pipe()
	defer r.Close()
	defer w.Close()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = stdOut

	var b bytes.Buffer
	io.Copy(&b, r)

	return b.String()
}
