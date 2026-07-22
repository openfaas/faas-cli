package commands

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

func TestBuildPluginPullOptions(t *testing.T) {
	platform := &v1.Platform{Architecture: "amd64", OS: "linux"}
	opts := crane.GetOptions(buildPluginPullOptions(platform)...)

	if opts.Platform != platform {
		t.Fatalf("want platform %p, got %p", platform, opts.Platform)
	}

	authOptionName := runtime.FuncForPC(reflect.ValueOf(opts.Remote[0]).Pointer()).Name()
	if !strings.Contains(authOptionName, "WithAuth") || strings.Contains(authOptionName, "WithAuthFromKeychain") {
		t.Fatalf("want anonymous auth option, got %q", authOptionName)
	}
}

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
