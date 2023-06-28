package schema

import (
	"fmt"
	"strings"
)

// BuildFormat defines the docker image tag format that is used during the build process
type BuildFormat int

// DefaultFormat as defined in the YAML file or appending :latest
const DefaultFormat BuildFormat = 0

// SHAFormat uses "latest-<sha>" as the docker tag
const SHAFormat BuildFormat = 1

// BranchAndSHAFormat uses "latest-<branch>-<sha>" as the docker tag
const BranchAndSHAFormat BuildFormat = 2

// DescribeFormat uses the git-describe output as the docker tag
const DescribeFormat BuildFormat = 3

// DigestFormat
const DigestFormat BuildFormat = 4

// Type implements pflag.Value
func (i *BuildFormat) Type() string {
	return "string"
}

// String implements Stringer
func (i *BuildFormat) String() string {
	if i == nil {
		return "latest"
	}

	switch *i {
	case DefaultFormat:
		return "latest"
	case SHAFormat:
		return "sha"
	case BranchAndSHAFormat:
		return "branch"
	case DescribeFormat:
		return "describe"
	default:
		return "latest"
	}
}

// Set implements pflag.Value
func (i *BuildFormat) Set(value string) error {
	switch strings.ToLower(value) {
	case "", "default", "latest":
		*i = DefaultFormat
	case "sha":
		*i = SHAFormat
	case "branch":
		*i = BranchAndSHAFormat
	case "describe":
		*i = DescribeFormat
	case "digest":
		*i = DigestFormat
	default:
		return fmt.Errorf("unknown image tag format: '%s'", value)
	}
	return nil
}

// BuildImageName builds a Docker image tag for build, push or deploy
func BuildImageName(format BuildFormat, image string, version string, branch string) string {
	imageVal := image
	splitImage := strings.Split(image, "/")
	if strings.Contains(splitImage[len(splitImage)-1], ":") == false {
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
	case DigestFormat:
		baseImage, _, found := strings.Cut(imageVal, ":")
		if !found {
			return imageVal + "-" + version
		}

		return fmt.Sprintf("%s:%s", baseImage, version)

	default:
		return imageVal
	}
}
