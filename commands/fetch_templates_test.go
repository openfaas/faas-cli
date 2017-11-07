// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func Test_PullTemplates(t *testing.T) {
	defer tearDown_fetch_templates(t)

	// Create fake server for testing.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/master_test.zip")
	}))
	defer ts.Close()

	err := PullTemplates(ts.URL)
	if err != nil {
		t.Error(err)
	}

	tearDown_fetch_templates(t)
}

func Test_fetchTemplates(t *testing.T) {
	defer tearDown_fetch_templates(t)

	// Create fake server for testing.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/master_test.zip")
	}))
	defer ts.Close()

	err := fetchTemplates(ts.URL+"/owner/repo", false)
	if err != nil {
		t.Error(err)
	}

	tearDown_fetch_templates(t)
}

// tearDown_fetch_templates_test cleans all files and directories created by the test
func tearDown_fetch_templates(t *testing.T) {

	// Remove existing archive file if it exists
	if _, err := os.Stat("template-owner-repo.zip"); err == nil {
		t.Log("The archive was not deleted")

		err := os.Remove("template-owner-repo.zip")
		if err != nil {
			t.Log(err)
		}
	}

	// Remove existing templates folder, if it exist
	if _, err := os.Stat("template/"); err == nil {
		t.Log("Found a template/ directory, removing it...")

		err := os.RemoveAll("template/")
		if err != nil {
			t.Log(err)
		}
	} else {
		t.Logf("Directory template was not created: %s", err)
	}

	// Verify the downloaded archive
	archive := "template-owner-repo.zip"
	if _, err := os.Stat(archive); err == nil {
		t.Fatalf("The archive %s was not deleted", archive)
	}
}
