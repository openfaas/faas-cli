// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package exec

import (
	"fmt"
	"log"
	"os"
	osexec "os/exec"

	"github.com/morikuni/aec"
)

// Command run a system command
func Command(tempPath string, builder []string) {
	targetCmd := osexec.Command(builder[0], builder[1:]...)
	targetCmd.Dir = tempPath
	targetCmd.Stdout = os.Stdout
	targetCmd.Stderr = os.Stderr

	targetCmd.Start()
	err := targetCmd.Wait()
	if err != nil {
		errString := fmt.Sprintf("ERROR - Could not execute command: %s", builder)
		log.Fatalf(aec.RedF.Apply(errString))
	}
}

// CommandWithOutput run a system command an return stdout
func CommandWithOutput(builder []string, skipFailure bool) string {
	output, err := osexec.Command(builder[0], builder[1:]...).CombinedOutput()
	if err != nil && !skipFailure {
		errString := fmt.Sprintf("ERROR - Could not execute command: %s", builder)
		log.Fatalf(aec.RedF.Apply(errString))
	}
	return string(output)
}
