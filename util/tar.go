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
	"strings"
)

func CreateTar(path string, buffer *bytes.Buffer) {
	tw := tar.NewWriter(buffer)
	addDirToTar(tw, path, path)
	tw.Close()
}

func addDirToTar(tw *tar.Writer, path string, basepath string) {
	dir, _ := ioutil.ReadDir(path)
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
