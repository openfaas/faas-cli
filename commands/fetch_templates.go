// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/openfaas/faas-cli/builder"
	"github.com/openfaas/faas-cli/versioncontrol"
)

// DefaultTemplateRepository contains the Git repo for the official templates
const DefaultTemplateRepository = "https://github.com/openfaas/templates.git"

const templateDirectory = "./template/"

// fetchTemplates fetch code templates using git clone.
func fetchTemplates(templateURL, refName, templateName string, overwrite bool) error {
	if len(templateURL) == 0 {
		return fmt.Errorf("pass valid templateURL")
	}

	dir, err := os.MkdirTemp("", "openfaas-templates-*")
	if err != nil {
		return err
	}

	if !pullDebug {
		defer os.RemoveAll(dir)
	}

	pullDebugPrint(fmt.Sprintf("Temp files in %s", dir))

	args := map[string]string{"dir": dir, "repo": templateURL}
	cmd := versioncontrol.GitCloneDefault

	if refName != "" {
		args["refname"] = refName
		cmd = versioncontrol.GitClone
	}

	if err := cmd.Invoke(".", args); err != nil {
		return err
	}

	preExistingLanguages, fetchedLanguages, err := moveTemplates(dir, templateName, overwrite)
	if err != nil {
		return err
	}

	if len(preExistingLanguages) > 0 {
		log.Printf("Cannot overwrite the following %d template(s): %v\n", len(preExistingLanguages), preExistingLanguages)
	}

	fmt.Printf("Wrote %d template(s) : %v from %s\n", len(fetchedLanguages), fetchedLanguages, templateURL)

	return err
}

// canWriteLanguage tells whether the language can be expanded from the zip or not.
// availableLanguages map keeps track of which languages we know to be okay to copy.
// overwrite flag will allow to force copy the language template
func canWriteLanguage(availableLanguages map[string]bool, language string, overwrite bool) bool {
	canWrite := false
	if availableLanguages != nil && len(language) > 0 {
		if _, found := availableLanguages[language]; found {
			return availableLanguages[language]
		}
		canWrite = templateFolderExists(language, overwrite)
		availableLanguages[language] = canWrite
	}

	return canWrite
}

// Takes a language input (e.g. "node"), tells whether or not it is OK to download
func templateFolderExists(language string, overwrite bool) bool {
	dir := templateDirectory + language
	if _, err := os.Stat(dir); err == nil && !overwrite {
		// The directory template/language/ exists
		return false
	}
	return true
}

func moveTemplates(repoPath, templateName string, overwrite bool) ([]string, []string, error) {
	var (
		existingLanguages []string
		fetchedLanguages  []string
		err               error
	)

	availableLanguages := make(map[string]bool)

	templateDir := filepath.Join(repoPath, templateDirectory)
	templates, err := os.ReadDir(templateDir)
	if err != nil {
		return nil, nil, fmt.Errorf("can't find templates in: %s", repoPath)
	}

	for _, file := range templates {
		if !file.IsDir() {
			continue
		}
		language := file.Name()

		if len(templateName) == 0 {

			canWrite := canWriteLanguage(availableLanguages, language, overwrite)
			if canWrite {
				fetchedLanguages = append(fetchedLanguages, language)
				// Do cp here
				languageSrc := filepath.Join(templateDir, language)
				languageDest := filepath.Join(templateDirectory, language)
				builder.CopyFiles(languageSrc, languageDest)
			} else {
				existingLanguages = append(existingLanguages, language)
				continue
			}
		} else if language == templateName {

			if canWriteLanguage(availableLanguages, language, overwrite) {
				fetchedLanguages = append(fetchedLanguages, language)
				// Do cp here
				languageSrc := filepath.Join(templateDir, language)
				languageDest := filepath.Join(templateDirectory, language)
				builder.CopyFiles(languageSrc, languageDest)
			}
		}
	}

	return existingLanguages, fetchedLanguages, nil
}

func pullTemplate(repository, templateName string) error {
	if _, err := os.Stat(repository); err != nil {
		if !versioncontrol.IsGitRemote(repository) && !versioncontrol.IsPinnedGitRemote(repository) {
			return fmt.Errorf("the repository URL must be a valid git repo uri")
		}
	}

	repository, refName := versioncontrol.ParsePinnedRemote(repository)

	if refName != "" {
		err := versioncontrol.GitCheckRefName.Invoke("", map[string]string{"refname": refName})
		if err != nil {
			fmt.Printf("Invalid tag or branch name `%s`\n", refName)
			fmt.Println("See https://git-scm.com/docs/git-check-ref-format for more details of the rules Git enforces on branch and reference names.")

			return err
		}
	}

	refStr := ""
	if len(refName) > 0 {
		refStr = fmt.Sprintf(" @ %s", refName)
	}
	fmt.Printf("Fetch templates from repository: %s%s\n", repository, refStr)
	if err := fetchTemplates(repository, refName, templateName, overwrite); err != nil {
		return fmt.Errorf("error while fetching templates: %s", err)
	}

	return nil
}
