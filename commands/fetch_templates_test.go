// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package commands

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Masterminds/semver"
	v2execute "github.com/alexellis/go-execute/v2"

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

	// exec "git version" to check which version of git is installed

	task := v2execute.ExecTask{
		Command: "git",
		Args:    []string{"version"},
	}

	res, err := task.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if res.ExitCode != 0 {
		t.Fatal("git version command failed")
	}

	_, v, _ := strings.Cut(strings.TrimSpace(res.Stdout), "git version ")

	// On darwin the string has extra text: "git version 2.39.2 (Apple Git-143)", so requires more trimming.
	if strings.Contains(v, " ") {
		v = strings.TrimSpace(v[:strings.Index(v, " ")])
	}

	s := semver.MustParse(v)
	initVersion := semver.MustParse("2.28.0")

	cmd := versioncontrol.GitInitRepoClassic
	if s.GreaterThan(initVersion) || s.Equal(initVersion) {
		cmd = versioncontrol.GitInitRepo2_28_0
	}

	if err := cmd.Invoke(dir, map[string]string{"dir": "."}); err != nil {
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
