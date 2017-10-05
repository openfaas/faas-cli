// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package builder

import (
	"testing"

	"os"

	"bytes"
	"os/exec"
	"regexp"
	"strings"
)

func Test_buildLanguage_LanguageNotSupported(t *testing.T) {
	defer teardown()

	if len(os.Getenv("TESTCASE")) > 0 {
		if os.Getenv("TESTCASE") == "Test_buildLanguage_LanguageNotSupported" {
			// Change directory to testdata
			if err := os.Chdir("testdata"); err != nil {
				t.Fatalf("Error on cd to testdata dir: %v", err)
			}

			BuildImage("image", "test-function", "function-name", "python1000", true, false)
		}
		return
	}

	stdOut := bytes.NewBuffer(nil)
	stdErr := bytes.NewBuffer(nil)

	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(), "TESTCASE=Test_buildLanguage_LanguageNotSupported")
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr
	err := cmd.Run()

	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		if idx := strings.Index(stdErr.String(), "Language template: python1000 not supported. Build a custom Dockerfile instead."); idx == -1 {
			t.Fatalf("Output is not what expected\n%s", stdErr)
		}
		return
	}
	t.Fatalf("Expected a non-0 exit code")
}

func Test_buildLanguage_Language(t *testing.T) {
	defer teardown()

	if len(os.Getenv("TESTCASE")) > 0 {
		if os.Getenv("TESTCASE") == "Test_buildLanguage_Language" {
			// Change directory to testdata
			if err := os.Chdir("testdata"); err != nil {
				t.Fatalf("Error on cd to testdata dir: %v", err)
			}

			BuildImage("image", "test-function", "function", "python3", true, false)
		}
		return
	}

	stdOut := bytes.NewBuffer(nil)
	stdErr := bytes.NewBuffer(nil)

	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(), "TESTCASE=Test_buildLanguage_Language")
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr
	cmd.Run()

	if found, err := regexp.MatchString(`(?m:Clearing temporary build folder: \./build/function/)`, stdOut.String()); err != nil || !found {
		t.Fatalf("Output is not what expected\nStdOut:\n%s\nStdErr:\n%s", stdOut, stdErr)
	}
	if found, err := regexp.MatchString(`(?m:Preparing test-function/ \./build/function/function)`, stdOut.String()); err != nil || !found {
		t.Fatalf("Output is not what expected\nStdOut:\n%s\nStdErr:\n%s", stdOut, stdErr)
	}
}

func Test_buildLanguage_Dockerfile(t *testing.T) {
	defer teardown()

	if len(os.Getenv("TESTCASE")) > 0 {
		if os.Getenv("TESTCASE") == "Test_buildLanguage_Dockerfile" {
			// Change directory to testdata
			if err := os.Chdir("testdata"); err != nil {
				t.Fatalf("Error on cd to testdata dir: %v", err)
			}

			BuildImage("image", "test-function-Dockerfile", "function", "Dockerfile", true, false)
		}
		return
	}

	stdOut := bytes.NewBuffer(nil)
	stdErr := bytes.NewBuffer(nil)

	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(), "TESTCASE=Test_buildLanguage_Dockerfile")
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr
	cmd.Run()

	if found, err := regexp.MatchString(`(?m:Building: image with Dockerfile\. Please wait\.\.)`, stdOut.String()); err != nil || !found {
		t.Fatalf("Output is not what expected\nStdOut:\n%s\nStdErr:\n%s", stdOut, stdErr)
	}
}

func teardown() {
	if err := os.Chdir("testdata"); err == nil {
		//os.RemoveAll("build")
		os.Chdir("..")
	}
}
