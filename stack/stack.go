// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package stack

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	envsubst "github.com/drone/envsubst"
	glob "github.com/ryanuber/go-glob"
	yaml "gopkg.in/yaml.v2"
)

const legacyProviderName = "faas"
const providerName = "openfaas"
const defaultSchemaVersion = "1.0"

// ValidSchemaVersions available schema versions
var ValidSchemaVersions = []string{
	"1.0",
}

// ParseYAMLFile parse YAML file into a stack of "services".
func ParseYAMLFile(yamlFile, regex, filter string, envsubst bool) (*Services, error) {
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
	return ParseYAMLData(fileData, regex, filter, envsubst)
}

func substituteEnvironment(data []byte) ([]byte, error) {

	ret, err := envsubst.Parse(string(data))
	if err != nil {
		return nil, err
	}

	res, resErr := ret.Execute(func(input string) string {
		if val, ok := os.LookupEnv(input); ok {
			return val
		}
		return ""
	})

	return []byte(res), resErr
}

// ParseYAMLData parse YAML data into a stack of "services".
func ParseYAMLData(fileData []byte, regex string, filter string, envsubst bool) (*Services, error) {
	var services Services
	regexExists := len(regex) > 0
	filterExists := len(filter) > 0

	var source []byte
	if envsubst {
		substData, substErr := substituteEnvironment(fileData)

		if substErr != nil {
			return &services, substErr
		}
		source = substData
	} else {
		source = fileData
	}

	err := yaml.Unmarshal(source, &services)
	if err != nil {
		fmt.Printf("Error with YAML file\n")
		return nil, err
	}

	for _, f := range services.Functions {
		if f.Language == "Dockerfile" {
			f.Language = "dockerfile"
		}
	}

	if services.Provider.Name != providerName {
		return nil, fmt.Errorf(`['%s'] is the only valid "provider.name" for the OpenFaaS CLI, but you gave: %s`, providerName, services.Provider.Name)
	}

	if len(services.Version) > 0 && !IsValidSchemaVersion(services.Version) {
		return nil, fmt.Errorf("%s are the only valid versions for the stack file - found: %s", ValidSchemaVersions, services.Version)
	}

	if regexExists && filterExists {
		return nil, fmt.Errorf("pass in a regex or a filter, not both")
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
			return nil, fmt.Errorf("no functions matching --filter/--regex were found in the YAML file")
		}

	}

	return &services, nil
}

func makeHTTPClient(timeout *time.Duration) http.Client {
	if timeout != nil {
		return http.Client{
			Timeout: *timeout,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout: *timeout,
					// KeepAlive: 0,
				}).DialContext,
				// MaxIdleConns:          1,
				// DisableKeepAlives:     true,
				IdleConnTimeout:       120 * time.Millisecond,
				ExpectContinueTimeout: 1500 * time.Millisecond,
			},
		}
	}

	// This should be used for faas-cli invoke etc.
	return http.Client{}
}

// fetchYAML pulls in file from remote location such as GitHub raw file-view
func fetchYAML(address *url.URL) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, address.String(), nil)
	if err != nil {
		return nil, err
	}

	timeout := 120 * time.Second
	client := makeHTTPClient(&timeout)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBytes, err := ioutil.ReadAll(res.Body)

	return resBytes, err
}

// IsValidSchemaVersion validates schema version
func IsValidSchemaVersion(schemaVersion string) bool {
	for _, validVersion := range ValidSchemaVersions {
		if schemaVersion == validVersion {
			return true
		}
	}
	return false
}
