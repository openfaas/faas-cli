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
			fmt.Errorf("unexpected error while executing template pull with --overwrite: %s", err.Error())
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

func Test_repositoryUrlRemoteRegExp(t *testing.T) {

	r := regexp.MustCompile(gitRemoteRepoRegex)
	validURLs := []struct {
		name string
		url  string
	}{
		{name: "git protocol with sha", url: "git://github.com/openfaas/faas.git#ff78lf9h"},
		{name: "git protocol without .git suffix", url: "git://host.xz/path/to/repo"},
		{name: "git protocol with branch", url: "git://github.com/openfaas/faas.git#master"},
		{name: "git protocol", url: "git://host.xz/path/to/repo.git/"},
		{name: "scp style with ip address", url: "git@192.168.101.127:user/project.git"},
		{name: "scp style with hostname", url: "git@github.com:user/project.git"},
		{name: "http protocol with ip address", url: "http://192.168.101.127/user/project.git"},
		{name: "http protocol", url: "http://github.com/user/project.git"},
		{name: "http protocol without .git suffix", url: "http://github.com/user/project"},
		{name: "https protocol with ip address", url: "https://192.168.101.127/user/project.git"},
		{name: "https protocol with hostname", url: "https://github.com/user/project.git"},
		{name: "https protocol with basic auth", url: "https://username:password@github.com/username/repository.git"},
		{name: "ssh protocol with hostname no port", url: "ssh://user@host.xz/path/to/repo.git/"},
		{name: "ssh protocol with hostname and port", url: "ssh://user@host.xz:port/path/to/repo.git/"},
	}

	for _, scenario := range validURLs {
		t.Run(fmt.Sprintf("%s is a valid remote git url", scenario.name), func(t *testing.T) {
			if !r.MatchString(scenario.url) {
				t.Errorf("Url %s should pass the regex %s", scenario.url, gitRemoteRepoRegex)
			}

		})
	}

	invalidURLs := []struct {
		name string
		url  string
	}{
		{name: "local repo file protocol", url: "file:///path/to/repo.git/"},
		{name: "ssh missing username and port", url: "host.xz:/path/to/repo.git"},
		{name: "ssh username and missing port", url: "user@host.xz:path/to/repo.git"},
		{name: "relative local path", url: "path/to/repo.git/"},
		{name: "magic relative local", url: "~/path/to/repo.git"},
	}
	for _, scenario := range invalidURLs {
		t.Run(fmt.Sprintf("%s is not a valid remote git url", scenario.name), func(t *testing.T) {
			if r.MatchString(scenario.url) {
				t.Errorf("Url %s should fail the regex %s", scenario.url, gitRemoteRepoRegex)
			}

		})
	}
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
