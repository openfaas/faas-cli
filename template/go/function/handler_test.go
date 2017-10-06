package function

import "testing"

func TestHandleReturnsCorrectResponse(t *testing.T) {
	expected := "Hello, Go. You said: Hello World"
	resp := Handle([]byte("Hello World"))

	if resp != expected {
		t.Fatalf("Expected: %v, Got: %v", expected, resp)
	}
}
