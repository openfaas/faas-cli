package flags

import (
	"testing"
	"time"
)

func TestTimestamp(t *testing.T) {

	cases := []struct {
		name     string
		value    string
		expected time.Time
		err      error
	}{
		{"valid rfc3339 parses", "2012-01-02T10:01:12Z", time.Date(2012, time.January, 2, 10, 1, 12, 0, time.UTC), nil},
		{"valid rfc3339 parses", "2012-01-02T10:01:12Z", time.Date(2012, time.January, 2, 10, 1, 12, 0, time.UTC), nil},
		{"in-valid rfc3339 parses", "2012-01-02T10:01:12Z", time.Date(2012, time.January, 2, 10, 1, 12, 0, time.UTC), nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var ts TimestampFlag
			err := ts.Set(tc.value)
			if tc.err != err {
				t.Errorf("expected err %s, got %s", tc.err, err)
			}
			if ts.AsTime().String() != tc.expected.String() {
				t.Errorf("expected time %s, got %s", tc.expected.String(), ts.String())
			}
		})
	}

}
