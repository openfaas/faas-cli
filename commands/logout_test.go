package commands

import "testing"

func Test_runLogout_NoFlags(t *testing.T) {
	gateway = ""
	err := runLogout(nil, nil)
	if err != ErrorMissingGateway {
		t.Errorf("'%s' is not the expected error '%s'", err, ErrorMissingGateway)
	}
}
