package proxy

import (
	"net/http"
	"testing"
)

func Test_AddUserAgentAddsUserAgentFromVersion(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://0.0.0.0/", nil)
	userAgentBefore := req.Header.Get("User-Agent")
	AddUserAgent(req)
	userAgentAfter := req.Header.Get("User-Agent")
	if userAgentAfter == "" {
		t.Fatalf("There is No User-Agent Header:\n")
	}
	if userAgentAfter == userAgentBefore {
		t.Fatalf("User-Agent Header Not Changed:\n")
	}
	if userAgentAfter != getUserAgent() {
		t.Fatalf("User-Agent Header Does Not Match - want: '%s' got:'%s'\n", userAgentAfter, getUserAgent())
	}
}
