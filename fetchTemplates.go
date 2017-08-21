package main

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

// fetchTemplates fetch code templates from GitHub master zip file.
func fetchTemplates() error {

	err := fetchMasterZip()

	zipFile, err := zip.OpenReader("./master.zip")
	if err != nil {
		return err
	}

	log.Printf("Attempting to expand templates from master.zip\n")

	for _, z := range zipFile.File {
		relativePath := strings.Replace(z.Name, "faas-cli-master/", "", -1)
		if strings.Index(relativePath, "template") == 0 {
			fmt.Printf("Found %s.\n", relativePath)
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

	return err
}

func fetchMasterZip() error {
	var err error
	if _, err = os.Stat("master.zip"); err != nil {
		templateURL := os.Getenv("templateUrl")
		if len(templateURL) == 0 {
			templateURL = "https://github.com/alexellis/faas-cli/archive/master.zip"
		}
		c := http.Client{}

		req, err := http.NewRequest("GET", templateURL, nil)
		if err != nil {
			log.Println(err.Error())
			return err
		}
		log.Printf("HTTP GET %s\n", templateURL)
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

		log.Printf("Writing %dKb to master.zip\n", len(bytesOut)/1024)
		err = ioutil.WriteFile("./master.zip", bytesOut, 0700)
		if err != nil {
			log.Println(err.Error())
		}
	}
	return err
}

func writeFile(rc io.ReadCloser, size uint64, relativePath string, perms os.FileMode) error {
	var err error

	defer rc.Close()
	fmt.Printf("Writing %d bytes to %s.\n", size, relativePath)
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
