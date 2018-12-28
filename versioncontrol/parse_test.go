package versioncontrol

import (
	"fmt"
	"testing"
)

func Test_IsGitRemote(t *testing.T) {

	validURLs := []struct {
		name string
		url  string
	}{
		{name: "git protocol without .git suffix", url: "git://host.xz/path/to/repo"},
		{name: "git protocol", url: "git://host.xz/path/to/repo.git/"},
		{name: "scp style with ip address", url: "git@192.168.101.127:user/project.git"},
		{name: "scp style with hostname", url: "git@github.com:user/project.git"},
		{name: "http protocol with ip address", url: "http://192.168.101.127/user/project.git"},
		{name: "http protocol", url: "http://github.com/user/project.git"},
		{name: "http protocol without .git suffix", url: "http://github.com/user/project"},
		{name: "https protocol with ip address", url: "https://192.168.101.127/user/project.git"},
		{name: "https protocol with hostname", url: "https://github.com/user/project.git"},
		{name: "https protocol with basic auth", url: "https://username:password@github.com/username/repository.git"},
		{name: "ssh protocol with hostname no port", url: "ssh://user@host.xz/path/to/repo.git/"},
		{name: "ssh protocol with hostname and port", url: "ssh://user@host.xz:port/path/to/repo.git/"},
	}

	for _, scenario := range validURLs {
		t.Run(fmt.Sprintf("%s is a valid remote git url", scenario.name), func(t *testing.T) {
			if !IsGitRemote(scenario.url) {
				t.Errorf("Url %s should pass the regex %s", scenario.url, gitRemoteRepoRegexpStr)
			}

		})
	}

	invalidURLs := []struct {
		name string
		url  string
	}{
		{name: "git protocol with hash", url: "git://github.com/openfaas/faas.git#ff78lf9h"},
		{name: "local repo file protocol", url: "file:///path/to/repo.git/"},
		{name: "ssh missing username and port", url: "host.xz:/path/to/repo.git"},
		{name: "ssh username and missing port", url: "user@host.xz:path/to/repo.git"},
		{name: "relative local path", url: "path/to/repo.git/"},
		{name: "magic relative local", url: "~/path/to/repo.git"},
	}
	for _, scenario := range invalidURLs {
		t.Run(fmt.Sprintf("%s is not a valid remote git url", scenario.name), func(t *testing.T) {
			if IsGitRemote(scenario.url) {
				t.Errorf("Url %s should fail the regex %s", scenario.url, gitRemoteRepoRegexpStr)
			}

		})
	}
}

