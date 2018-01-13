package versioncontrol

// GitClone defines the command to clone a repo into a directory
var GitClone = &vcsCmd{
	name:   "Git",
	cmd:    "git",
	cmds:   []string{"clone {repo} {dir} --depth=1"},
	scheme: []string{"git", "https", "http", "git+ssh", "ssh"},
}
