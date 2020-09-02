package proxy

import (
	"context"
	"fmt"
	"net/http"

	"testing"

	"regexp"

	"github.com/openfaas/faas-cli/test"
)

func Test_ScaleFunction(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusAccepted)
	defer s.Close()

	cliAuth := NewTestAuth(nil)
	proxyClient, _ := NewClient(cliAuth, s.URL, nil, &defaultCommandTimeout)

	err := proxyClient.ScaleFunction(context.Background(), "function-to-scale", "", 0)

	if err != nil {
		t.Fatalf("Got Error: %s,", err.Error())
	}
}

func Test_ScaleFunction_404(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusNotFound)
	defer s.Close()

	cliAuth := NewTestAuth(nil)
	proxyClient, _ := NewClient(cliAuth, s.URL, nil, &defaultCommandTimeout)

	err := proxyClient.ScaleFunction(context.Background(), "function-to-scale", "", 0)

	expectedErr := fmt.Errorf("function %s not found", "function-to-scale")
	if err.Error() != expectedErr.Error() {
		t.Fatalf("Want: %s, got: %s", expectedErr.Error(), err.Error())
	}
}

func Test_ScaleFunction_Unauthorized(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusUnauthorized)
	defer s.Close()

	cliAuth := NewTestAuth(nil)
	proxyClient, _ := NewClient(cliAuth, s.URL, nil, &defaultCommandTimeout)

	err := proxyClient.ScaleFunction(context.Background(), "function-to-scale", "", 0)

	expectedErr := fmt.Errorf("unauthorized action, please setup authentication for this server")
	if err.Error() != expectedErr.Error() {
		t.Fatalf("Want: %s, got: %s", expectedErr.Error(), err.Error())
	}
}

func Test_ScaleFunction_Not2xxAnd404(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusInternalServerError)
	defer s.Close()

	cliAuth := NewTestAuth(nil)
	proxyClient, _ := NewClient(cliAuth, s.URL, nil, &defaultCommandTimeout)

	err := proxyClient.DeleteFunction(context.Background(), "function-to-scale", "")

	r := regexp.MustCompile(`(?m:Server returned unexpected status code)`)
	if !r.MatchString(err.Error()) {
		t.Fatalf("Output not matched: %s", err.Error())
	}
}
