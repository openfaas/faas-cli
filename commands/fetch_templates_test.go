// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/openfaas/faas-cli/builder"
	"github.com/openfaas/faas-cli/versioncontrol"
)

func Test_PullTemplates(t *testing.T) {
	localTemplateRepository := setupLocalTemplateRepo(t)
	defer os.RemoveAll(localTemplateRepository)
	defer tearDownFetchTemplates(t)

	t.Run("pullTemplates", func(t *testing.T) {
		defer tearDownFetchTemplates(t)
		if err := pullTemplates(localTemplateRepository); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("fetchTemplates with master ref", func(t *testing.T) {
		defer tearDownFetchTemplates(t)

		if err := fetchTemplates(localTemplateRepository, "master", false); err != nil {
			t.Fatal(err)
		}

	})

	t.Run("fetchTemplates with default ref", func(t *testing.T) {
		defer tearDownFetchTemplates(t)

		err := fetchTemplates(localTemplateRepository, "", false)
		if err != nil {
			t.Error(err)
		}

	})
}

// setupLocalTemplateRepo will create a local copy of the core OpenFaaS templates, this
// can be refered to as a local git repository.
func setupLocalTemplateRepo(t *testing.T) string {
	dir, err := os.MkdirTemp("", "openfaas-templates-test-*")
	if err != nil {
		t.Error(err)
	}

	// Copy the submodule to temp directory to avoid altering it during tests
	testRepoGit := filepath.Join("testdata", "templates")
	builder.CopyFiles(testRepoGit, dir)

	// Remove submodule .git file
	os.Remove(filepath.Join(dir, ".git"))
	if err := versioncontrol.GitInitRepo.Invoke(dir, map[string]string{"dir": "."}); err != nil {
		t.Fatal(err)
	}

	return dir
}

// tearDownFetchTemplates cleans all files and directories created by the test
func tearDownFetchTemplates(t *testing.T) {

	// Remove existing templates folder, if it exist
	if _, err := os.Stat("./template/"); err == nil {
		t.Log("Found a ./template/ directory, removing it.")

		if err := os.RemoveAll("./template/"); err != nil {
			t.Log(err)
		}
	} else {
		t.Logf("Directory template was not created: %s", err)
	}
}
