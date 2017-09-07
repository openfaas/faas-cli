// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package stack

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/ryanuber/go-glob"
	yaml "gopkg.in/yaml.v2"
)

const providerName = "faas"

// ParseYAMLFileForStack parses a YAML file and returns a stack of "services".
func ParseYAMLFileForStack(file string, regex string, filter string) (*Services, error) {
	if object, err := ParseYAML(
		file,
		iParseYAMLDataForStack,
		regex,
		filter,
	); err != nil {
		return nil, err
	} else {
		return object.(*Services), nil
	}
}

// ParseYAMLDataForStack parses YAML data into a stack of "services".
func ParseYAMLDataForStack(fileData []byte, args ...string) (*Services, error) {
	if object, err := iParseYAMLDataForStack(fileData, args...); err != nil {
		return nil, err
	} else {
		return object.(*Services), nil
	}
}

// iParseYAMLDataForStack parse YAML data into a stack of "services".
// Use the alias ParseYAMLDataForStack
func iParseYAMLDataForStack(fileData []byte, args ...string) (interface{}, error) {
	if len(args) != 2 {
		panic("ParseYAMLData func need exactly 3 arguments, (fileData, regex, filter)")
	}

	regex := args[0]
	filter := args[1]

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
					return nil, err
				}
			} else {
				match = glob.Glob(filter, function.Name)
			}

			if !match {
				delete(services.Functions, function.Name)
			}
		}

		if len(services.Functions) == 0 {
			return nil, errors.New("No functions matching --filter/--regex were found in the YAML file")
		}

	}

	return &services, nil
}

// fetchYAML pulls in file from remote location such as GitHub raw file-view
func fetchYAML(address *url.URL) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, address.String(), nil)
	if err != nil {
		return nil, err
	}

	timeout := 120 * time.Second
	client := proxy.MakeHTTPClient(&timeout)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBytes, err := ioutil.ReadAll(res.Body)

	return resBytes, err
}
