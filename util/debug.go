// Copyright (c) OpenFaaS Project 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package util

import (
	"fmt"
	"os"

	"github.com/morikuni/aec"
)

var useDockerCLIWarningDisplayed bool

func IsDebugEnabled() bool {
	val, exists := os.LookupEnv("debug")
	return exists && (val == "1" || val == "true")
}

func UseDockerCLI() bool {
	val, exists := os.LookupEnv("openfaas_docker_cli")
	// TODO remove the condition !exists to change the default mode to use DOCKER API
	// !exists means activated by default
	if !useDockerCLIWarningDisplayed && !exists {
		useDockerCLIWarningDisplayed = true
		fmt.Println(aec.YellowF.Apply(`WARNING: In future release, Docker API will be used by default instead of Docker CLI
If you want to test the future behaviour, set environment variable openfaas_docker_cli=0`))
	}
	return !exists || (exists && (val == "1" || val == "true"))
}

func DebugPrint(format string, a ...interface{}) {
	if IsDebugEnabled() {
		fmt.Printf(aec.LightYellowF.Apply(format), a...)
	}
}
