// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package stack

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"errors"

	"github.com/ryanuber/go-glob"
	yaml "gopkg.in/yaml.v2"
)

const providerName = "faas"

// ParseYAMLData parse YAML file into a stack of "services".
func ParseYAMLFile(yamlFile, regex, filter string) (*Services, error) {
	var err error
	var fileData []byte
	urlParsed, err := url.Parse(yamlFile)
	if err == nil && len(urlParsed.Scheme) > 0 {
		fmt.Println("Parsed: " + urlParsed.String())
		fileData, err = fetchYAML(urlParsed)
		if err != nil {
			return nil, err
		}
	} else {
		fileData, err = ioutil.ReadFile(yamlFile)
		if err != nil {
			return nil, err
		}
	}
	return ParseYAMLData(fileData, regex, filter)
}

// ParseYAMLData parse YAML data into a stack of "services".
func ParseYAMLData(fileData []byte, regex string, filter string) (*Services, error) {
	var services Services
	regexExists := len(regex) > 0
	filterExists := len(filter) > 0

	err := yaml.Unmarshal(fileData, &services)
	if err != nil {
		fmt.Printf("Error with YAML file\n")
		return nil, err
	}

	if services.Provider.Name != providerName {
		return nil, fmt.Errorf("'%s' is the only valid provider for this tool - found: %s", providerName, services.Provider.Name)
	}

	if regexExists && filterExists {
		return nil, errors.New("Pass in a regex or a filter, not both.")
	}

	if regexExists || filterExists {
		for k, function := range services.Functions {
			var match bool
			var err error
			function.Name = k

			if regexExists {
				match, err = regexp.MatchString(regex, function.Name)
				if err != nil {
					fmt.Printf("Invalid regex passed to --regex option\n")
					return nil, err
				}
			} else {
				match = glob.Glob(filter, function.Name)
			}

			if !match {
				delete(services.Functions, function.Name)
			}
		}

	}

	return &services, nil
}

// fetchYAML pulls in file from remote location such as GitHub raw file-view
func fetchYAML(address *url.URL) ([]byte, error) {
	req, err := http.NewRequest("GET", address.String(), nil)
	if err != nil {
		return nil, err
	}
	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBytes, err := ioutil.ReadAll(res.Body)

	return resBytes, err
}
