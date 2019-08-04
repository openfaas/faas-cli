// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package commands

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"
)

func Test_templatePull(t *testing.T) {
	localTemplateRepository := setupLocalTemplateRepo(t)
	defer os.RemoveAll(localTemplateRepository)

	t.Run("ValidRepo", func(t *testing.T) {
		defer tearDownFetchTemplates(t)

		faasCmd.SetArgs([]string{"template", "pull", localTemplateRepository})
		err := faasCmd.Execute()
		if err != nil {
			t.Errorf("unexpected error while puling valid repo: %s", err.Error())
		}

		// Verify created directories
		if _, err := os.Stat("template"); err != nil {
			t.Fatalf("The directory %s was not created", "template")
		}
	})

	t.Run("WithOverwriting", func(t *testing.T) {
		defer tearDownFetchTemplates(t)

		faasCmd.SetArgs([]string{"template", "pull", localTemplateRepository})
		err := faasCmd.Execute()
		if err != nil {
			t.Errorf("unexpected error while executing template pull: %s", err.Error())
		}

		var buf bytes.Buffer
		log.SetOutput(&buf)

		r := regexp.MustCompile(`(?m:Cannot overwrite the following \d+ template\(s\):)`)

		faasCmd.SetArgs([]string{"template", "pull", localTemplateRepository})
		err = faasCmd.Execute()
		if err != nil {
			t.Errorf("unexpected error while executing template pull: %s", err.Error())
		}

		if !r.MatchString(buf.String()) {
			t.Fatal(buf.String())
		}

		buf.Reset()

		faasCmd.SetArgs([]string{"template", "pull", localTemplateRepository, "--overwrite"})
		err = faasCmd.Execute()
		if err != nil {
			t.Errorf("unexpected error while executing template pull with --overwrite: %s", err.Error())
		}

		str := buf.String()
		if r.MatchString(str) {
			t.Fatal()
		}

		// Verify created directories
		if _, err := os.Stat("template"); err != nil {
			t.Fatalf("The directory %s was not created", "template")
		}
	})

	t.Run("InvalidUrlError", func(t *testing.T) {
		faasCmd.SetArgs([]string{"template", "pull", "user@host.xz:openfaas/faas-cli.git"})
		err := faasCmd.Execute()
		if !strings.Contains(err.Error(), "The repository URL must be a valid git repo uri") {
			t.Errorf("Output does not contain the required error: %s", err.Error())
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
		t.Run(fmt.Sprintf("%s", scenario.name), func(t *testing.T) {
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
