package commands

import "testing"

func Test_getDownloadArch(t *testing.T) {
	tables := []struct {
		arch     string
		wantArch string
		os       string
		wantOS   string
	}{
		{
			arch:     "x86_64",
			wantArch: "amd64",
			os:       "Linux",
			wantOS:   "linux",
		},
		{
			arch:     "aarch64",
			wantArch: "arm64",
			os:       "Linux",
			wantOS:   "linux",
		},
		{
			arch:     "aarch64",
			wantArch: "arm64",
			os:       "Darwin",
			wantOS:   "darwin",
		},
		{
			arch:     "x86_64",
			wantArch: "amd64",
			os:       "Darwin",
			wantOS:   "darwin",
		},
		{
			arch:     "amd64",
			wantArch: "amd64",
			os:       "Windows",
			wantOS:   "windows",
		},
	}

	for _, table := range tables {
		gotArch, gotOS := getDownloadArch(table.arch, table.os)

		if gotArch != table.wantArch {
			t.Errorf("Incorrect arch, got: %s, want: %s.", gotArch, table.wantArch)
		}

		if gotOS != table.wantOS {
			t.Errorf("Incorrect os, got: %s, want: %s.", gotArch, table.wantArch)
		}
	}
}
