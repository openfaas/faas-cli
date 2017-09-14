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
)

const defaultTemplateRepository = "https://github.com/alexellis/faas-cli/archive/master.zip"
const cacheDirectory = "./.cache/"
const templateDirectory = "./template/"

var cache = make(map[string]bool)

// fetchTemplates fetch code templates from GitHub master zip file.
func fetchTemplates(templateURL string, keepArchive bool, overwrite bool) error {
	if len(templateURL) == 0 {
		templateURL = os.Getenv("templateUrl")
		if len(templateURL) == 0 {
			templateURL = defaultTemplateRepository
		}
	} else {
		templateURL = templateURL + "/archive/master.zip"
	}
	archive, err := fetchMasterZip(templateURL)

	zipFile, err := zip.OpenReader("./" + archive)
	if err != nil {
		return err
	}

	log.Printf("Attempting to expand templates from %s\n", archive)

	for _, z := range zipFile.File {
		var rc io.ReadCloser

		relativePath := z.Name[strings.Index(z.Name, "/")+1:]

		var language string
		if languageSplit := strings.Split(relativePath, "/"); len(languageSplit) > 1 {
			language = languageSplit[1]
		} else {
			continue
		}

		if !canWriteLanguage(language, overwrite) {
			continue
		}

		if strings.Index(relativePath, "template") == 0 {
			fmt.Printf("Found \"%s\"\n", relativePath)
			if rc, err = z.Open(); err != nil {
				break
			}

			if err = createPath(relativePath, z.Mode()); err != nil {
				break
			}

			// If relativePath is just a directory, then skip expanding it.
			if len(relativePath) > 1 && relativePath[len(relativePath)-1:] != string(os.PathSeparator) {
				if err = writeFile(rc, z.UncompressedSize64, relativePath, z.Mode()); err != nil {
					break
				}
			}
		}
	}

	// Remove the archive
	if keepArchive == false {
		if err := os.Remove(archive); err != nil {
			log.Printf("Could not remove %s\n", archive)
		} else {
			fmt.Printf("Archive %s was removed\n", archive)
		}
	}

	fmt.Println("")

	return err
}

func fetchMasterZip(templateURL string) (string, error) {
	var err error

	if err = createPath(cacheDirectory, os.FileMode(0755)); err != nil {
		log.Fatalf("Could not create %s directory", cacheDirectory)
	}

	templateURLSplit := strings.Split(templateURL, "/")
	templateURLSplit = templateURLSplit[len(templateURLSplit)-4 : len(templateURLSplit)-2]
	templateRepository := strings.Join(templateURLSplit, "-")

	archive := fmt.Sprintf(cacheDirectory+"template-%s.zip", templateRepository)

	if _, err = os.Stat(archive); err != nil {
		c := http.Client{}

		req, err := http.NewRequest("GET", templateURL, nil)
		if err != nil {
			log.Println(err.Error())
			return "", err
		}
		log.Printf("HTTP GET %s\n", templateURL)
		res, err := c.Do(req)
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
	return archive, err
}

func writeFile(rc io.ReadCloser, size uint64, relativePath string, perms os.FileMode) error {
	var err error

	defer rc.Close()
	fmt.Printf("Writing %d bytes to \"%s\"\n", size, relativePath)
	if strings.HasSuffix(relativePath, "/") {
		mkdirErr := os.MkdirAll(relativePath, perms)
		if mkdirErr != nil {
			return fmt.Errorf("error making directory %s got: %s", relativePath, mkdirErr)
		}
		return err
	}

	// Create a file instead.
	f, err := os.OpenFile(relativePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perms)
	if err != nil {
		return fmt.Errorf("error writing to %s got: %s", relativePath, err)
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
func canWriteLanguage(language string, overwrite bool) bool {
	if len(language) > 0 {
		if cache, ok := cache[language]; ok {
			return cache
		}

		dir := templateDirectory + language
		if _, err := os.Stat(dir); err == nil && overwrite == false {
			fmt.Println(err, overwrite)
			log.Printf("Directory %s exists, overwrite is not allowed\n", dir)
			cache[language] = false
		} else {
			cache[language] = true
		}

		return cache[language]
	}

	log.Println("Language is unkwown, skip")
	return false
}
