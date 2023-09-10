package commands

import "testing"

func Test_preCreateNamespace_NoArgs_Fails(t *testing.T) {
	res := preCreateNamespace(nil, []string{})

	want := "namespace name required"
	if res.Error() != want {
		t.Errorf("want %q, got %q", want, res.Error())
	}
}

func Test_preCreateNamespace_MoreThan1Arg_Fails(t *testing.T) {
	res := preCreateNamespace(nil, []string{
		"secret1",
		"secret2",
	})

	want := "too many values for namespace name"
	if res.Error() != want {
		t.Errorf("want %q, got %q", want, res.Error())
	}
}

func Test_preCreateNamespace_ExtactlyOneArgIsFine(t *testing.T) {
	res := preCreateNamespace(nil, []string{
		"namespace1",
	})

	if res != nil {
		t.Errorf("expected no validation error, but got %q", res.Error())
	}
}
