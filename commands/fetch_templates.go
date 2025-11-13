// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	execute "github.com/alexellis/go-execute/v2"
	"github.com/openfaas/faas-cli/builder"
	"github.com/openfaas/faas-cli/versioncontrol"
)

// DefaultTemplateRepository contains the Git repo for the official templates
const DefaultTemplateRepository = "https://github.com/openfaas/templates.git"

const TemplateDirectory = "./template/"

const ShaPrefix = "sha-"

// fetchTemplates fetch code templates using git clone.
func fetchTemplates(templateURL, refName, templateName string, overwriteTemplates bool) error {
	if len(templateURL) == 0 {
		return fmt.Errorf("pass valid templateURL")
	}

	refMsg := ""
	if len(refName) > 0 {
		refMsg = " [" + refName + "]"
	}

	log.Printf("Fetching templates from %s%s", templateURL, refMsg)

	extractedPath, err := os.MkdirTemp("", "openfaas-templates-*")
	if err != nil {
		return fmt.Errorf("unable to create temporary directory: %s", err)
	}

	if !pullDebug {
		defer os.RemoveAll(extractedPath)
	}

	pullDebugPrint(fmt.Sprintf("Temp files in %s", extractedPath))

	args := map[string]string{"dir": extractedPath, "repo": templateURL}
	cmd := versioncontrol.GitCloneDefault

	if len(refName) > 0 {
		if strings.HasPrefix(refName, ShaPrefix) {
			cmd = versioncontrol.GitCloneFullDepth
		} else {
			args["refname"] = refName

			cmd = versioncontrol.GitCloneBranch
			args["refname"] = refName
		}
	}

	if err := cmd.Invoke(".", args); err != nil {
		return fmt.Errorf("error invoking git clone: %w", err)
	}

	if len(refName) > 0 && strings.HasPrefix(refName, ShaPrefix) {

		targetCommit := strings.TrimPrefix(refName, ShaPrefix)

		if !regexp.MustCompile(`^[a-fA-F0-9]{7,40}$`).MatchString(targetCommit) {
			return fmt.Errorf("invalid SHA format: %s - must be 7-40 hex characters", targetCommit)
		}

		t := execute.ExecTask{
			Command: "git",
			Args:    []string{"-C", extractedPath, "checkout", targetCommit},
		}
		res, err := t.Execute(context.Background())
		if err != nil {
			return fmt.Errorf("error checking out ref %s: %w", targetCommit, err)
		}
		if res.ExitCode != 0 {
			out := res.Stdout + " " + res.Stderr
			return fmt.Errorf("error checking out ref %s: %s", targetCommit, out)
		}
	}

	if os.Getenv("FAAS_DEBUG") == "1" {
		task := execute.ExecTask{
			Command: "git",
			Args:    []string{"-C", extractedPath, "log", "-1", "--oneline"},
		}

		res, err := task.Execute(context.Background())
		if err != nil {
			return fmt.Errorf("error executing git log: %w", err)
		}
		if res.ExitCode != 0 {
			e := fmt.Errorf("exit code: %d, stderr: %s, stdout: %s", res.ExitCode, res.Stderr, res.Stdout)
			return fmt.Errorf("error from: git log: %w", e)
		}

		log.Printf("[git] log: %s", strings.TrimSpace(res.Stdout))
	}

	// Get the long SHA digest from the clone repository.
	sha, err := versioncontrol.GetGitSHAFor(extractedPath, false)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("can't get current working directory: %s", err)
	}
	localTemplatesDir := filepath.Join(cwd, TemplateDirectory)
	if _, err := os.Stat(localTemplatesDir); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(localTemplatesDir, 0755); err != nil && !os.IsExist(err) {
			return fmt.Errorf("error creating template directory: %s - %w", localTemplatesDir, err)
		}
	}

	protectedLanguages, fetchedLanguages, err := moveTemplates(localTemplatesDir, extractedPath, templateName, overwriteTemplates, templateURL, refName, sha)
	if err != nil {
		return err
	}

	if len(protectedLanguages) > 0 {
		return fmt.Errorf("unable to overwrite the following: %v", protectedLanguages)
	}

	fmt.Printf("Wrote %d template(s) : %v\n", len(fetchedLanguages), fetchedLanguages)

	return err
}

// canWriteLanguage tells whether the language can be expanded from the zip or not.
// availableLanguages map keeps track of which languages we know to be okay to copy.
// overwrite flag will allow to force copy the language template
func canWriteLanguage(existingLanguages []string, language string, overwriteTemplate bool) bool {
	if overwriteTemplate {
		return true
	}

	return !slices.Contains(existingLanguages, language)
}

