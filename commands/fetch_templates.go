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
	defaultTemplateRepository = "https://github.com/openfaas/faas-cli"
	templateDirectory         = "./template/"
)

var cacheCanWriteLanguage = make(map[string]bool)

// fetchTemplates fetch code templates from GitHub master zip file.
func fetchTemplates(templateURL string, overwrite bool) error {
	var existingLanguages []string
	countFetchedTemplates := 0

	if len(templateURL) == 0 {
		templateURL = defaultTemplateRepository
	}

	archive, err := fetchMasterZip(templateURL)
	if err != nil {
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

				if !canWriteLanguage(language, overwrite) {
					existingLanguages = append(existingLanguages, language)
					continue
				}
				countFetchedTemplates++
			} else {
				// template/language/*

				if !canWriteLanguage(language, overwrite) {
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
		log.Printf("Cannot overwrite the following (%d) directories: %v\n", len(existingLanguages), existingLanguages)
	}

	zipFile.Close()

	log.Printf("Fetched %d template(s) from %s\n", countFetchedTemplates, templateURL)

	err = removeArchive(archive)

	return err
}

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

// canWriteLanguage tells whether the language can be processed or not
// if overwrite is activated, the directory template/language/ is removed before to keep it in sync
func canWriteLanguage(language string, overwrite bool) bool {

	if len(language) > 0 {
		if _, ok := cacheCanWriteLanguage[language]; ok {
			return cacheCanWriteLanguage[language]
		}

		dir := templateDirectory + language
		if _, err := os.Stat(dir); err == nil {
			// The directory template/language/ exists
			if overwrite == false {
				cacheCanWriteLanguage[language] = false
			} else {
				// Clean up the directory to keep in sync with new updates
				if err := os.RemoveAll(dir); err != nil {
					log.Panicf("Directory %s cannot be removed", dir)
				}
				cacheCanWriteLanguage[language] = true
			}
		} else {
			cacheCanWriteLanguage[language] = true
		}

		return cacheCanWriteLanguage[language]
	}

	return false
}
