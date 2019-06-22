package commands

import (
	"testing"
	"time"

	"github.com/openfaas/faas-provider/logs"
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
		{"can disable follow", []string{"funcFoo", "--follow=false"}, logs.Request{Name: "funcFoo", Follow: false, Tail: -1}},
		{"can limit number of messages returned", []string{"funcFoo", "--tail=5"}, logs.Request{Name: "funcFoo", Follow: true, Tail: 5}},
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
