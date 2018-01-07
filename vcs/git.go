package vcs

// Git describes how to use Git.
var Git = &vcsCmd{
	name: "Git",
	cmd:  "git",

	createCmd: []string{"clone {repo} {dir}"},

	scheme:  []string{"git", "https", "http", "git+ssh", "ssh"},
	pingCmd: "ls-remote {scheme}://{repo}",
}