func Test_IsPinnedGitRemote(t *testing.T) {

	validURLs := []struct {
		name string
		url  string
	}{
		{name: "git protocol without .git suffix", url: "git://host.xz/path/to/repo" + pinCharater + "feature-branch"},
		{name: "git protocol", url: "git://host.xz/path/to/repo.git/" + pinCharater + "tagname"},
		{name: "scp style with ip address", url: "git@192.168.101.127:user/project.git" + pinCharater + "v1.2.3"},
		{name: "scp style with hostname", url: "git@github.com:user/project.git" + pinCharater + "feature-branch"},
		{name: "http protocol with ip address", url: "http://192.168.101.127/user/project.git" + pinCharater + "tagname"},
		{name: "http protocol", url: "http://github.com/user/project.git" + pinCharater + "v1.2.3"},
		{name: "http protocol without .git suffix", url: "http://github.com/user/project" + pinCharater + "feature-branch"},
		{name: "https protocol with ip address", url: "https://192.168.101.127/user/project.git" + pinCharater + "tagname"},
		{name: "https protocol with hostname", url: "https://github.com/user/project.git" + pinCharater + "v1.2.3"},
		{name: "https protocol with basic auth", url: "https://username:password@github.com/username/repository.git" + pinCharater + "feature/branch"},
		{name: "ssh protocol with hostname no port", url: "ssh://user@host.xz/path/to/repo.git/" + pinCharater + "v1.2.3"},
		{name: "ssh protocol with hostname and port", url: "ssh://user@host.xz:port/path/to/repo.git/" + pinCharater + "tagname"},
	}

	for _, scenario := range validURLs {
		t.Run(fmt.Sprintf("%s is a valid pinned remote git url", scenario.name), func(t *testing.T) {
			if !IsPinnedGitRemote(scenario.url) {
				t.Errorf("Url %s should pass the regex %s", scenario.url, gitPinnedRemoteRegexpStr)
			}

		})
	}

	invalidURLs := []struct {
		name string
		url  string
	}{
		{name: "ssh protocol with hostname no port without pin", url: "ssh://user@host.xz/path/to/repo.git/"},
		{name: "ssh protocol with hostname and port without pin", url: "ssh://user@host.xz:port/path/to/repo.git/"},
		{name: "scp style with ip address without pin", url: "git@192.168.101.127:user/project.git"},
		{name: "scp style with hostname without pin", url: "git@github.com:user/project.git"},
		{name: "git protocol without .git suffix and no tag", url: "git://host.xz/path/to/repo"},
		{name: "git protocol with hash", url: "git://github.com/openfaas/faas.git#ff78lf9h@feature/branch"},
		{name: "local repo file protocol", url: "file:///path/to/repo.git/@feature/branch"},
		{name: "ssh missing username and port", url: "host.xz:/path/to/repo.git" + pinCharater + "feature-branch"},
		{name: "ssh username and missing port", url: "user@host.xz:path/to/repo.git" + pinCharater + "v1.2.3"},
		{name: "relative local path", url: "path/to/repo.git/@feature/branch"},
		{name: "magic relative local", url: "~/path/to/repo.git@feature/branch"},
	}
	for _, scenario := range invalidURLs {
		t.Run(fmt.Sprintf("%s is not a valid pinned remote git url", scenario.name), func(t *testing.T) {
			if IsPinnedGitRemote(scenario.url) {
				t.Errorf("Url %s should fail the regex %s", scenario.url, gitPinnedRemoteRegexpStr)
			}

		})
	}
}

func Test_ParsePinnedRemote(t *testing.T) {

	cases := []struct {
		name    string
		url     string
		refName string
	}{
		{name: "git protocol without .git suffix", url: "git://host.xz/path/to/repo", refName: "feature-branch"},
		{name: "git protocol", url: "git://host.xz/path/to/repo.git/", refName: "tagname"},
		{name: "scp style with ip address", url: "git@192.168.101.127:user/project.git", refName: "v1.2.3"},
		{name: "scp style with hostname", url: "git@github.com:user/project.git", refName: "feature/branch"},
		{name: "http protocol with ip address", url: "http://192.168.101.127/user/project.git", refName: "tagname"},
		{name: "http protocol", url: "http://github.com/user/project.git", refName: "v1.2.3"},
		{name: "http protocol without .git suffix", url: "http://github.com/user/project", refName: "feature/branch"},
		{name: "https protocol with ip address", url: "https://192.168.101.127/user/project.git", refName: "tagname"},
		{name: "https protocol with hostname", url: "https://github.com/user/project.git", refName: "v1.2.3"},
		{name: "https protocol with basic auth", url: "https://username:password@github.com/username/repository.git", refName: "feature/branch"},
		{name: "ssh protocol with hostname no port", url: "ssh://user@host.xz/path/to/repo.git/", refName: "v1.2.3"},
		{name: "ssh protocol with hostname and port", url: "ssh://user@host.xz:port/path/to/repo.git/", refName: "tagname"},
	}

	for _, scenario := range cases {
		t.Run(fmt.Sprintf("can parse refname from url with %s", scenario.name), func(t *testing.T) {
			remote, refName := ParsePinnedRemote(scenario.url + pinCharater + scenario.refName)
			if remote != scenario.url {
				t.Errorf("expected remote url: %s, got: %s", scenario.url, remote)
			}

			if refName != scenario.refName {
				t.Errorf("expected refName: %s, got: %s", scenario.refName, refName)
			}

		})

		t.Run(fmt.Sprintf("can parse default refname from url with %s", scenario.name), func(t *testing.T) {
			remote, refName := ParsePinnedRemote(scenario.url)
			if remote != scenario.url {
				t.Errorf("expected remote url: %s, got: %s", scenario.url, remote)
			}

			if refName != "master" {
				t.Errorf("expected refName: master, got: %s", refName)
			}

		})
	}
}
