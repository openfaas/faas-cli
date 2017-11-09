package commands

import (
	"testing"
)

var testcases = []struct {
	testcase_name string
	input         string
	output        string
}{
	{
		testcase_name: "Testcase 1",
		input:         "",
		output:        "template\nbuild",
	},
	{
		testcase_name: "Testcase 2",
		input:         "/path/to/folder\n*.pyc\n.DS_STORE",
		output:        "/path/to/folder\n*.pyc\n.DS_STORE\ntemplate\nbuild",
	},
	{
		testcase_name: "Testcase 3",
		input:         "/path/to/folder\ntemplate\n*.pyc\n.DS_STORE",
		output:        "/path/to/folder\ntemplate\n*.pyc\n.DS_STORE\nbuild",
	},
}

func Test_TestCase(t *testing.T) {
	for _, testcase := range testcases {
		output := updateContent(testcase.input)
		if output != testcase.output {
			t.Errorf("[%s] failed", testcase.testcase_name)
		}
	}
}
