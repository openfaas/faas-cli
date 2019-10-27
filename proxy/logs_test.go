package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openfaas/faas-cli/test"
	"github.com/openfaas/faas-provider/logs"
)

type testTokenAuth struct {
	Token string
}

func (t *testTokenAuth) Set(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+t.Token)
	return nil
}

func Test_GetLogs_TokenAuth(t *testing.T) {
	expectedToken := "abc123"
	params := logs.Request{Name: "testFunc"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenValue := r.Header.Get("Authorization")
		if tokenValue != "Bearer "+expectedToken {
			t.Fatalf("Expected header token %v, got %v", expectedToken, tokenValue)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(&testTokenAuth{expectedToken}, srv.URL, nil, nil)
	_, err := client.GetLogs(context.Background(), params)
	if err != nil {
		t.Fatalf("Error returned: %s", err.Error())
	}

}

func Test_GetLogs_200OK(t *testing.T) {

	params := logs.Request{Name: "testFunc"}

	lines := []logs.Message{
		logs.Message{Name: params.Name, Text: "test"},
		logs.Message{Name: params.Name, Text: "test2"},
	}

	s := test.MockHttpServer(t, []test.Request{
		{
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       logRespBody(lines...),
		},
	})
	defer s.Close()

	testAuth := NewTestAuth(nil)
	client := NewClient(testAuth, s.URL, nil, nil)
	logs, err := client.GetLogs(context.Background(), params)
	if err != nil {
		t.Errorf("Error returned: %s", err.Error())
	}

	var sent int
	for line := range logs {
		expected := lines[sent]
		if expected.Text != line.Text {
			t.Fatalf("Expeceted: %#v - Actual: %#v", expected.Text, line.Text)
		}
		sent++
	}
}

func Test_GetLogs_401Unauthorized(t *testing.T) {

	s := test.MockHttpServer(t, []test.Request{
		{
			ResponseStatusCode: http.StatusUnauthorized,
			ResponseBody:       "not allowed",
		},
	})
	defer s.Close()

	params := logs.Request{Name: "test"}

	testAuth := NewTestAuth(nil)
	client := NewClient(testAuth, s.URL, nil, nil)
	_, err := client.GetLogs(context.Background(), params)
	if err == nil {
		t.Fatal("Expected error, got: nil")
	}

	if err.Error() != "unauthorized access, run \"faas-cli login\" to setup authentication for this server" {
		t.Fatalf("Expected unauthorized error, got: %#v", err)
	}
}

func Test_GetLogs_UnexpectedStatus(t *testing.T) {

	cases := []int{
		http.StatusBadRequest, http.StatusForbidden, http.StatusInternalServerError,
	}

	for _, v := range cases {
		s := test.MockHttpServer(t, []test.Request{
			{
				ResponseStatusCode: v,
				ResponseBody:       "bad request, try again",
			},
		})
		defer s.Close()

		testAuth := NewTestAuth(nil)
		client := NewClient(testAuth, s.URL, nil, nil)
		_, err := client.GetLogs(context.Background(), logs.Request{Name: "test"})
		if err == nil {
			t.Fatal("Expected error, got: nil")
		}

		expectedErr := fmt.Sprintf("server returned unexpected status code: %d - bad request, try again", v)
		if err.Error() != expectedErr {
			t.Fatalf("Expected %#v, got: %#v", expectedErr, err)
		}
	}
}

// create new-line delimited json string to treat as a logs response body
func logRespBody(messages ...logs.Message) string {
	var s strings.Builder

	e := json.NewEncoder(&s)
	for _, m := range messages {
		e.Encode(m)
	}

	return s.String()
}
