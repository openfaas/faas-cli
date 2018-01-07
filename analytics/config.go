// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package analytics

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/satori/go.uuid"
)

const analyticsUUIDFile = "analytics_uuid"

func configDir() (string, error) {
	h, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("unable to detect homedir: %v", err)
	}
	fullpath, err := homedir.Expand(h)
	if err != nil {
		return "", fmt.Errorf("unable to expand homedir [%s]: %v", h, err)
	}

	return path.Clean(fullpath + "/.openfaas/"), nil
}

func configFile() string {
	dir, _ := configDir()
	return path.Clean(dir + "/" + analyticsUUIDFile)
}

func configDirExists() bool {
	dir, _ := configDir()
	if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
		return true
	}
	return false
}

func setUUID() (string, error) {
	if !configDirExists() {
		dir, _ := configDir()
		err := os.Mkdir(dir, 0700)
		if err != nil {
			return "", fmt.Errorf("Unable to create config dir: %v\n", err)
		}
	}
	uuidFile := configFile()
	uuidStr := uuid.NewV4().String()
	d1 := []byte(uuidStr)
	err := ioutil.WriteFile(uuidFile, d1, 0644)
	if err != nil {
		return "", fmt.Errorf("unable to write analytics ID file: %v", err)
	}
	fmt.Fprintln(os.Stderr, "# Creating analytics file in:", uuidFile)
	fmt.Fprintln(os.Stderr, "# Please see https://github.com/openfaas/faas-cli/blob/master/analytics.md for more information.")
	return uuidStr, nil
}

func getUUID() (string, error) {
	dat, err := ioutil.ReadFile(configFile())
	if err != nil {
		return "", fmt.Errorf("Error reading file: %v", err)
	}
	uuidStr := string(dat)

	_, err = uuid.FromString(uuidStr)
	if err != nil {
		return "", fmt.Errorf("Unable to get valid UUID from file: %v", err)
	}

	return uuidStr, nil
}
