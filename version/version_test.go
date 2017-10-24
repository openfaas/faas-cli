package version

import (
	"testing"
)

func Test_EmptyVersionMeansBuildVersionReturnsDev(t *testing.T) {
	Version = ""
	output := BuildVersion()
	expected := "dev"
	if output != expected {
		t.Fatalf("Version is not from Build - want: %s, got: %s\n", expected, output)
	}
}

func Test_VersionReturnedFromBuildVersion(t *testing.T) {
	Version = "testing-manual"
	output := BuildVersion()
	expected := Version
	if output != expected {
		t.Fatalf("Version is not from Build - want: %s, got: %s\n", expected, output)
	}
}
