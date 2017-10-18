package proxy

import (
	"net/http"

	"testing"
)

func Test_BasicAuthCanAddIfSuppliedWithNonZeroLengthValues(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://someurl/", nil)
	BasicAuthIfSet(req, "Aladdin", "open sesame")
	header := req.Header.Get("Authorization")
	expected := "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ=="
	if header != expected {
		t.Errorf("got header %q, want %q", header, expected)
	}
}

func Test_BasicAuthWontAddIfNotSuppliedWithNonZeroLengthValues(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://someurl/", nil)
	BasicAuthIfSet(req, "", "")
	header := req.Header.Get("Authorization")
	expected := ""
	if header != expected {
		t.Errorf("got header %q, want %q", header, expected)
	}
}
