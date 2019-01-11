// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"testing"
)

func Test_preRunSecretRemoveCmd_NoArgs_Fails(t *testing.T) {

	res := preRunSecretRemoveCmd(nil, []string{})

	want := "secret name required"
	if res.Error() != want {
		t.Errorf("want %q, got %q", want, res.Error())
	}
}

func Test_preRunSecretRemoveCmd_MoreThan1Arg_Fails(t *testing.T) {

	res := preRunSecretRemoveCmd(nil, []string{
		"secret1",
		"secret2",
	})

	want := "too many values for secret name"
	if res.Error() != want {
		t.Errorf("want %q, got %q", want, res.Error())
	}
}

func Test_preRunSecretRemoveCmd_ExtactlyOneArgIsFine(t *testing.T) {

	res := preRunSecretRemoveCmd(nil, []string{
		"secret1",
	})

	if res != nil {
		t.Errorf("expected no validation error, but got %q", res.Error())
	}
}
