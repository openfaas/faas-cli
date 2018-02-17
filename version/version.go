package version

import "strings"

const UserAgent = "OpenFaaS CLI"

var (
	Version, GitCommit string
)

func BuildVersion() string {
	if IsDev() {
		return "dev"
	}
	return Version
}

func IsDev() bool {
	return len(Version) == 0
}

// CompareVersion compares two semver and returns -1 if the first one is greater, 1 if the second one is greater,
// and 0 if identical
func CompareVersion(oldVer string, newVer string) int {
	oldVer = strings.Replace(oldVer, "v", "", 1)
	newVer = strings.Replace(newVer, "v", "", 1)

	if oldVer == newVer {
		return 0
	}

	oldVerParts := strings.Split(oldVer, ".")
	newVerParts := strings.Split(newVer, ".")

	return compareVersionParts(oldVerParts, newVerParts, 0)
}

func compareVersionParts(oldVerParts []string, newVerParts []string, i int) int {
	if oldVerParts[i] < newVerParts[i] {
		return 1
	} else if oldVerParts[i] > newVerParts[i] {
		return -1
	} else {
		if i == len(oldVerParts)-1 && i == len(newVerParts)-1 {
			// last part
			return 0
		} else if i == len(oldVerParts)-1 && i < len(newVerParts)-1 {
			// newVer has more part
			return 1
		} else if i < len(oldVerParts)-1 && i == len(newVerParts)-1 {
			// oldVer has more part
			return -1
		} else {
			return compareVersionParts(oldVerParts, newVerParts, i+1)
		}
	}
}
