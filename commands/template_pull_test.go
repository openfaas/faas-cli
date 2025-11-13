// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func Test_templatePull(t *testing.T) {
	localTemplateRepository := setupLocalTemplateRepo(t)
	defer os.RemoveAll(localTemplateRepository)

	t.Run("ValidRepo", func(t *testing.T) {
		defer tearDownFetchTemplates(t)

		faasCmd.SetArgs([]string{"template", "pull", localTemplateRepository})
		if err := faasCmd.Execute(); err != nil {
			t.Fatalf("unexpected error while puling valid repo (%q): %s", localTemplateRepository, err.Error())
		}

		// Verify created directories
		if _, err := os.Stat("template"); err != nil {
			t.Fatalf("The directory %s was not created", "template")
		}

	})

	t.Run("WithOverwriting", func(t *testing.T) {
		defer tearDownFetchTemplates(t)

		if err := templatePull(localTemplateRepository, true); err != nil {
			t.Fatalf("unexpected error while executing initial template pull: %s", err.Error())
		}

		//execute command to ls in the template directory

		lsCmd := exec.Command("ls", "template")
		_, err := lsCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("error while listing template directory: %s", err.Error())
		}

		if err := templatePull(localTemplateRepository, false); err == nil {
			t.Fatalf("error expected overwriting existing templates with --overwrite=false:")
		}

		if err := templatePull(localTemplateRepository, true); err != nil {
			t.Fatalf("unexpected error while executing template pull with --overwrite: %s", err.Error())
		}

		// Verify created directories
		if _, err := os.Stat("template"); err != nil {
			t.Fatalf("The directory %s was not created", "template")
		}
	})

	t.Run("InvalidUrlError", func(t *testing.T) {
		faasCmd.SetArgs([]string{"template", "pull", "user@host.xz:openfaas/faas-cli.git"})
		err := faasCmd.Execute()
		want := "the repository URL must be a valid git repo uri"
		got := err.Error()
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("The error should contain:\n%q\n, but was:\n%q", want, got)
		}
	})
}

func Test_templatePullPriority(t *testing.T) {
	templateURLs := []struct {
		name      string
		envURL    string
		cliURL    string
		resultURL string
	}{
		{
			name:      "Use Default URL when none provided",
			resultURL: DefaultTemplateRepository,
		},
		{
			name:      "Use Env URL when only env provided",
			envURL:    "https://github.com/user/project.git",
			resultURL: "https://github.com/user/project.git",
		},
		{
			name:      "Use Cli URL when only cli provided",
			cliURL:    "git@github.com:user/project.git",
			resultURL: "git@github.com:user/project.git",
		},
		{
			name:      "Use Cli URL when both cli and env provided",
			cliURL:    "git@github.com:user/project.git",
			envURL:    "https://github.com/user/project.git",
			resultURL: "git@github.com:user/project.git",
		},
	}
	for _, scenario := range templateURLs {
		t.Run(scenario.name, func(t *testing.T) {
			repository = getTemplateURL(scenario.cliURL, scenario.envURL, DefaultTemplateRepository)
			if repository != scenario.resultURL {
				t.Errorf("result URL,  want %s got %s", scenario.resultURL, repository)
			}

		})

	}
}

// templatePullLocalTemplateRepo executes `template pull` on a local repository to get templates
func templatePullLocalTemplateRepo(t *testing.T) {
	localTemplateRepository := setupLocalTemplateRepo(t)
	defer os.RemoveAll(localTemplateRepository)

	faasCmd.SetArgs([]string{"template", "pull", localTemplateRepository})
	err := faasCmd.Execute()
	if err != nil {
		fmt.Printf("error while executing template pull: %s", err.Error())
	}
}
