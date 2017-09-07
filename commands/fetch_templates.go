// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"archive/zip"
	"errors"
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
	defaultTemplateRepository = "https://github.com/openfaas/faas-cli"
	templateDirectory         = "./template/"
)

// fetchTemplates fetch code templates from GitHub master zip file.
func fetchTemplates(templateURL string, overwrite bool) error {
	var existingLanguages []string
	availableLanguages := make(map[string]bool)
	var fetchedTemplates []string

	if len(templateURL) == 0 {
		templateURL = defaultTemplateRepository
	}

	archive, err := fetchMasterZip(templateURL)
	if err != nil {
		removeArchive(archive)
		return err
	}

	zipFile, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	log.Printf("Attempting to expand templates from %s\n", archive)

	for _, z := range zipFile.File {
		var rc io.ReadCloser

		relativePath := z.Name[strings.Index(z.Name, "/")+1:]
		if strings.Index(relativePath, "template/") != 0 {
			// Process only directories inside "template" at root
			continue
		}

		var language string
		if languageSplit := strings.Split(relativePath, "/"); len(languageSplit) > 2 {
			language = languageSplit[1]

			if len(languageSplit) == 3 && relativePath[len(relativePath)-1:] == "/" {
				// template/language/

				if !canWriteLanguage(&availableLanguages, language, overwrite) {
					existingLanguages = append(existingLanguages, language)
					continue
				}
				fetchedTemplates = append(fetchedTemplates, language)
			} else {
				// template/language/*

				if !canWriteLanguage(&availableLanguages, language, overwrite) {
					continue
				}
			}
		} else {
			// template/
			continue
		}

		if rc, err = z.Open(); err != nil {
			break
		}

		if err = createPath(relativePath, z.Mode()); err != nil {
			break
		}

		// If relativePath is just a directory, then skip expanding it.
		if len(relativePath) > 1 && relativePath[len(relativePath)-1:] != "/" {
			if err = writeFile(rc, z.UncompressedSize64, relativePath, z.Mode()); err != nil {
				break
			}
		}
	}

	if len(existingLanguages) > 0 {
		log.Printf("Cannot overwrite the following %d directories: %v\n", len(existingLanguages), existingLanguages)
	}

	zipFile.Close()

	log.Printf("Fetched %d template(s) : %v from %s\n", len(fetchedTemplates), fetchedTemplates, templateURL)

	err = removeArchive(archive)

	return err
}

// removeArchive removes the given file
func removeArchive(archive string) error {
	log.Printf("Cleaning up zip file...")
	if _, err := os.Stat(archive); err == nil {
		return os.Remove(archive)
	} else {
		return err
	}
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
			log.Println(err.Error())
			return "", err
		}
		log.Printf("HTTP GET %s\n", templateURL)
		res, err := client.Do(req)
		if err != nil {
			log.Println(err.Error())
			return "", err
		}
		if res.StatusCode != http.StatusOK {
			err := errors.New(fmt.Sprintf("%s is not valid, status code %d", templateURL, res.StatusCode))
			log.Println(err.Error())
			return "", err
		}
		if res.Body != nil {
			defer res.Body.Close()
		}
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println(err.Error())
			return "", err
		}

		log.Printf("Writing %dKb to %s\n", len(bytesOut)/1024, archive)
		err = ioutil.WriteFile(archive, bytesOut, 0700)
		if err != nil {
			log.Println(err.Error())
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

// Tells whether the language can be expanded from the zip or not
// availableLanguages map keeps track of which languages we know to be okay to copy.
// overwrite flag will allow to force copy the language template
func canWriteLanguage(availableLanguages *map[string]bool, language string, overwrite bool) bool {
	canWrite := false
	if len(language) > 0 {
		if _, ok := (*availableLanguages)[language]; ok {
			return (*availableLanguages)[language]
		}
		canWrite = checkLanguage(language, overwrite)
		(*availableLanguages)[language] = canWrite
	}

	return canWrite
}

// Takes a language input (e.g. "node"), tells whether or not it is OK to download
func checkLanguage(language string, overwrite bool) bool {
	dir := templateDirectory + language
	if _, err := os.Stat(dir); err == nil && !overwrite {
		// The directory template/language/ exists
		return false
	}
	return true
}
