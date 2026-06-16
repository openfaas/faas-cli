// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"os"

	"github.com/openfaas/faas-cli/commands"
)

func main() {
	commands.Execute(os.Args)
}
