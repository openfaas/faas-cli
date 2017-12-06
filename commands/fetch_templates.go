// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openfaas/faas-cli/proxy"
)

const (
	defaultTemplateRepository = "https://github.com/openfaas/templates"
	templateDirectory         = "./template/"
	rootLanguageDirSplitCount = 3
)

type extractAction int

const (
	shouldExtractData extractAction = iota
	newTemplateFound
	directoryAlreadyExists
	skipWritingData
)

// fetchTemplates fetch code templates from GitHub master zip file.
func fetchTemplates(templateURL string, overwrite bool) error {

	if len(templateURL) == 0 {
		templateURL = defaultTemplateRepository
	}

	archive, err := fetchMasterZip(templateURL)
	if err != nil {
		removeArchive(archive)
		return err
	}

	log.Printf("Attempting to expand templates from %s\n", archive)

	preExistingLanguages, fetchedLanguages, err := expandTemplatesFromZip(archive, overwrite)
	if err != nil {
		return err
	}

	if len(preExistingLanguages) > 0 {
		log.Printf("Cannot overwrite the following %d directories: %v\n", len(preExistingLanguages), preExistingLanguages)
	}

	log.Printf("Fetched %d template(s) : %v from %s\n", len(fetchedLanguages), fetchedLanguages, templateURL)

	err = removeArchive(archive)

	return err
}

// expandTemplatesFromZip builds a list of languages that: already exist and
// could not be overwritten and // a list of languages that are newly downloaded.
func expandTemplatesFromZip(archivePath string, overwrite bool) ([]string, []string, error) {
	var existingLanguages []string
	var fetchedLanguages []string

	availableLanguages := make(map[string]bool)

	zipFile, err := zip.OpenReader(archivePath)

	if err != nil {
		return nil, nil, err
	}
	defer zipFile.Close()

	for _, z := range zipFile.File {

		relativePath := z.Name[strings.Index(z.Name, "/")+1:]
		if strings.Index(relativePath, "template/") != 0 {
			// Process only directories inside "template" at root
			continue
		}

		action, language, isDirectory := canExpandTemplateData(availableLanguages, relativePath)

		var expandFromZip bool

		switch action {

		case shouldExtractData:
			expandFromZip = true
		case newTemplateFound:
			expandFromZip = true
			fetchedLanguages = append(fetchedLanguages, language)
		case directoryAlreadyExists:
			expandFromZip = false
			existingLanguages = append(existingLanguages, language)
		case skipWritingData:
			expandFromZip = false
		default:
			return nil, nil, fmt.Errorf(fmt.Sprintf("don't know what to do when extracting zip: %s", archivePath))
		}

		if expandFromZip {
			var rc io.ReadCloser

			if rc, err = z.Open(); err != nil {
				break
			}
			defer rc.Close()

			if err = createPath(relativePath, z.Mode()); err != nil {
				break
			}

			// If relativePath is just a directory, then skip expanding it.
			if len(relativePath) > 1 && !isDirectory {
				if err = writeFile(rc, z.UncompressedSize64, relativePath, z.Mode()); err != nil {
					return nil, nil, err
				}
			}
		}
	}

	return existingLanguages, fetchedLanguages, nil
}

// canExpandTemplateData returns what we should do with the file in form of ExtractAction enum
// with the language name and whether it is a directory
func canExpandTemplateData(availableLanguages map[string]bool, relativePath string) (extractAction, string, bool) {
	if pathSplit := strings.Split(relativePath, "/"); len(pathSplit) > 2 {
		language := pathSplit[1]

		// We know that this path is a directory if the last character is a "/"
		isDirectory := strings.HasSuffix(relativePath, "/")

		// Check if this is the root directory for a language (at ./template/lang)
		if len(pathSplit) == rootLanguageDirSplitCount && isDirectory {
			if !canWriteLanguage(availableLanguages, language, overwrite) {
				return directoryAlreadyExists, language, isDirectory
			}
			return newTemplateFound, language, isDirectory
		}

		if canWriteLanguage(availableLanguages, language, overwrite) == false {
			return skipWritingData, language, isDirectory
		}

		return shouldExtractData, language, isDirectory
	}
	// template/
	return skipWritingData, "", true
}

// removeArchive removes the given file
func removeArchive(archive string) error {
	log.Printf("Cleaning up zip file...")
	var err error

	if _, err = os.Stat(archive); err == nil {
		err = os.Remove(archive)
	}

	return err
}

// fetchMasterZip downloads a zip file from a repository URL
func fetchMasterZip(templateURL string) (string, error) {
	var err error

	templateURL = strings.TrimRight(templateURL, "/")
	templateURL = templateURL + "/archive/master.zip"
	archive := "master.zip"

	if _, err := os.Stat(archive); err != nil {
		timeout := 120 * time.Second
		client := proxy.MakeHTTPClient(&timeout)

		req, err := http.NewRequest(http.MethodGet, templateURL, nil)
		if err != nil {
			return "", err
		}

		log.Printf("HTTP GET %s\n", templateURL)
		res, err := client.Do(req)
		if err != nil {
			return "", err
		}

		if res.StatusCode != http.StatusOK {
			err := fmt.Errorf(fmt.Sprintf("%s is not valid, status code %d", templateURL, res.StatusCode))
			log.Println(err.Error())
			return "", err
		}

		if res.Body != nil {
			defer res.Body.Close()
		}

		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", err
		}

		log.Printf("Writing %dKb to %s\n", len(bytesOut)/1024, archive)
		err = ioutil.WriteFile(archive, bytesOut, 0700)
		if err != nil {
			return "", err
		}
	}

	fmt.Println("")
	return archive, err
}

func writeFile(rc io.ReadCloser, size uint64, relativePath string, perms os.FileMode) error {
	var err error

	defer rc.Close()
	f, err := os.OpenFile(relativePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perms)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.CopyN(f, rc, int64(size))

	return err
}

func createPath(relativePath string, perms os.FileMode) error {
	dir := filepath.Dir(relativePath)
	err := os.MkdirAll(dir, perms)
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
