// Copyright (c) OpenFaaS Project 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package builder

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func Test_CopyFiles(t *testing.T) {
	aTestCases := []struct {
		subdirLevel int
	}{
		{
			subdirLevel: 0,
		},
		{
			subdirLevel: 2,
		},
	}

	for _, aTest := range aTestCases {
		t.Run(fmt.Sprintf("Source directory with %d subdirectories", aTest.subdirLevel), func(t *testing.T) {
			// set up a source folder
			srcDir, srcDirErr := setupSourceFolder(os.TempDir(), 5, aTest.subdirLevel)

			if srcDirErr != nil {
				log.Fatal("Error creating source folder")
			}
			defer os.RemoveAll(srcDir)

			// create a destination folder to copy the files to
			destDir, destDirErr := ioutil.TempDir(os.TempDir(), "openfaas-test-destination-")
			if destDirErr != nil {
				t.Errorf("Error creating destination folder: %s", destDirErr)
			}
			defer os.RemoveAll(destDir)

			CopyFiles(srcDir, destDir, true)

			compareDirs(srcDir, destDir, t)
		})
	}
}

func compareDirs(srcDir string, destDir string, t *testing.T) {
	if srcFiles, err := ioutil.ReadDir(srcDir); err != nil {
		t.Error(err)
	} else {
		// Verify if all files from src are copied to dest
		for _, sF := range srcFiles {
			if dF, err := os.Stat(destDir + "/" + sF.Name()); err != nil {
				t.Error(err)
			} else {
				if sF.IsDir() {
					compareDirs(srcDir+"/"+sF.Name(), destDir+"/"+sF.Name(), t)
				} else if !sF.IsDir() && !dF.IsDir() && sF.Size() != dF.Size() {
					t.Errorf("Size of %s (source) does not match %s's (destination)", sF.Name(), dF.Name())
				}
			}
		}
	}
}

func setupSourceFolder(tmpDir string, numberOfFiles int, subdirLevel int) (string, error) {
	aModes := []int{
		0600,
		0640,
		0644,
		0700,
		0755,
	}
	data := []byte("open faas")

	// create a folder for source files
	srcDir, dirError := ioutil.TempDir(tmpDir, "openfaas-test-source-")
	if dirError != nil {
		return "", dirError
	}

	// create n files inside the created folder
	for i := 1; i <= numberOfFiles; i++ {
		srcFile := fmt.Sprintf("%s/test-file-%d", srcDir, i)
		// create files in different modes
		fileErr := ioutil.WriteFile(srcFile, data, os.FileMode(aModes[i%len(aModes)]))
		if fileErr != nil {
			return "", fileErr
		}
	}

	if subdirLevel > 0 {
		subdirLevel -= 1
		setupSourceFolder(srcDir, numberOfFiles, subdirLevel)
	}

	return srcDir, nil
}
