package flags

import (
	"errors"
	"testing"
)

func TestLogFormat(t *testing.T) {
	cases := []struct {
		name  string
		value string
		err   error
	}{
		{"can accept plain", "plain", nil},
		{"can accept keyvalue", "keyvalue", nil},
		{"can accept json", "json", nil},
		{"unknown strings cause error string", "nonsense", errors.New("unknown log format: 'nonsense'")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var f LogFormat
			err := f.Set(tc.value)
			if tc.err != nil && tc.err.Error() != err.Error() {
				t.Fatalf("expected error %s, got %s", tc.err, err)
			}
			if tc.err == nil && f.String() != tc.value {
				t.Errorf("expected format %s, got %s", tc.value, f.String())
			}
		})
	}

}
