package commands

import (
	"testing"
)

func Test_ValidShell(t *testing.T) {

	testArgs := [][]string{
		{"completion", "--shell", "bash"},
		{"completion", "--shell", "zsh"},
	}

	for _, arg := range testArgs {
		faasCmd.SetArgs(arg)

		err := faasCmd.Execute()

		if err != nil {
			t.Errorf("err was supposed to be nil but it was: %s", err)
			t.Fail()
		}
	}
}
