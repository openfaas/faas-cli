package versioncontrol

import (
	"strings"

	"github.com/openfaas/faas-cli/exec"
)

// GitClone defines the command to clone a repo into a directory
var GitClone = &vcsCmd{
	name:   "Git",
	cmd:    "git",
	cmds:   []string{"clone {repo} {dir} --depth=1 --config core.autocrlf=false -b {refname}"},
	scheme: []string{"git", "https", "http", "git+ssh", "ssh"},
}

// GitCheckout defines the command to clone a repo into a directory
var GitCheckout = &vcsCmd{
	name:   "Git",
	cmd:    "git",
	cmds:   []string{"-C {dir} checkout {refname}"},
	scheme: []string{"git", "https", "http", "git+ssh", "ssh"},
}

// GitCheckRefName defines the command that validates if a string is a valid reference name or sha
var GitCheckRefName = &vcsCmd{
	name:   "Git",
	cmd:    "git",
	cmds:   []string{"check-ref-format --allow-onelevel {refname}"},
	scheme: []string{"git", "https", "http", "git+ssh", "ssh"},
}

// GitInitRepo initializes the working directory add commit all files & directories
var GitInitRepo = &vcsCmd{
	name: "Git",
	cmd:  "git",
	cmds: []string{
		"init {dir}",
		"config core.autocrlf false",
		"config user.email \"contact@openfaas.com\"",
		"config user.name \"OpenFaaS\"",
		"add {dir}",
		"commit -m \"Test-commit\"",
	},
	scheme: []string{"git", "https", "http", "git+ssh", "ssh"},
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
	sha := exec.CommandWithOutput(getDescribeCommand, true)
	if strings.Contains(sha, "Not a git repository") {
		return ""
	}
	sha = strings.TrimSuffix(sha, "\n")

	return sha
}

// GetGitSHA returns the short Git commit SHA from local repo
func GetGitSHA() string {
	getShaCommand := []string{"git", "rev-parse", "--short", "HEAD"}
	sha := exec.CommandWithOutput(getShaCommand, true)
	if strings.Contains(sha, "Not a git repository") {
		return ""
	}
	sha = strings.TrimSuffix(sha, "\n")

	return sha
}

func GetGitBranch() string {
	getBranchCommand := []string{"git", "rev-parse", "--symbolic-full-name", "--abbrev-ref", "HEAD"}
	branch := exec.CommandWithOutput(getBranchCommand, true)
	if strings.Contains(branch, "Not a git repository") {
		return ""
	}
	branch = strings.TrimSuffix(branch, "\n")
	return branch
}
