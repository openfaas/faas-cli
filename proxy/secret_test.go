package proxy

import (
	"context"
	"net/http"
	"regexp"
	"testing"

	"github.com/openfaas/faas-cli/test"
	types "github.com/openfaas/faas-provider/types"
)

func Test_GetSecretList_200OK(t *testing.T) {
	s := test.MockHttpServer(t, []test.Request{
		{
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       expectedSecretList,
		},
	})
	defer s.Close()
	client := NewClient(NewTestAuth(nil), s.URL, nil, nil)
	secrets, err := client.GetSecretList(context.Background(), "openfaas-fn")
	if err != nil {
		t.Errorf("Error returned: %s", err.Error())
	}

	for k, v := range secrets {
		if expectedSecretList[k] != v {
			t.Fatalf("Expeceted: %#v - Actual: %#v", expectedListFunctionsResponse[k], v)
		}
	}
}

func Test_GetSecretList_202Accepted(t *testing.T) {
	s := test.MockHttpServer(t, []test.Request{
		{
			ResponseStatusCode: http.StatusAccepted,
			ResponseBody:       expectedSecretList,
		},
	})
	defer s.Close()
	client := NewClient(NewTestAuth(nil), s.URL, nil, nil)
	secrets, err := client.GetSecretList(context.Background(), "")
	if err != nil {
		t.Errorf("Error returned: %s", err.Error())
	}

	for k, v := range secrets {
		if expectedSecretList[k] != v {
			t.Fatalf("Expeceted: %#v - Actual: %#v", expectedListFunctionsResponse[k], v)
		}
	}
}

func Test_GetSecretList_Not200(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusBadRequest)
	client := NewClient(NewTestAuth(nil), s.URL, nil, nil)
	_, err := client.GetSecretList(context.Background(), "openfaas-fn")

	if err == nil {
		t.Fatalf("Error was not returned")
	}

	r := regexp.MustCompile(`(?m:server returned unexpected status code)`)
	if !r.MatchString(err.Error()) {
		t.Fatalf("Error not matched: %s", err)
	}
}

func Test_GetSecretList_Unauthorized401(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusUnauthorized)
	client := NewClient(NewTestAuth(nil), s.URL, nil, nil)
	_, err := client.GetSecretList(context.Background(), "fn")

	if err == nil {
		t.Fatalf("Error was not returned")
	}

	r := regexp.MustCompile(`(?m:unauthorized access, run \"faas-cli login\" to setup authentication for this server)`)
	if !r.MatchString(err.Error()) {
		t.Fatalf("Error not matched: %s", err)
	}
}

var expectedSecretList = []types.Secret{
	{
		Name: "Secret1",
	},
	{
		Name: "Secret2",
	},
}

func Test_CreateSecret_200OK(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusOK)
	secret := types.Secret{
		Name:      "secret-name",
		Value:     "secret-value",
		Namespace: "openfaas-fn",
	}
	client := NewClient(NewTestAuth(nil), s.URL, nil, nil)
	status, _ := client.CreateSecret(context.Background(), secret)

	if status != http.StatusOK {
		t.Errorf("expected: %d, got: %d", http.StatusOK, status)
	}
}

func Test_CreateSecret_201Created(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusCreated)
	secret := types.Secret{
		Name:      "secret-name",
		Value:     "secret-value",
		Namespace: "openfaas-fn",
	}
	client := NewClient(NewTestAuth(nil), s.URL, nil, nil)
	status, _ := client.CreateSecret(context.Background(), secret)

	if status != http.StatusCreated {
		t.Errorf("expected: %d, got: %d", http.StatusCreated, status)
	}
}

func Test_CreateSecret_202Accepted(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusAccepted)
	secret := types.Secret{
		Name:      "secret-name",
		Value:     "secret-value",
		Namespace: "openfaas-fn",
	}
	client := NewClient(NewTestAuth(nil), s.URL, nil, nil)
	status, _ := client.CreateSecret(context.Background(), secret)

	if status != http.StatusAccepted {
		t.Errorf("expected: %d, got: %d", http.StatusAccepted, status)
	}
}

func Test_CreateSecret_Not200(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusBadRequest)

	secret := types.Secret{
		Name:      "secret-name",
		Value:     "secret-value",
		Namespace: "openfaas-fn",
	}
	client := NewClient(NewTestAuth(nil), s.URL, nil, nil)
	status, output := client.CreateSecret(context.Background(), secret)

	if status != http.StatusBadRequest {
		t.Errorf("expected: %d, got: %d", http.StatusBadRequest, status)
	}

	r := regexp.MustCompile(`(?m:server returned unexpected status code)`)
	if !r.MatchString(output) {
		t.Fatalf("Error not matched: %s", output)
	}
}

func Test_CreateSecret_Unauthorized401(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusUnauthorized)

	secret := types.Secret{
		Name:      "secret-name",
		Value:     "secret-value",
		Namespace: "openfaas-fn",
	}
	client := NewClient(NewTestAuth(nil), s.URL, nil, nil)
	status, output := client.CreateSecret(context.Background(), secret)

	if status != http.StatusUnauthorized {
		t.Errorf("expected: %d, got: %d", http.StatusUnauthorized, status)
	}

	r := regexp.MustCompile(`(?m:unauthorized access, run \"faas-cli login\" to setup authentication for this server)`)
	if !r.MatchString(output) {
		t.Fatalf("Error not matched: %s", output)
	}
}

func Test_CreateSecret_Conflict409(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusConflict)

	secret := types.Secret{
		Name:      "secret-name",
		Value:     "secret-value",
		Namespace: "openfaas-fn",
	}
	client := NewClient(NewTestAuth(nil), s.URL, nil, nil)
	status, output := client.CreateSecret(context.Background(), secret)

	if status != http.StatusConflict {
		t.Errorf("want: %d, got: %d", http.StatusConflict, status)
	}

	r := regexp.MustCompile(`(?m:secret with the name "` + secret.Name + `" already exists)`)
	if !r.MatchString(output) {
		t.Fatalf("Error not matched: %s", output)
	}
}

func Test_CreateSecret_ForbiddenNamespace(t *testing.T) {
	s := test.MockHttpServerStatus(t, http.StatusBadRequest)

	secret := types.Secret{
		Name:      "secret-name",
		Value:     "secret-value",
		Namespace: "kube-system",
	}
	client := NewClient(NewTestAuth(nil), s.URL, nil, nil)
	status, output := client.CreateSecret(context.Background(), secret)

	if status != http.StatusBadRequest {
		t.Errorf("want: %d, got: %d", http.StatusBadRequest, status)
	}

	r := regexp.MustCompile(`(?m:server returned unexpected status code: 400)`)
	if !r.MatchString(output) {
		t.Fatalf("Error not matched: %s", output)
	}
}
