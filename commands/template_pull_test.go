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
		faasCmd.Execute()

		// Verify created directories
		if _, err := os.Stat("template"); err != nil {
			t.Fatalf("The directory %s was not created", "template")
		}
	})

	t.Run("WithOverwriting", func(t *testing.T) {
		defer tearDownFetchTemplates(t)

		faasCmd.SetArgs([]string{"template", "pull", localTemplateRepository})
		faasCmd.Execute()

		var buf bytes.Buffer
		log.SetOutput(&buf)

		r := regexp.MustCompile(`(?m:Cannot overwrite the following \d+ template\(s\):)`)

		faasCmd.SetArgs([]string{"template", "pull", localTemplateRepository})
		faasCmd.Execute()

		if !r.MatchString(buf.String()) {
			t.Fatal(buf.String())
		}

		buf.Reset()

		faasCmd.SetArgs([]string{"template", "pull", localTemplateRepository, "--overwrite"})
		faasCmd.Execute()

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
		var buf bytes.Buffer

		faasCmd.SetArgs([]string{"template", "pull", "user@host.xz:openfaas/faas-cli.git"})
		faasCmd.SetOutput(&buf)
		err := faasCmd.Execute()

		if !strings.Contains(err.Error(), "The repository URL must be a valid git repo uri") {
			t.Fatal("Output does not contain the required string", err.Error())
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
		{name: "git protocol with branch", url: "git://github.com/openfaas/faas.git#master"},
		{name: "git protocol", url: "git://host.xz/path/to/repo.git/"},
		{name: "scp style with ip address", url: "git@192.168.101.127:user/project.git"},
		{name: "scp style with hostname", url: "git@github.com:user/project.git"},
		{name: "http protocol with ip address", url: "http://192.168.101.127/user/project.git"},
		{name: "http protocol", url: "http://github.com/user/project.git"},
		{name: "https protocol with ip address", url: "https://192.168.101.127/user/project.git"},
		{name: "https protocol with hostname", url: "https://github.com/user/project.git"},
		{name: "https protocol with basic auth", url: "https://username:password@github.com/username/repository.git"},
		{name: "ssh protocol with hostname no port", url: "ssh://user@host.xz/path/to/repo.git/"},
		{name: "ssh protocol with hostname and port", url: "ssh://user@host.xz:port/path/to/repo.git/"},
	}

	for _, scenario := range validURLs {
		t.Run(fmt.Sprintf("%s is a valid remote git url", scenario.name), func(t *testing.T) {
			if !r.MatchString(scenario.url) {
				t.Errorf("Url %s should pass the regex match", scenario.url)
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
				t.Errorf("Url %s should fail the regex match", scenario.url)
			}

		})
	}
}
