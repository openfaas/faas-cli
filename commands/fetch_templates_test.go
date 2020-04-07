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
	localTemplateRepository := setupLocalTemplateRepo(t, "")
	defer os.RemoveAll(localTemplateRepository)

	const templatePath = "path/to/template"
	nestedTemplateRepository := setupLocalTemplateRepo(t, templatePath)
	defer os.RemoveAll(nestedTemplateRepository)

	defer tearDownFetchTemplates(t)

	t.Run("simplePull", func(t *testing.T) {
		defer tearDownFetchTemplates(t)
		if err := PullTemplates(localTemplateRepository); err != nil {
			t.Error(err)
		}
	})

	t.Run("nestedPull", func(t *testing.T) {
		defer tearDownFetchTemplates(t)
		if err := PullTemplatesPath(nestedTemplateRepository, templatePath); err != nil {
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

	t.Run("fetchTemplatesPath", func(t *testing.T) {
		defer tearDownFetchTemplates(t)

		err := fetchTemplatesPath(nestedTemplateRepository, "master", templatePath, false)
		if err != nil {
			t.Error(err)
		}
	})
}

// setupLocalTemplateRepo will create a local copy of the core OpenFaaS templates, this
// can be referred to as a local git repository.
// If path is a non empty relative path, the templates will be created in a directory
// nested under the returned repo path
func setupLocalTemplateRepo(t *testing.T, path string) string {
	tdir, err := ioutil.TempDir("", "openFaasTestTemplates")
	if err != nil {
		t.Error(err)
	}

	dir := filepath.Join(tdir, path)
	if err = os.MkdirAll(dir, 0700); err != nil {
		t.Error(err)
	}

	// Copy the submodule to temp directory to avoid altering it during tests
	testRepoGit := filepath.Join("testdata", "templates")
	builder.CopyFiles(testRepoGit, dir)
	// Remove submodule .git file
	os.Remove(filepath.Join(dir, ".git"))
	if err := versioncontrol.GitInitRepo.Invoke(tdir, map[string]string{"dir": "."}); err != nil {
		t.Fatal(err)
	}

	return tdir
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
