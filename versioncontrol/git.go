package versioncontrol

// GitClone defines the command to clone a repo into a directory
var GitClone = &vcsCmd{
	name:   "Git",
	cmd:    "git",
	cmds:   []string{"clone {repo} {dir} --depth=1"},
	scheme: []string{"git", "https", "http", "git+ssh", "ssh"},
}

// GitInitRepo initializes the working directory add commit all files & directories
var GitInitRepo = &vcsCmd{
	name: "Git",
	cmd:  "git",
	cmds: []string{
		"init {dir}",
		"config user.email \"contact@openfaas.com\"",
		"config user.name \"OpenFaaS\"",
		"add {dir}",
		"commit -m \"Test-commit\"",
	},
	scheme: []string{"git", "https", "http", "git+ssh", "ssh"},
}
