package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type Request struct {
	Method             string
	Uri                string
	ResponseStatusCode int
	ResponseBody       interface{}
}

type server struct {
	URL                string // Shortcut to httptest.Server.URL
	server             *httptest.Server
	requestCounter     int
	nbExpectedRequests int
	t                  *testing.T
}

// MockHttpServer creates a test server which will send responses in the given order
// It is possible to check on Method and Uri if set
// Responses can contain JSON-encoded body if ResponseBody is set
func MockHttpServer(t *testing.T, requests []Request) *server {
	s := server{
		requestCounter:     0,
		nbExpectedRequests: len(requests),
		t:                  t,
	}

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request Request
		request, requests = requests[0], requests[1:]

		if len(request.Method) > 0 && r.Method != request.Method {
			t.Fatalf(
				"Request n° %d: Expected Method '%s' but got '%s'",
				s.requestCounter+1,
				request.Method,
				r.Method,
			)
		}

		if len(request.Uri) > 0 && r.RequestURI != request.Uri {
			t.Fatalf(
				"Request n° %d: Expected Uri '%s' but got '%s'",
				s.requestCounter+1,
				request.Uri,
				r.RequestURI,
			)
		}

		w.Header().Add("Content-Type", "application/json")

		// Status code defaults to 200
		if request.ResponseStatusCode > 0 {
			w.WriteHeader(request.ResponseStatusCode)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		// If the body is set, send back JSON-encoded body
		if request.ResponseBody != nil {
			s, ok := request.ResponseBody.(string)
			if !ok {
				if b, err := json.Marshal(request.ResponseBody); err != nil {
					t.Fatal(err)
				} else {
					w.Write(b)
				}
			} else {
				w.Write([]byte(s))
			}
		}

		s.requestCounter++
	}))

	s.URL = s.server.URL

	return &s
}

// MockHttpServerStatus creates a test server which will send empty responses with the given status code
// the responses which will be sent are in the given order
func MockHttpServerStatus(t *testing.T, statusCode ...int) *server {
	var requests []Request
	for _, s := range statusCode {
		requests = append(requests, Request{
			ResponseStatusCode: s,
		})
	}

	return MockHttpServer(t, requests)
}

// Close closes the test server
func (s *server) Close() {
	s.server.Close()

	s.assertNbRequests()
}

// assertNbRequests verify if the number of received requests matches the expected number
func (s *server) assertNbRequests() {
	if s.nbExpectedRequests != s.requestCounter {
		s.t.Fatalf(
			"Expected %d requests but received %d",
			s.nbExpectedRequests,
			s.requestCounter,
		)
	}
}

func CaptureStdout(f func()) string {
	stdOut := os.Stdout
	r, w, _ := os.Pipe()
	defer r.Close()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = stdOut

	var b bytes.Buffer
	io.Copy(&b, r)

	return b.String()
}
