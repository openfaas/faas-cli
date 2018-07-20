// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package builder

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/morikuni/aec"
)

// ExecCommand run a system command
func ExecCommand(tempPath string, builder []string) {
	targetCmd := exec.Command(builder[0], builder[1:]...)
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

// ExecCommand run a system command an return stdout
func ExecCommandWithOutput(builder []string) string {
	output, err := exec.Command(builder[0], builder[1:]...).Output()
	if err != nil {
		errString := fmt.Sprintf("ERROR - Could not execute command: %s", builder)
		log.Fatalf(aec.RedF.Apply(errString))
	}
	return string(output)
}

//Generate image version of type gittag-gitsha
func GetVersion() string {
	verifyGitDirCommand := []string{"/bin/sh", "-c", "if [ -d .git ]; then echo True; fi;"}
	gitDir := ExecCommandWithOutput(verifyGitDirCommand)
	gitDir = strings.TrimSuffix(gitDir, "\n")
	if gitDir != "True" {
		return ""
	}

	getShaCommand := []string{"git", "rev-parse", "--short", "HEAD"}
	sha := ExecCommandWithOutput(getShaCommand)
	sha = strings.TrimSuffix(sha, "\n")

	getTagCommand := []string{"git", "tag", "--points-at", sha}
	tag := ExecCommandWithOutput(getTagCommand)
	tag = strings.TrimSuffix(tag, "\n")
	if len(tag) == 0 {
		tag = "latest"
	}

	return ":" + tag + "-" + sha
}
