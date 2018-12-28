package versioncontrol

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
