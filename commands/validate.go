// Copyright (c) OpenFaaS Author(s) 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"errors"
)

func validateLanguageFlag(language string) (string, error) {
	var err error

	if language == "Dockerfile" {
		language = "dockerfile"

		err = errors.New(`language "Dockerfile" was converted to "dockerfile" automatically`)
	}

	return language, err
}
