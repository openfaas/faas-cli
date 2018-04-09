package commands

import (
	"testing"
)

func Test_parseMap_ValidParts(t *testing.T) {
	mapped, err := parseMap([]string{"k=v"}, "foo")

	if err != nil {
		t.Errorf("err was supposed to be nil but was: %s", err.Error())
		t.Fail()
	}

	if mapped["k"] != "v" {
		t.Errorf("value for 'k', want: %s got: %s", "v", mapped["k"])
		t.Fail()
	}
}

func Test_parseMap_NoSeparator(t *testing.T) {
	_, err := parseMap([]string{"kv"}, "foo")

	want := "label format is not correct, needs key=value"
	if err != nil && err.Error() != want {
		t.Errorf("Expected an error due to missing seperator, want: %s got: %s", want, err.Error())
		t.Fail()
	}
}

func Test_parseMap_EmptyKey(t *testing.T) {
	_, err := parseMap([]string{"=v"}, "foo")

	want := "empty foo name: [=v]"
	if err == nil {
		t.Errorf("Expected an error due to missing key")
		t.Fail()
	} else if err.Error() != want {
		t.Errorf("missing key error, want: %s got: %s", want, err.Error())
		t.Fail()
	}
}

func Test_parseMap_MultipleSeparators(t *testing.T) {
	mapped, err := parseMap([]string{"k=v=z"}, "foo")

	if err != nil {
		t.Errorf("Expected second separator to be included in value")
		t.Fail()
	}

	if mapped["k"] != "v=z" {
		t.Errorf("value for 'k', want: %s got: %s", "v=z", mapped["k"])
		t.Fail()
	}
}
