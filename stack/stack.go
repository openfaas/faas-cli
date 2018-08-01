// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package stack

import (
	"fmt"
	"github.com/ryanuber/go-glob"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

const providerName = "faas"
const providerNameLong = "openfaas"

// ParseYAMLFile parse YAML file into a stack of "services".
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

	for _, f := range services.Functions {
		if f.Language == "Dockerfile" {
			f.Language = "dockerfile"
		}
	}

	if services.Provider.Name != providerName && services.Provider.Name != providerNameLong {
		return nil, fmt.Errorf("['%s', '%s'] is the only valid provider for this tool - found: %s", providerName, providerNameLong, services.Provider.Name)
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

// Custom unmarshaler to allow scheduling of extended resources such as GPUs, FPGAs from various vendors
func (config *FunctionResources) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var raw map[string]string
	var others = make(map[string]string)

	r, _ := regexp.Compile("^[[:graph:]]+[.]{1}[[:alnum:]]+/(gpu|fpga)$")

	if err := unmarshal(&raw); err != nil {
		return err
	}
	for key, value := range raw {
		switch {
		case key == "cpu":
			config.CPU = value
		case key == "memory":
			config.Memory = value
		case r.MatchString(key):
			others[key] = value
		default:
			fmt.Errorf("Ignoring unknown extended resource: %s with value: %s\n", key, value)
		}
	}
	config.Others = others
	return nil
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
