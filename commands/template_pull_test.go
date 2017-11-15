// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package commands

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
)

func Test_templatePull(t *testing.T) {
	defer tearDown_fetch_templates(t)

	ts := httpTestServer(t)
	defer ts.Close()

	repository = ts.URL + "/owner/repo"
	faasCmd.SetArgs([]string{"template", "pull", repository})
	faasCmd.Execute()

	// Verify created directories
	if _, err := os.Stat("template"); err != nil {
		t.Fatalf("The directory %s was not created", "template")
	}
}

func Test_templatePull_with_overwriting(t *testing.T) {
	defer tearDown_fetch_templates(t)

	ts := httpTestServer(t)
	defer ts.Close()

	repository = ts.URL + "/owner/repo"
	faasCmd.SetArgs([]string{"template", "pull", repository})
	faasCmd.Execute()

	var buf bytes.Buffer
	log.SetOutput(&buf)

	r := regexp.MustCompile(`(?m:Cannot overwrite the following \d+ directories:)`)

	faasCmd.SetArgs([]string{"template", "pull", repository})
	faasCmd.Execute()

	if !r.MatchString(buf.String()) {
		t.Fatal(buf.String())
	}

	buf.Reset()

	faasCmd.SetArgs([]string{"template", "pull", repository, "--overwrite"})
	faasCmd.Execute()

	str := buf.String()
	if r.MatchString(str) {
		t.Fatal()
	}

	// Verify created directories
	if _, err := os.Stat("template"); err != nil {
		t.Fatalf("The directory %s was not created", "template")
	}
}

func Test_templatePull_no_arg(t *testing.T) {
	defer tearDown_fetch_templates(t)
	var buf bytes.Buffer

	faasCmd.SetArgs([]string{"template", "pull"})
	faasCmd.SetOutput(&buf)
	faasCmd.Execute()

	if strings.Contains(buf.String(), "Error: A repository URL must be specified") {
		t.Fatal("Output does not contain the required string")
	}
}

func Test_templatePull_error_not_valid_url(t *testing.T) {
	var buf bytes.Buffer

	faasCmd.SetArgs([]string{"template", "pull", "git@github.com:openfaas/faas-cli.git"})
	faasCmd.SetOutput(&buf)
	err := faasCmd.Execute()

	if !strings.Contains(err.Error(), "The repository URL must be in the format https://github.com/<owner>/<repository>") {
		t.Fatal("Output does not contain the required string", err.Error())
	}
}

// httpTestServer returns a testing http server
func httpTestServer(t *testing.T) *httptest.Server {
	const sampleMasterZipPath string = "testdata/master_test.zip"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if _, err := os.Stat(sampleMasterZipPath); os.IsNotExist(err) {
			t.Error(err)
		}

		fileData, err := ioutil.ReadFile(sampleMasterZipPath)
		if err != nil {
			t.Error(err)
		}

		w.Write(fileData)
	}))

	return ts
}

func Test_repositoryUrlRegExp(t *testing.T) {
	var url string
	r := regexp.MustCompile(repositoryRegexpGithub)

	url = "http://github.com/owner/repo"
	if r.MatchString(url) {
		t.Errorf("Url %s must start with https", url)
	}

	url = "https://github.com/owner/repo.git"
	if r.MatchString(url) {
		t.Errorf("Url %s must not end with .git or must start with https", url)
	}

	url = "https://github.com/owner/repo//"
	if r.MatchString(url) {
		t.Errorf("Url %s must end with no or one slash", url)
	}

	url = "https://github.com/owner/repo"
	if !r.MatchString(url) {
		t.Errorf("Url %s must be valid", url)
	}

	url = "https://github.com/owner/repo/"
	if !r.MatchString(url) {
		t.Errorf("Url %s must be valid", url)
	}
}
