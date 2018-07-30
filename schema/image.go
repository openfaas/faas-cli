package schema

import (
	"strings"
)

type BuildFormat int

// DefaultFormat as defined in the YAML file or appending :latest
const DefaultFormat BuildFormat = 0

const SHAFormat BuildFormat = 1

const BranchAndSHAFormat BuildFormat = 2

// BuildImageName builds a Docker image tag for build, push or deploy
func BuildImageName(format BuildFormat, image string, SHA string, branch string) string {
	imageVal := image
	if strings.Contains(image, ":") == false {
		imageVal += ":latest"
	}

	if format == SHAFormat {
		return imageVal + "-" + SHA
	}
	if format == BranchAndSHAFormat {
		return imageVal + "-" + branch + "-" + SHA
	}

	return imageVal
}
