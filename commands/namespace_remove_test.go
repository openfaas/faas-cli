package commands

import "testing"

func Test_preDeleteNamespace_NoArgs_Fails(t *testing.T) {
	res := preRemoveNamespace(nil, []string{})

	want := "namespace name required"
	if res.Error() != want {
		t.Errorf("want %q, got %q", want, res.Error())
	}
}

func Test_preDeleteNamespace_MoreThan1Arg_Fails(t *testing.T) {
	res := preRemoveNamespace(nil, []string{
		"secret1",
		"secret2",
	})

	want := "too many values for namespace name"
	if res.Error() != want {
		t.Errorf("want %q, got %q", want, res.Error())
	}
}

func Test_preDeleteNamespace_ExtactlyOneArgIsFine(t *testing.T) {
	res := preRemoveNamespace(nil, []string{
		"namespace1",
	})

	if res != nil {
		t.Errorf("expected no validation error, but got %q", res.Error())
	}
}
