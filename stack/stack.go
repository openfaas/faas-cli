package stack

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	yaml "gopkg.in/yaml.v2"
)

const providerName = "faas"

// ParseYAML parse a YAML file into a stack of "services".
func ParseYAML(yamlFile string) (*Services, error) {
	var services Services
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

	err = yaml.Unmarshal(fileData, &services)
	if err != nil {
		fmt.Printf("Error with YAML file\n")
		return nil, err
	}

	if services.Provider.Name != providerName {
		return nil, fmt.Errorf("'%s' is the only valid provider for this tool - found: %s", providerName, services.Provider.Name)
	}

	return &services, err
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
