package schema

import (
	"strings"
)

type BuildFormat int

// DefaultFormat as defined in the YAML file or appending :latest
const DefaultFormat BuildFormat = 0

// SHAFormat uses "latest-<sha>" as the docker tag
const SHAFormat BuildFormat = 1

// BranchAndSHAFormat uses "latest-<branch>-<sha>" as the docker tag
const BranchAndSHAFormat BuildFormat = 2

// DescribeFormat uses the git-describe output as the docker tag
const DescribeFormat BuildFormat = 3

// BuildImageName builds a Docker image tag for build, push or deploy
func BuildImageName(format BuildFormat, image string, version string, branch string) string {
	imageVal := image
	if strings.Contains(image, ":") == false {
		imageVal += ":latest"
	}

	switch format {
	case SHAFormat:
		return imageVal + "-" + version
	case BranchAndSHAFormat:
		return imageVal + "-" + branch + "-" + version
	case DescribeFormat:
		// should we trim the existing image tag and do a proper replace with
		// the describe describe value
		return imageVal + "-" + version
	default:
		return imageVal
	}
}
