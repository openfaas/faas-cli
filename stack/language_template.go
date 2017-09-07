// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package stack

import (
	"fmt"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

func ParseYAMLForLanguageTemplate(file string) (*LanguageTemplate, error) {
	if object, err := ParseYAML(
		file,
		iParseYAMLDataForLanguageTemplate,
	); err != nil {
		return nil, err
	} else {
		return object.(*LanguageTemplate), nil
	}
}

// ParseYAMLDataForLanguageTemplate parses YAML data into language template
func ParseYAMLDataForLanguageTemplate(fileData []byte, args ...string) (*LanguageTemplate, error) {
	if object, err := iParseYAMLDataForLanguageTemplate(fileData, args...); err != nil {
		return nil, err
	} else {
		return object.(*LanguageTemplate), nil
	}
}

// iParseYAMLDataForLanguageTemplate parses YAML data into language template
// Use the alias ParseYAMLDataForLanguageTemplate
func iParseYAMLDataForLanguageTemplate(fileData []byte, args ...string) (interface{}, error) {
	var langTemplate LanguageTemplate
	var err error

	err = yaml.Unmarshal(fileData, &langTemplate)
	if err != nil {
		fmt.Printf("Error with YAML file\n")
		return nil, err
	}

	return &langTemplate, err
}

func IsValidTemplate(lang string) bool {
	var found bool
	if strings.ToLower(lang) == "dockerfile" {
		found = true
	} else if _, err := os.Stat("./template/" + lang); err == nil {
		found = true
	}

	return found
}
