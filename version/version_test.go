package version

import (
	"fmt"
	"testing"
)

func Test_BuildVersion(t *testing.T) {
	testcases := []struct {
		version  string
		expected string
	}{
		{version: "", expected: "dev"},
		{version: "testing-manual", expected: "testing-manual"},
	}
	for _, testcase := range testcases {
		t.Run(fmt.Sprintf("version '%s'", testcase.version), func(t *testing.T) {
			Version = testcase.version
			actual := BuildVersion()
			if actual != testcase.expected {
				t.Fatalf("expected: %s, actual: %s\n", testcase.expected, actual)
			}
		})
	}
}

func Test_CompareVersion(t *testing.T) {
	testcases := []struct {
		old      string
		new      string
		expected int
	}{
		{old: "1.2.3", new: "1.2.3", expected: 0},
		{old: "v1.2.3", new: "1.2.3", expected: 0},
		{old: "1.2.3", new: "v1.2.3", expected: 0},
		{old: "2017.03.13.a", new: "2017.03.13.a", expected: 0},

		{old: "1.2.3", new: "1.2.4", expected: 1},
		{old: "1.2.3", new: "1.3.3", expected: 1},
		{old: "1.2.3", new: "2.2.3", expected: 1},
		{old: "1.2.3", new: "1.2.3.1", expected: 1},
		{old: "1.2.3", new: "1.2.3.1.2", expected: 1},
		{old: "1.2.3.b", new: "1.2.3.c", expected: 1},
		{old: "1.2.3.b", new: "1.2.3.c.d.e", expected: 1},

		{old: "2017.03.15", new: "2017.03.13", expected: -1},
		{old: "2017.05.13", new: "2017.03.13", expected: -1},
		{old: "2018.03.13", new: "2017.03.13", expected: -1},
		{old: "2017.03.13.1", new: "2017.03.13", expected: -1},
		{old: "2017.03.13.1231", new: "2017.03.13", expected: -1},
		{old: "2017.03.13b", new: "2017.03.13", expected: -1},
		{old: "2017.03.13.a.b.c", new: "2017.03.13", expected: -1},
	}
	for _, testcase := range testcases {
		t.Run(fmt.Sprintf("1st version '%s' - 2nd version '%s'", testcase.old, testcase.new), func(t *testing.T) {
			actual := CompareVersion(testcase.old, testcase.new)
			if actual != testcase.expected {
				t.Fatalf("expected: %d, actual: %d\n", testcase.expected, actual)
			}
		})
	}
}
