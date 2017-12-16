package commands

import "testing"

func Test_runLogin_NoFlags(t *testing.T) {
	err := runLogin(nil, nil)
	if err != ErrorMissingUsername {
		t.Errorf("'%s' is not the expected error '%s'", err, ErrorMissingUsername)
	}
}

func Test_runLogin_MissingPassword(t *testing.T) {
	username = "username_test"
	err := runLogin(nil, nil)
	if err != ErrorMissingPassword {
		t.Errorf("'%s' is not the expected error '%s'", err, ErrorMissingPassword)
	}
}

func Test_runLogin_PasswordStdinAndPasswordGiven(t *testing.T) {
	username = "username_test"
	password = "password_test"
	passwordStdin = true
	err := runLogin(nil, nil)
	if err != ErrorPasswordStdinAndPassword {
		t.Errorf("'%s' is not the expected error '%s'", err, ErrorPasswordStdinAndPassword)
	}
}
