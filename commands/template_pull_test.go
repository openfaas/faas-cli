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
	validURLs := []string{
		"git://github.com/openfaas/faas.git#ff78lf9h",
		"git://github.com/openfaas/faas.git#gh-pages",
		"git://github.com/openfaas/faas.git#master",
		"git://github.com/openfaas/faas.git#quick_fix",
		"git://github.com/openfaas/faas.git#v0.1.0",
		"git://host.xz/path/to/repo.git/",
		"git@192.168.101.127:user/project.git",
		"git@github.com:user/project.git",
		"http://192.168.101.127/user/project.git",
		"http://github.com/user/project.git",
		"http://host.xz/path/to/repo.git/",
		"https://192.168.101.127/user/project.git",
		"https://github.com/user/project.git",
		"https://host.xz/path/to/repo.git/",
		"https://username:password@github.com/username/repository.git",
		"ssh://user@host.xz/path/to/repo.git/",
		"ssh://user@host.xz:port/path/to/repo.git/",
	}

	for _, url := range validURLs {
		t.Run(fmt.Sprintf("%s is a valid remote git url", url), func(t *testing.T) {
			if !r.MatchString(url) {
				t.Errorf("Url %s should pass the regex match", url)
			}

		})
	}

	invalidURLs := []string{
		"file:///path/to/repo.git/",
		"host.xz:/path/to/repo.git",
		"host.xz:path/to/repo.git",
		"user@host.xz:path/to/repo.git",
		"path/to/repo.git/",
		"~/path/to/repo.git",
	}
	for _, url := range invalidURLs {
		t.Run(fmt.Sprintf("%s is not a valid remote git url", url), func(t *testing.T) {
			if r.MatchString(url) {
				t.Errorf("Url %s should fail the regex match", url)
			}

		})
	}
}
