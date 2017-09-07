package stack

import (
	"fmt"
	"io/ioutil"
	"net/url"
)

func ParseYAML(file string, parseFunc func([]byte, ...string) (interface{}, error), parseFuncArgs ...string) (interface{}, error) {
	var err error
	var fileData []byte
	urlParsed, err := url.Parse(file)
	if err == nil && len(urlParsed.Scheme) > 0 {
		fmt.Println("Parsed: " + urlParsed.String())
		fileData, err = fetchYAML(urlParsed)
		if err != nil {
			return nil, err
		}
	} else {
		fileData, err = ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
	}

	return parseFunc(fileData, parseFuncArgs...)
}
