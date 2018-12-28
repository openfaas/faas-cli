// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package commands

import (
	"io/ioutil"
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

	t.Run("simplePull", func(t *testing.T) {
		defer tearDownFetchTemplates(t)
		if err := PullTemplates(localTemplateRepository); err != nil {
			t.Error(err)
		}
	})

	t.Run("fetchTemplates", func(t *testing.T) {
		defer tearDownFetchTemplates(t)

		err := fetchTemplates(localTemplateRepository, "master", false)
		if err != nil {
			t.Error(err)
		}

	})
}

// setupLocalTemplateRepo will create a local copy of the core OpenFaaS templates, this
// can be refered to as a local git repository.
func setupLocalTemplateRepo(t *testing.T) string {
	dir, err := ioutil.TempDir("", "openFaasTestTemplates")
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
	if _, err := os.Stat("template/"); err == nil {
		t.Log("Found a template/ directory, removing it...")

		err := os.RemoveAll("template/")
		if err != nil {
			t.Log(err)
		}
	} else {
		t.Logf("Directory template was not created: %s", err)
	}
}
