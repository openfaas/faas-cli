package commands

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/openfaas/faas-provider/logs"
)

func Test_JSONLogFormatter(t *testing.T) {
	now := time.Now()
	msg := logs.Message{
		Timestamp: now,
		Name:      "test-func",
		Instance:  "123test",
		Text:      "test message\n",
	}
	msgJSON, _ := json.Marshal(msg)

	cases := []struct {
		name            string
		timeFormat      string
		includeName     bool
		includeInstance bool
		expected        string
	}{
		{"default behavior", "rfc3339", true, true, string(msgJSON)},
		{"default behavior with all empty options", "", false, false, string(msgJSON)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			formatted := JSONFormatMessage(msg, tc.timeFormat, tc.includeName, tc.includeInstance)
			if formatted != tc.expected {
				t.Fatalf("incorrect message format:\n got %s\n expected %s\n", formatted, tc.expected)
			}
		})
	}
}

func Test_PlainLogFormatter(t *testing.T) {
	ts := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	msg := logs.Message{
		Timestamp: ts,
		Name:      "test-func",
		Instance:  "123test",
		Text:      "test message\n",
	}

	cases := []struct {
		name            string
		timeFormat      string
		includeName     bool
		includeInstance bool
		expected        string
	}{
		{"default settings", time.RFC3339, true, true, "2009-11-10T23:00:00Z test-func (123test) test message"},
		{"default can modify timestamp", "2006-01-02 15:04:05.999999999 -0700 MST", true, true, msg.String()},
		{"can hide name", time.RFC3339, false, true, "2009-11-10T23:00:00Z (123test) test message"},
		{"can hide instance", time.RFC3339, true, false, "2009-11-10T23:00:00Z test-func test message"},
		{"can hide all metadata", "", false, false, "test message"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			formatted := PlainFormatMessage(msg, tc.timeFormat, tc.includeName, tc.includeInstance)
			if strings.TrimSpace(formatted) != strings.TrimSpace(tc.expected) {
				t.Fatalf("incorrect message format:\n got %s\n expected %s\n", formatted, tc.expected)
			}
		})
	}
}

func Test_KeyValueLogFormatter(t *testing.T) {
	ts := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	msg := logs.Message{
		Timestamp: ts,
		Name:      "test-func",
		Instance:  "123test",
		Text:      "test message\n",
	}

	cases := []struct {
		name            string
		timeFormat      string
		includeName     bool
		includeInstance bool
		expected        string
	}{
		{"default settings", time.RFC3339, true, true, "timestamp=\"2009-11-10T23:00:00Z\" name=\"test-func\" instance=\"123test\" text=\"test message\""},
		{"default settings", "2006-01-02 15:04:05.999999999 -0700 MST", true, true, "timestamp=\"2009-11-10 23:00:00 +0000 UTC\" name=\"test-func\" instance=\"123test\" text=\"test message\""},
		{"can hide name", time.RFC3339, false, true, "timestamp=\"2009-11-10T23:00:00Z\" instance=\"123test\" text=\"test message\""},
		{"can hide instance", time.RFC3339, true, false, "timestamp=\"2009-11-10T23:00:00Z\" name=\"test-func\" text=\"test message\""},
		{"can hide all metadata", "", false, false, "text=\"test message\""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			formatted := KeyValueFormatMessage(msg, tc.timeFormat, tc.includeName, tc.includeInstance)
			if strings.TrimSpace(formatted) != strings.TrimSpace(tc.expected) {
				t.Fatalf("incorrect message format:\n got %s\n expected %s\n", formatted, tc.expected)
			}
		})
	}
}