// moveTemplates moves the templates from the repository to the template directory
// It returns the existing languages and the fetched languages
// It also returns an error if the templates cannot be read
func moveTemplates(localTemplatesDir, extractedPath, templateName string, overwriteTemplate bool, repository string, refName string, sha string) ([]string, []string, error) {

	var (
		existingLanguages  []string
		fetchedLanguages   []string
		protectedLanguages []string
		err                error
	)

	templateEntries, err := os.ReadDir(localTemplatesDir)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read directory: %s", localTemplatesDir)
	}

	// OK if nothing exists yet
	for _, entry := range templateEntries {
		if !entry.IsDir() {
			continue
		}

		templateFile := filepath.Join(localTemplatesDir, entry.Name(), "template.yml")
		if _, err := os.Stat(templateFile); err != nil && !os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("can't find template.yml in: %s", templateFile)
		}

		existingLanguages = append(existingLanguages, entry.Name())
	}

	extractedTemplates, err := os.ReadDir(filepath.Join(extractedPath, TemplateDirectory))
	if err != nil {
		return nil, nil, fmt.Errorf("can't find templates in: %s", filepath.Join(extractedPath, TemplateDirectory))
	}

	for _, entry := range extractedTemplates {
		if !entry.IsDir() {
			continue
		}
		language := entry.Name()
		refSuffix := ""
		if refName != "" {
			refSuffix = "@" + refName
		}

		if canWriteLanguage(existingLanguages, language, overwriteTemplate) {
			// Do cp here
			languageSrc := filepath.Join(extractedPath, TemplateDirectory, language)
			languageDest := filepath.Join(localTemplatesDir, language)
			langName := language
			if refName != "" {
				languageDest += "@" + refName
				langName = language + "@" + refName
			}
			fetchedLanguages = append(fetchedLanguages, langName)

			if err := builder.CopyFiles(languageSrc, languageDest); err != nil {
				return nil, nil, err
			}

			if err := writeTemplateMeta(languageDest, repository, refName, sha); err != nil {
				return nil, nil, err
			}
		} else {
			protectedLanguages = append(protectedLanguages, language+refSuffix)
			continue
		}
	}

	return protectedLanguages, fetchedLanguages, nil
}

func writeTemplateMeta(languageDest, repository, refName, sha string) error {
	templateMeta := TemplateMeta{
		Repository: repository,
		WrittenAt:  time.Now(),
		RefName:    refName,
		Sha:        sha,
	}

	metaBytes, err := json.Marshal(templateMeta)
	if err != nil {
		return fmt.Errorf("error marshalling template meta: %s", err)
	}

	metaPath := filepath.Join(languageDest, "meta.json")
	if err := os.WriteFile(metaPath, metaBytes, 0644); err != nil {
		return fmt.Errorf("error writing template meta: %s", err)
	}

	return nil
}

func pullTemplate(repository, templateName string, overwriteTemplates bool) error {

	baseRepository := repository

	// Sometimes a templates git repo can be a local path
	if _, err := os.Stat(repository); err != nil && os.IsNotExist(err) {
		base, _, found := strings.Cut(repository, "#")
		if found {
			baseRepository = base
		} else {

			_, ref, found := strings.Cut(templateName, "#")
			if found {
				repository = baseRepository + "#" + ref
			}
		}
	}

	if !isValidFilesystemPath(repository) {
		if !versioncontrol.IsGitRemote(baseRepository) && !versioncontrol.IsPinnedGitRemote(baseRepository) {
			return fmt.Errorf("the repository URL must be a valid git repo uri")
		}
	}

	repository, refName := versioncontrol.ParsePinnedRemote(repository)
	isShaRefName := strings.HasPrefix(refName, ShaPrefix)
	if refName != "" && !isShaRefName {
		err := versioncontrol.GitCheckRefName.Invoke("", map[string]string{"refname": refName})
		if err != nil {
			fmt.Printf("Invalid tag or branch name `%s`\n", refName)
			fmt.Println("See https://git-scm.com/docs/git-check-ref-format for more details of the rules Git enforces on branch and reference names.")

			return err
		}
	}

	if err := fetchTemplates(repository, refName, templateName, overwriteTemplates); err != nil {
		return fmt.Errorf("error while fetching templates: %w", err)
	}

	return nil
}

type TemplateMeta struct {
	Repository string    `json:"repository"`
	RefName    string    `json:"ref_name,omitempty"`
	Sha        string    `json:"sha,omitempty"`
	WrittenAt  time.Time `json:"written_at"`
}

func isValidFilesystemPath(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
