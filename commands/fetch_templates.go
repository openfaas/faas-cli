// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/openfaas/faas-cli/vcs"
)

const (
	defaultTemplateRepository = "https://github.com/openfaas/templates.git"
	templateDirectory         = "./template/"
)

// fetchTemplates fetch code templates from GitHub master zip file.
func fetchTemplates(templateURL string, overwrite bool) error {
	if len(templateURL) == 0 {
		templateURL = defaultTemplateRepository
	}

	dir, err := ioutil.TempDir("", "openFaasTemplates")
	if err != nil {
		log.Fatal(err)
	}
	if !debug {
		defer os.RemoveAll(dir) // clean up
	}

	log.Printf("Attempting to expand templates from %s\n", templateURL)
	debugPrint(fmt.Sprintf("Temp files in %s", dir))
	if err := vcs.Git.Create(dir, templateURL); err != nil {
		return err
	}

	preExistingLanguages, fetchedLanguages, err := moveTemplates(dir, overwrite)
	if err != nil {
		return err
	}

	if len(preExistingLanguages) > 0 {
		log.Printf("Cannot overwrite the following %d template(s): %v\n", len(preExistingLanguages), preExistingLanguages)
	}

	log.Printf("Fetched %d template(s) : %v from %s\n", len(fetchedLanguages), fetchedLanguages, templateURL)

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

func moveTemplates(repoPath string, overwrite bool) ([]string, []string, error) {
	var (
		existingLanguages []string
		fetchedLanguages  []string
		err               error
	)

	availableLanguages := make(map[string]bool)

	templateDir := filepath.Join(repoPath, templateDirectory)
	templates, err := ioutil.ReadDir(templateDir)
	if err != nil {
		return nil, nil, fmt.Errorf("can't find templates in: %s", repoPath)
	}

	for _, file := range templates {
		if !file.IsDir() {
			continue
		}
		language := file.Name()

		canWrite := canWriteLanguage(availableLanguages, language, overwrite)
		if canWrite {
			fetchedLanguages = append(fetchedLanguages, language)
			// Do cp here
			languageSrc := filepath.Join(templateDir, language)
			languageDest := filepath.Join(templateDirectory, language)
			copy(languageSrc, languageDest, file)
		} else {
			existingLanguages = append(existingLanguages, language)
			continue
		}
	}

	return existingLanguages, fetchedLanguages, nil
}

// "recursively copy a file object, info must be non-nil
func copy(src, dest string, info os.FileInfo) error {
	if info.IsDir() {
		return dcopy(src, dest, info)
	}
	return fcopy(src, dest, info)
}

// fcopy will copy a file with the same mode as the src file
func fcopy(src, dest string, info os.FileInfo) error {

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = os.Chmod(f.Name(), info.Mode()); err != nil {
		return err
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	_, err = io.Copy(f, s)
	return err
}

// dcopy will recursively copy a directory to dest
func dcopy(src, dest string, info os.FileInfo) error {

	if err := os.MkdirAll(dest, info.Mode()); err != nil {
		return err
	}

	infos, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, info := range infos {
		if err := copy(
			filepath.Join(src, info.Name()),
			filepath.Join(dest, info.Name()),
			info,
		); err != nil {
			return err
		}
	}

	return nil
}
