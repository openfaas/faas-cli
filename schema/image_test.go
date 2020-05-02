package schema

import "testing"

func Test_BuildImageName_DefaultFormat(t *testing.T) {
	want := "img:latest"
	got := BuildImageName(DefaultFormat, "img", "ef384", "master")

	if got != want {
		t.Errorf("BuildImageName want: \"%s\", got: \"%s\"", want, got)
	}
}

func Test_BuildImageName_SHAFormat(t *testing.T) {
	want := "img:latest-ef384"
	got := BuildImageName(SHAFormat, "img", "ef384", "master")

	if got != want {
		t.Errorf("BuildImageName want: \"%s\", got: \"%s\"", want, got)
	}
}

func Test_BuildImageName_SHAFormat_WithNumericVersion(t *testing.T) {
	want := "img:0.2-ef384"
	got := BuildImageName(SHAFormat, "img:0.2", "ef384", "master")

	if got != want {
		t.Errorf("BuildImageName want: \"%s\", got: \"%s\"", want, got)
	}
}

func Test_BuildImageName_BranchAndSHAFormat(t *testing.T) {
	want := "img:latest-master-ef384"
	got := BuildImageName(BranchAndSHAFormat, "img", "ef384", "master")

	if got != want {
		t.Errorf("BuildImageName want: \"%s\", got: \"%s\"", want, got)
	}
}

func Test_BuildImageName_RegistryWithPort(t *testing.T) {
	want := "registry.domain:8080/image:latest"
	got := BuildImageName(DefaultFormat, "registry.domain:8080/image", "ef384", "master")

	if got != want {
		t.Errorf("BuildImageName want: \"%s\", got: \"%s\"", want, got)
	}
}

func Test_BuildImageName_RegistryWithPortAndTag(t *testing.T) {
	want := "registry.domain:8080/image:foo"
	got := BuildImageName(DefaultFormat, "registry.domain:8080/image:foo", "ef384", "master")

	if got != want {
		t.Errorf("BuildImageName want: \"%s\", got: \"%s\"", want, got)
	}
}
