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
func ExecCommandWithOutput(builder []string, skipFailure bool) string {
	output, err := exec.Command(builder[0], builder[1:]...).CombinedOutput()
	if err != nil && !skipFailure {
		errString := fmt.Sprintf("ERROR - Could not execute command: %s", builder)
		log.Fatalf(aec.RedF.Apply(errString))
	}
	return string(output)
}

// GetGitDescribe returns the human readable name for the current commit using `git-describe`
func GetGitDescribe() string {
	// git-describe - Give an object a human readable name based on an available ref
	// --tags                use any tag, even unannotated
	// --always              show abbreviated commit object as fallback

	// using --tags, means that the output should look like v1.2.2-1-g3443110 where the last
	// <most-recent-parent-tag>-<number-of-commits-to-that-tag>-g<short-sha>
	// using --always, means that if the repo does not use tags, then we will still get the <short-sha>
	// as output, similar to GetGitSHA
	getDescribeCommand := []string{"git", "describe", "--tags", "--always"}
	sha := ExecCommandWithOutput(getDescribeCommand, true)
	if strings.Contains(sha, "Not a git repository") {
		return ""
	}
	sha = strings.TrimSuffix(sha, "\n")

	return sha
}

// GetGitSHA returns the short Git commit SHA from local repo
func GetGitSHA() string {
	getShaCommand := []string{"git", "rev-parse", "--short", "HEAD"}
	sha := ExecCommandWithOutput(getShaCommand, true)
	if strings.Contains(sha, "Not a git repository") {
		return ""
	}
	sha = strings.TrimSuffix(sha, "\n")

	return sha
}

func GetGitBranch() string {
	getBranchCommand := []string{"git", "rev-parse", "--symbolic-full-name", "--abbrev-ref", "HEAD"}
	branch := ExecCommandWithOutput(getBranchCommand, true)
	if strings.Contains(branch, "Not a git repository") {
		return ""
	}
	branch = strings.TrimSuffix(branch, "\n")
	return branch
}
