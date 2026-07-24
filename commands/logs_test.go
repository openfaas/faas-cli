package commands

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/openfaas/faas-provider/logs"
	"github.com/spf13/cobra"
)

func Test_logsCmdFlagParsing(t *testing.T) {
	nowFunc = func() time.Time {
		ts, _ := time.Parse(time.RFC3339, "2019-01-01T01:00:00Z")
		return ts
	}

	fiveMinAgoStr := "2019-01-01T00:55:00Z"
	fiveMinAgo, _ := time.Parse(time.RFC3339, fiveMinAgoStr)

	scenarios := []struct {
		name     string
		args     []string
		expected logs.Request
	}{
		{"name only passed, follow on by default", []string{"funcFoo"}, logs.Request{Name: "funcFoo", Follow: true, Tail: -1}},
		{"can disable follow", []string{"funcFoo", "--tail=false"}, logs.Request{Name: "funcFoo", Follow: false, Tail: -1}},
		{"can limit number of messages returned", []string{"funcFoo", "--lines=5"}, logs.Request{Name: "funcFoo", Follow: true, Tail: 5}},
		{"can set timestamp to send logs since using duration", []string{"funcFoo", "--since=5m"}, logs.Request{Name: "funcFoo", Follow: true, Tail: -1, Since: &fiveMinAgo}},
		{"can set timestamp to send logs since using timestamp", []string{"funcFoo", "--since-time=" + fiveMinAgoStr}, logs.Request{Name: "funcFoo", Follow: true, Tail: -1, Since: &fiveMinAgo}},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			functionLogsCmd.ResetFlags()

			initLogCmdFlags(functionLogsCmd)
			functionLogsCmd.ParseFlags(s.args)

			logRequest := logRequestFromFlags(functionLogsCmd, functionLogsCmd.Flags().Args())
			if logRequest.String() != s.expected.String() {
				t.Errorf("expected log request %s, got %s", s.expected, logRequest)
			}
		})
	}
}

func strP(s string) *string {
	return &s
}

const stackWithGateway = `version: 1.0
provider:
  name: openfaas
  gateway: http://gw.example.com:8080
functions:
  fn:
    lang: dockerfile
    handler: ./fn
    image: fn:latest
`

// notAStackFile is a valid YAML file that is not an OpenFaaS stack file, of the
// kind that can sit in a working directory under the name stack.yaml.
const notAStackFile = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
`

func Test_logsGatewayURL(t *testing.T) {
	scenarios := []struct {
		name string
		// stackFile is written to a temporary directory and, unless the
		// arguments name it with -f, assigned to yamlFile in the way that
		// checkAndSetDefaultYaml does for the working directory.
		stackFile string
		explicit  bool
		env       string
		args      []string
		want      string
		wantErr   bool
	}{
		{
			name: "default gateway when there is no stack file",
			want: defaultGateway,
		},
		{
			name:      "gateway comes from an auto-detected stack file",
			stackFile: stackWithGateway,
			want:      "http://gw.example.com:8080",
		},
		{
			name:      "gateway comes from a stack file given with -f",
			stackFile: stackWithGateway,
			explicit:  true,
			want:      "http://gw.example.com:8080",
		},
		{
			name:      "an auto-detected file that is not a stack file is not fatal",
			stackFile: notAStackFile,
			want:      defaultGateway,
		},
		{
			name:      "a file given with -f that is not a stack file is fatal",
			stackFile: notAStackFile,
			explicit:  true,
			wantErr:   true,
		},
		{
			name:      "an auto-detected file that is not a stack file still allows OPENFAAS_URL",
			stackFile: notAStackFile,
			env:       "http://env.example.com:8080",
			want:      "http://env.example.com:8080",
		},
		{
			name: "OPENFAAS_URL is used when there is no stack file",
			env:  "http://env.example.com:8080",
			want: "http://env.example.com:8080",
		},
		{
			name:      "the stack file takes precedence over OPENFAAS_URL",
			stackFile: stackWithGateway,
			env:       "http://env.example.com:8080",
			want:      "http://gw.example.com:8080",
		},
		{
			name:      "an explicit --gateway takes precedence over the stack file",
			stackFile: stackWithGateway,
			args:      []string{"--gateway", "http://flag.example.com:8080"},
			want:      "http://flag.example.com:8080",
		},
		{
			name:      "an explicit --gateway equal to the default takes precedence",
			stackFile: stackWithGateway,
			env:       "http://env.example.com:8080",
			args:      []string{"--gateway", defaultGateway},
			want:      defaultGateway,
		},
		{
			name: "an explicit --gateway takes precedence over OPENFAAS_URL",
			env:  "http://env.example.com:8080",
			args: []string{"-g", "http://flag.example.com:8080"},
			want: "http://flag.example.com:8080",
		},
		{
			name:     "a stack file given with -f that does not exist is fatal",
			explicit: true,
			wantErr:  true,
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			t.Setenv(openFaaSURLEnvironment, s.env)

			path := ""
			if len(s.stackFile) > 0 {
				path = filepath.Join(t.TempDir(), defaultYAML)
				if err := os.WriteFile(path, []byte(s.stackFile), 0600); err != nil {
					t.Fatalf("unable to write stack file: %s", err)
				}
			} else if s.explicit {
				path = filepath.Join(t.TempDir(), defaultYAML)
			}

			args := s.args
			if s.explicit {
				args = append([]string{"-f", path}, args...)
			}

			cmd := newTestLogsCmd(t, args)
			if !s.explicit {
				// checkAndSetDefaultYaml sets this from the working directory
				// whether or not the user asked for a stack file.
				yamlFile = path
			}

			got, err := logsGatewayURL(cmd)
			if s.wantErr {
				if err == nil {
					t.Fatalf("want an error, but got none, with gateway %q", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("want no error, but got: %s", err)
			}

			if got != s.want {
				t.Fatalf("want gateway %q, but got %q", s.want, got)
			}
		})
	}
}

// newTestLogsCmd builds the logs command under a root command holding the
// persistent --yaml flag, so that each scenario starts with the flags unset.
func newTestLogsCmd(t *testing.T, args []string) *cobra.Command {
	t.Helper()

	root := &cobra.Command{Use: "faas-cli"}
	root.PersistentFlags().StringVarP(&yamlFile, "yaml", "f", "", "Path to YAML file describing function(s)")

	cmd := &cobra.Command{Use: "logs"}
	root.AddCommand(cmd)
	initLogCmdFlags(cmd)

	if err := cmd.ParseFlags(args); err != nil {
		t.Fatalf("unable to parse flags %v: %s", args, err)
	}

	return cmd
}
