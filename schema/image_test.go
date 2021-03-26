package schema

import "testing"

func Test_BuildImageName_DefaultFormat(t *testing.T) {
	want := "img:latest"
	got := BuildImageName(DefaultFormat, "img", "ef384", "master")

	if got != want {
		t.Errorf("BuildImageName want: \"%s\", got: \"%s\"", want, got)
	}
}

func Test_BuildImageName_DefaultFormat_WithCustomServerPort(t *testing.T) {
	want := "registry:8080/honk/img:latest"
	got := BuildImageName(DefaultFormat, "registry:8080/honk/img", "ef384", "master")

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

func Test_BuildImageName_SHAFormat_WithCustomServerPort(t *testing.T) {
	want := "registry:5000/honk/img:latest-ef384"
	got := BuildImageName(SHAFormat, "registry:5000/honk/img", "ef384", "master")

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

func Test_BuildImageName_SHAFormat_WithNumericVersion_WithCustomServerPort(t *testing.T) {
	want := "registry:3000/honk/img:0.2-ef384"
	got := BuildImageName(SHAFormat, "registry:3000/honk/img:0.2", "ef384", "master")

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

func Test_BuildImageName_BranchAndSHAFormat_WithCustomServerPort(t *testing.T) {
	want := "registry:80/honk/img:latest-master-ef384"
	got := BuildImageName(BranchAndSHAFormat, "registry:80/honk/img", "ef384", "master")

	if got != want {
		t.Errorf("BuildImageName want: \"%s\", got: \"%s\"", want, got)
	}
}
