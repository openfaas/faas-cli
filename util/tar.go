// Copyright (c) OpenFaaS Project 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package util

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// CreateTar creates a tar buffer from a directory path
func CreateTar(path string, buffer *bytes.Buffer) {
	tw := tar.NewWriter(buffer)
	defer tw.Close()

	path = strings.TrimRight(filepath.ToSlash(path), "/") + "/"
	addDirToTar(tw, path, path)
}

// addDirToTar adds a directory to the tar writer
func addDirToTar(tw *tar.Writer, path string, basepath string) {
	if dir, err := ioutil.ReadDir(path); err != nil {
		fmt.Printf("ERROR: %s\n", err)
	} else {
		for _, f := range dir {
			if f.IsDir() {
				addDirToTar(tw, path+f.Name()+"/", basepath)
			} else {
				relativePath := path + f.Name()
				DebugPrint("Adding file to tar : %s\n", relativePath)
				if err := addFileToTar(tw, relativePath, basepath); err != nil {
					fmt.Printf("ERROR: %s\n", err)
				}
			}
		}
	}
}

// addFileToTar adds file to the tar writer
func addFileToTar(tw *tar.Writer, path string, basepath string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if stat, err := file.Stat(); err != nil {
		return err
	} else {
		if header, err := tar.FileInfoHeader(stat, stat.Name()); err != nil {
			return err
		} else {
			header.Name = strings.Replace(path, basepath, "", 1)
			header.Size = stat.Size()
			header.Mode = int64(stat.Mode())
			header.ModTime = stat.ModTime()

			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			if _, err := io.Copy(tw, file); err != nil {
				return err
			}
		}
	}

	return nil
}
