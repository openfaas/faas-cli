package versioncontrol

import (
	"regexp"
	"strings"
)

const (
	pinCharater              = `#`
	gitRemoteRegexpStr       = `(git|ssh|https?|git@[-\w.]+):(\/\/)?([^#]*?(?:\.git)?\/?)`
	gitPinnedRemoteRegexpStr = gitRemoteRegexpStr + pinCharater + `[-\/\d\w._]+$`
	gitRemoteRepoRegexpStr   = gitRemoteRegexpStr + `$`
)

var (
	gitPinnedRegexp = regexp.MustCompile(gitPinnedRemoteRegexpStr)
	gitRemoteRegexp = regexp.MustCompile(gitRemoteRepoRegexpStr)
)

// IsGitRemote validates if the supplied string is a valid git remote url value
func IsGitRemote(repoURL string) bool {
	// If using a Regexp in multiple goroutines,
	// giving each goroutine its own copy helps to avoid lock contention.
	// https://golang.org/pkg/regexp/#Regexp.Copy
	return gitRemoteRegexp.Copy().MatchString(repoURL)
}

// IsPinnedGitRemote validates if the supplied string is a valid git remote url value
func IsPinnedGitRemote(repoURL string) bool {
	// If using a Regexp in multiple goroutines,
	// giving each goroutine its own copy helps to avoid lock contention.
	// https://golang.org/pkg/regexp/#Regexp.Copy
	return gitPinnedRegexp.Copy().MatchString(repoURL)
}

// ParsePinnedRemote returns the remote url and contraint value from repository url
func ParsePinnedRemote(repoURL string) (remoteURL, refName string) {
	refName = "master"
	remoteURL = repoURL

	// If using a Regexp in multiple goroutines,
	// giving each goroutine its own copy helps to avoid lock contention.
	// https://golang.org/pkg/regexp/#Regexp.Copy
	if !IsPinnedGitRemote(repoURL) {
		return remoteURL, refName
	}

	// handle ssh special case

	atIndex := strings.LastIndex(repoURL, pinCharater)
	if atIndex > 0 {
		refName = repoURL[atIndex+len(pinCharater):]
		remoteURL = repoURL[:atIndex]
	}

	if !IsGitRemote(remoteURL) {
		return repoURL, "master"
	}

	return remoteURL, refName
}
