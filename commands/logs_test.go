package commands

import (
	"errors"
	"testing"
	"time"

	"github.com/openfaas/faas-cli/schema"
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
		expected schema.LogRequest
	}{
		{"name only passed, follow on by default", []string{"funcFoo"}, schema.LogRequest{Name: "funcFoo", Follow: true}},
		{"can disable follow", []string{"funcFoo", "--follow=false"}, schema.LogRequest{Name: "funcFoo", Follow: false}},
		{"can supply filter pattern", []string{"funcFoo", "--pattern=abc"}, schema.LogRequest{Name: "funcFoo", Follow: true, Pattern: strP("abc")}},
		{"can invert filter pattern", []string{"funcFoo", "--pattern=abc", "--invert=true"}, schema.LogRequest{Name: "funcFoo", Follow: true, Pattern: strP("abc"), Invert: true}},
		{"can limit number of messages returned", []string{"funcFoo", "--limit=5"}, schema.LogRequest{Name: "funcFoo", Follow: true, Limit: 5}},
		{"can set timestamp to send logs since using duration", []string{"funcFoo", "--since=5m"}, schema.LogRequest{Name: "funcFoo", Follow: true, Since: &fiveMinAgo}},
		{"can set timestamp to send logs since using timestamp", []string{"funcFoo", "--since=" + fiveMinAgoStr}, schema.LogRequest{Name: "funcFoo", Follow: true, Since: &fiveMinAgo}},
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

func Test_mustTimestampP(t *testing.T) {
	nowFunc = func() time.Time {
		ts, _ := time.Parse(time.RFC3339, "2019-01-01T01:00:00Z")
		return ts
	}

	scenarios := []struct {
		name     string
		value    string
		err      error
		panic    bool
		expected string
	}{
		{"empty string returns nil", "", nil, false, "nil"},
		{"duration is parsed", "1h", nil, false, "2019-01-01T00:00:00Z"},
		{"timestamp is parsed", "2019-01-01T05:01:10Z", nil, false, "2019-01-01T05:01:10Z"},
		{"will panic if timestamp cannot be parsed", "-2019-01-01T05:01:10Z", nil, true, "2019-01-01T05:01:10Z"},
		{"will panic if supplied an error", "", errors.New("some flag error"), true, ""},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !s.panic {
					t.Errorf("mustTimestampP should only panic when given an error or when time.Parse errors")
				}
			}()

			ts := mustTimestampP(s.value, s.err)
			switch s.expected {
			case "nil":
				if ts != nil {
					t.Errorf("expected nil time, got %v", ts)
				}
			default:
				if ts.Format(time.RFC3339) != s.expected {
					t.Errorf("expected %s time, got %s", s.expected, ts)
				}
			}
		})
	}
}
