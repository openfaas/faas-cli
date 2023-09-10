package commands

import "testing"

func Test_preUpdateNamespace_NoArgs_Fails(t *testing.T) {
	res := preUpdateNamespace(nil, []string{})

	want := "namespace name required"
	if res.Error() != want {
		t.Errorf("want %q, got %q", want, res.Error())
	}
}

func Test_preUpdateNamespace_MoreThan1Arg_Fails(t *testing.T) {
	res := preUpdateNamespace(nil, []string{
		"secret1",
		"secret2",
	})

	want := "too many values for namespace name"
	if res.Error() != want {
		t.Errorf("want %q, got %q", want, res.Error())
	}
}

func Test_preUpdateNamespace_ExtactlyOneArgIsFine(t *testing.T) {
	res := preUpdateNamespace(nil, []string{
		"namespace1",
	})

	if res != nil {
		t.Errorf("expected no validation error, but got %q", res.Error())
	}
}
