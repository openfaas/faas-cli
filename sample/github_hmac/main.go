// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"handler/function"
)

func main() {
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Unable to read standard input: %s", err.Error())
	}

	fmt.Println(function.Handle(input))
}
