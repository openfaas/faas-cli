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

const ZipFileName string = "master.zip"

// fetchTemplates fetch code templates from GitHub master zip file.
func fetchTemplates(templateUrl string) error {

	err := fetchMasterZip(templateUrl)

	zipFile, err := zip.OpenReader(ZipFileName)
	if err != nil {
		return err
	}

	log.Printf("Attempting to expand templates from %s\n", ZipFileName)

	for _, z := range zipFile.File {
		relativePath := strings.Replace(z.Name, "faas-cli-master/", "", -1)
		if strings.Index(relativePath, "template") == 0 {
			fmt.Printf("Found \"%s\"\n", relativePath)
			rc, err := z.Open()
			if err != nil {
				return err
			}

			err = createPath(relativePath, z.Mode())
			if err != nil {
				return err
			}

			// If relativePath is just a directory, then skip expanding it.
			if len(relativePath) > 1 && relativePath[len(relativePath)-1:] != string(os.PathSeparator) {
				err = writeFile(rc, z.UncompressedSize64, relativePath, z.Mode())
				if err != nil {
					return err
				}
			}
		}
	}

	log.Printf("Cleaning up zip file...")
	if _, err := os.Stat(ZipFileName); err == nil {
		os.Remove(ZipFileName)
	} else {
		return err
	}
	fmt.Println("")

	return err
}

func fetchMasterZip(templateUrl string) error {
	var err error
	if _, err = os.Stat(ZipFileName); err != nil {

		if len(templateUrl) == 0 {
			templateUrl = "https://github.com/alexellis/faas-cli/archive/" + ZipFileName
		}
		c := http.Client{}

		req, err := http.NewRequest("GET", templateUrl, nil)
		if err != nil {
			log.Println(err.Error())
			return err
		}
		log.Printf("HTTP GET %s\n", templateUrl)
		res, err := c.Do(req)
		if err != nil {
			log.Println(err.Error())
			return err
		}
		if res.Body != nil {
			defer res.Body.Close()
		}
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println(err.Error())
			return err
		}

		log.Printf("Writing %dKb to %s\n", len(bytesOut)/1024, ZipFileName)
		err = ioutil.WriteFile(ZipFileName, bytesOut, 0700)
		if err != nil {
			log.Println(err.Error())
		}
	}
	fmt.Println("")
	return err
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
