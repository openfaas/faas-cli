// Copyright (c) OpenFaaS Project 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package builder

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func Test_cp(t *testing.T) {
	aModes := []int{
		0600,
		0640,
		0644,
		0700,
		0755,
	}

	for _, m := range aModes {
		t.Run(fmt.Sprintf("Mode %s", os.FileMode(m).String()), func(t *testing.T) {
			data := []byte("open faas")

			srcDir, srcDirErr := ioutil.TempDir(os.TempDir(), "openfaas-test-source-")
			if srcDirErr != nil {
				t.Errorf("Error creating source folder: %s", srcDirErr)
			}
			defer os.RemoveAll(srcDir)

			destDir, destDirErr := ioutil.TempDir(os.TempDir(), "openfaas-test-destination-")
			if destDirErr != nil {
				t.Errorf("Error creating destination folder: %s", destDirErr)
			}
			defer os.RemoveAll(destDir)

			fileName := fmt.Sprintf("mode_%s", os.FileMode(m).String())
			srcFilePath := srcDir + "/" + fileName
			destFilePath := destDir + "/" + fileName

			if fileErr := ioutil.WriteFile(srcFilePath, data, os.FileMode(m)); fileErr != nil {
				t.Errorf("Cannot create file: %s", fileErr)
			}

			if err := cp(srcFilePath, destFilePath); err != nil {
				t.Error(err)
			} else {
				if destFile, err := os.Stat(destFilePath); err != nil {
					t.Error(err)
				} else if srcFile, err := os.Stat(destFilePath); err != nil {
					t.Error(err)
				} else {
					if destFile.Mode() != srcFile.Mode() {
						t.Errorf("destination file mode %s does not match source file mode %s", destFile.Mode(), srcFile.Mode())
					}
				}
			}
		})
	}
}
