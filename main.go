// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"fmt"
	"os"

	"github.com/openfaas/faas-cli/commands"
)

func main() {
	customArgs, err := translateLegacyOpts(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	commands.Execute(customArgs)
}
