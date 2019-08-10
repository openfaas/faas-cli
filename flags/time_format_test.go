package flags

import (
	"testing"
)

func TestTimeFormat(t *testing.T) {

	cases := []struct {
		name     string
		value    string
		expected string
	}{
		{"can parse short name rfc850", "rfc850", "Monday, 02-Jan-06 15:04:05 MST"},
		{"can accept an arbitrary format string", "2006-01-02 15:04:05.999999999 -0700 MST", "2006-01-02 15:04:05.999999999 -0700 MST"},
		{"can accept arbitrary string", "nonsense", "nonsense"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var f TimeFormat
			err := f.Set(tc.value)
			if err != nil {
				t.Fatalf("should not be able to error")
			}
			if f.String() != tc.expected {
				t.Errorf("expected time %s, got %s", tc.expected, f.String())
			}
		})
	}

}
