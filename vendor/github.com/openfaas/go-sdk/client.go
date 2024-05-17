package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/openfaas/faas-provider/logs"
	"github.com/openfaas/faas-provider/types"
)

// Client is used to manage OpenFaaS and invoke functions
type Client struct {
	// URL of the OpenFaaS gateway
	GatewayURL *url.URL

	// Authentication provider for authenticating request to the OpenFaaS API.
	ClientAuth ClientAuth

	// TokenSource for getting an ID token that can be exchanged for an
	// OpenFaaS function access token to invoke functions.
	FunctionTokenSource TokenSource

	// Http client used for calls to the OpenFaaS gateway.
	client *http.Client

	// OpenFaaS function access token cache for invoking functions.
	fnTokenCache TokenCache
}

// Wrap http request Do function to support debug capabilities
func (s *Client) do(req *http.Request) (*http.Response, error) {
	if os.Getenv("FAAS_DEBUG") == "1" {
		dump, err := dumpRequest(req)
		if err != nil {
			return nil, err
		}

		fmt.Println(dump)
	}

	return s.client.Do(req)
}

// ClientAuth an interface for client authentication.
// to add authentication to the client implement this interface
type ClientAuth interface {
	Set(req *http.Request) error
}

type ClientOption func(*Client)

// WithFunctionTokenSource configures the function token source for the client.
func WithFunctionTokenSource(tokenSource TokenSource) ClientOption {
	return func(c *Client) {
		c.FunctionTokenSource = tokenSource
	}
}

// WithAuthentication configures the authentication provider fot the client.
func WithAuthentication(auth ClientAuth) ClientOption {
	return func(c *Client) {
		c.ClientAuth = auth
	}
}

// WithFunctionTokenCache configures the token cache used by the client to cache access
// tokens for function invocations.
func WithFunctionTokenCache(cache TokenCache) ClientOption {
	return func(c *Client) {
		c.fnTokenCache = cache
	}
}

// NewClient creates a Client for managing OpenFaaS and invoking functions
func NewClient(gatewayURL *url.URL, auth ClientAuth, client *http.Client) *Client {
	return NewClientWithOpts(gatewayURL, client, WithAuthentication(auth))
}

// NewClientWithOpts creates a Client for managing OpenFaaS and invoking functions
// It takes a list of ClientOptions to configure the client.
func NewClientWithOpts(gatewayURL *url.URL, client *http.Client, options ...ClientOption) *Client {
	c := &Client{
		GatewayURL: gatewayURL,

		client: client,
	}

	for _, option := range options {
		option(c)
	}

	if c.ClientAuth != nil && c.FunctionTokenSource == nil {
		// Use auth as the default function token source for IAM function authentication
		// if it implements the TokenSource interface.
		functionTokenSource, ok := c.ClientAuth.(TokenSource)
		if ok {
			c.FunctionTokenSource = functionTokenSource
		}
	}

	return c
}

// GetNamespaces get openfaas namespaces
func (s *Client) GetNamespaces(ctx context.Context) ([]string, error) {
	namespaces := []string{}

	u, _ := url.Parse(s.GatewayURL.String())
	u.Path = "/system/namespaces"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return namespaces, fmt.Errorf("unable to create request: %s, error: %w", u.String(), err)
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return namespaces, fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}

	res, err := s.do(req)
	if err != nil {
		return namespaces, fmt.Errorf("unable to make request: %w", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	bytesOut, err := io.ReadAll(res.Body)
	if err != nil {
		return namespaces, err
	}

	if res.StatusCode == http.StatusUnauthorized {
		return namespaces, fmt.Errorf("check authorization, status code: %d", res.StatusCode)
	}

	if len(bytesOut) == 0 {
		return namespaces, nil
	}

	if err := json.Unmarshal(bytesOut, &namespaces); err != nil {
		return namespaces, fmt.Errorf("unable to marshal to JSON: %s, error: %w", string(bytesOut), err)
	}

	return namespaces, err
}

// GetNamespaces get openfaas namespaces
func (s *Client) GetNamespace(ctx context.Context, namespace string) (types.FunctionNamespace, error) {
	u, _ := url.Parse(s.GatewayURL.String())
	u.Path = fmt.Sprintf("/system/namespace/%s", namespace)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return types.FunctionNamespace{}, fmt.Errorf("unable to create request for %s, error: %w", u.String(), err)
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return types.FunctionNamespace{}, fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}

	res, err := s.do(req)
	if err != nil {
		return types.FunctionNamespace{}, fmt.Errorf("unable to make HTTP request: %w", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}
	body, _ := io.ReadAll(res.Body)

	switch res.StatusCode {
	case http.StatusOK:
		fnNamespace := types.FunctionNamespace{}
		if err := json.Unmarshal(body, &fnNamespace); err != nil {
			return types.FunctionNamespace{},
				fmt.Errorf("unable to unmarshal value: %q, error: %w", string(body), err)
		}
		return fnNamespace, err

	case http.StatusNotFound:
		return types.FunctionNamespace{}, fmt.Errorf("namespace %s not found", namespace)

	case http.StatusUnauthorized:
		return types.FunctionNamespace{}, fmt.Errorf("unauthorized action, please setup authentication for this server")

	default:
		return types.FunctionNamespace{}, fmt.Errorf("unexpected status code: %d, message: %q", res.StatusCode, string(body))
	}
}

// CreateNamespace creates a namespace
func (s *Client) CreateNamespace(ctx context.Context, spec types.FunctionNamespace) (int, error) {

	// set openfaas label
	if spec.Labels == nil {
		spec.Labels = map[string]string{}
	}
	spec.Labels["openfaas"] = "1"

	// set openfaas annotation
	if spec.Annotations == nil {
		spec.Annotations = map[string]string{}
	}
	spec.Annotations["openfaas"] = "1"

	bodyBytes, err := json.Marshal(spec)
	if err != nil {
		return http.StatusBadRequest, err
	}

	bodyReader := bytes.NewReader(bodyBytes)

	u, _ := url.Parse(s.GatewayURL.String())
	u.Path = "/system/namespace/"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bodyReader)
	if err != nil {
		return http.StatusBadGateway, err
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}

	res, err := s.do(req)
	if err != nil {
		return http.StatusBadGateway, err
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	switch res.StatusCode {
	case http.StatusAccepted, http.StatusOK, http.StatusCreated:
		return res.StatusCode, nil

	case http.StatusUnauthorized:
		return res.StatusCode, fmt.Errorf("unauthorized action, please setup authentication for this server")

	default:
		return res.StatusCode, fmt.Errorf("unexpected status code: %d, message: %q", res.StatusCode, string(body))
	}
}

// UpdateNamespace updates a namespace
func (s *Client) UpdateNamespace(ctx context.Context, spec types.FunctionNamespace) (int, error) {

	// set openfaas label
	if spec.Labels == nil {
		spec.Labels = map[string]string{}
	}
	spec.Labels["openfaas"] = "1"

	// set openfaas annotation
	if spec.Annotations == nil {
		spec.Annotations = map[string]string{}
	}
	spec.Annotations["openfaas"] = "1"

	bodyBytes, err := json.Marshal(spec)
	if err != nil {
		return http.StatusBadRequest, err
	}

	bodyReader := bytes.NewReader(bodyBytes)

	u, _ := url.Parse(s.GatewayURL.String())
	u.Path = fmt.Sprintf("/system/namespace/%s", spec.Name)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), bodyReader)
	if err != nil {
		return http.StatusBadGateway, err
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}

	res, err := s.do(req)
	if err != nil {
		return http.StatusBadGateway, err
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	switch res.StatusCode {
	case http.StatusAccepted, http.StatusOK, http.StatusCreated:
		return res.StatusCode, nil

	case http.StatusNotFound:
		return res.StatusCode, fmt.Errorf("namespace %s not found", spec.Name)

	case http.StatusUnauthorized:
		return res.StatusCode, fmt.Errorf("unauthorized action, please setup authentication for this server")

	default:
		return res.StatusCode, fmt.Errorf("unexpected status code: %d, message: %q", res.StatusCode, string(body))
	}
}

// DeleteNamespace deletes a namespace
func (s *Client) DeleteNamespace(ctx context.Context, namespace string) error {

	delReq := types.FunctionNamespace{
		Name: namespace,
		Labels: map[string]string{
			"openfaas": "1",
		},
	}

	var err error

	bodyBytes, _ := json.Marshal(delReq)
	bodyReader := bytes.NewReader(bodyBytes)

	u, _ := url.Parse(s.GatewayURL.String())
	u.Path = fmt.Sprintf("/system/namespace/%s", namespace)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("cannot connect to OpenFaaS on URL: %s, error: %s", u.String(), err)
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}
	res, err := s.do(req)
	if err != nil {
		return fmt.Errorf("cannot connect to OpenFaaS on URL: %s, error: %s", s.GatewayURL, err)

	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusAccepted, http.StatusOK, http.StatusCreated:
		break

	case http.StatusNotFound:
		return fmt.Errorf("namespace %s not found", namespace)

	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized action, please setup authentication for this server")

	default:
		var err error
		bytesOut, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("unexpected status code: %d, message: %q", res.StatusCode, string(bytesOut))
	}
	return nil
}

// GetFunctions lists all functions
func (s *Client) GetFunctions(ctx context.Context, namespace string) ([]types.FunctionStatus, error) {
	u, _ := url.Parse(s.GatewayURL.String())
	u.Path = "/system/functions"

	if len(namespace) > 0 {
		query := u.Query()
		query.Set("namespace", namespace)
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return []types.FunctionStatus{}, fmt.Errorf("unable to create request for %s, error: %w", u.String(), err)
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return []types.FunctionStatus{}, fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}

	res, err := s.do(req)
	if err != nil {
		return []types.FunctionStatus{}, fmt.Errorf("unable to make HTTP request: %w", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, _ := io.ReadAll(res.Body)

	functions := []types.FunctionStatus{}
	if err := json.Unmarshal(body, &functions); err != nil {
		return []types.FunctionStatus{},
			fmt.Errorf("unable to unmarshal value: %q, error: %w", string(body), err)
	}

	return functions, nil
}

func (s *Client) GetInfo(ctx context.Context) (SystemInfo, error) {
	u, _ := url.Parse(s.GatewayURL.String())
	u.Path = "/system/info"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return SystemInfo{}, fmt.Errorf("unable to create request for %s, error: %w", u.String(), err)
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return SystemInfo{}, fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}

	res, err := s.do(req)
	if err != nil {
		return SystemInfo{}, fmt.Errorf("unable to make HTTP request: %w", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, _ := io.ReadAll(res.Body)

	info := SystemInfo{}
	if err := json.Unmarshal(body, &info); err != nil {
		return SystemInfo{},
			fmt.Errorf("unable to unmarshal value: %q, error: %w", string(body), err)
	}

	return info, nil
}

// GetFunction gives a richer payload than GetFunctions, but for a specific function
func (s *Client) GetFunction(ctx context.Context, name, namespace string) (types.FunctionStatus, error) {
	u, _ := url.Parse(s.GatewayURL.String())
	u.Path = "/system/function/" + name

	if len(namespace) > 0 {
		query := u.Query()
		query.Set("namespace", namespace)
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return types.FunctionStatus{}, fmt.Errorf("unable to create request for %s, error: %w", u.String(), err)
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return types.FunctionStatus{}, fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}

	res, err := s.do(req)
	if err != nil {
		return types.FunctionStatus{}, fmt.Errorf("unable to make HTTP request: %w", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, _ := io.ReadAll(res.Body)

	function := types.FunctionStatus{}
	if err := json.Unmarshal(body, &function); err != nil {
		return types.FunctionStatus{},
			fmt.Errorf("unable to unmarshal value: %q, error: %w", string(body), err)
	}

	return function, nil
}

func (s *Client) Deploy(ctx context.Context, spec types.FunctionDeployment) (int, error) {
	return s.deploy(ctx, http.MethodPost, spec)

}

func (s *Client) Update(ctx context.Context, spec types.FunctionDeployment) (int, error) {
	return s.deploy(ctx, http.MethodPut, spec)
}

func (s *Client) deploy(ctx context.Context, method string, spec types.FunctionDeployment) (int, error) {

	bodyBytes, err := json.Marshal(spec)
	if err != nil {
		return http.StatusBadRequest, err
	}

	bodyReader := bytes.NewReader(bodyBytes)

	u, _ := url.Parse(s.GatewayURL.String())
	u.Path = "/system/functions"

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return http.StatusBadGateway, err
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}

	res, err := s.do(req)
	if err != nil {
		return http.StatusBadGateway, err
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	switch res.StatusCode {
	case http.StatusAccepted, http.StatusOK, http.StatusCreated:
		return res.StatusCode, nil

	case http.StatusUnauthorized:
		return res.StatusCode, fmt.Errorf("unauthorized action, please setup authentication for this server")

	default:
		return res.StatusCode, fmt.Errorf("unexpected status code: %d, message: %q", res.StatusCode, string(body))
	}
}

// ScaleFunction scales a function to a number of replicas
func (s *Client) ScaleFunction(ctx context.Context, functionName, namespace string, replicas uint64) error {

	scaleReq := types.ScaleServiceRequest{
		ServiceName: functionName,
		Replicas:    replicas,
		Namespace:   namespace,
	}

	var err error

	bodyBytes, _ := json.Marshal(scaleReq)
	bodyReader := bytes.NewReader(bodyBytes)

	u, _ := url.Parse(s.GatewayURL.String())
	u.Path = filepath.Join("/system/scale-function", functionName)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("cannot connect to OpenFaaS on URL: %s, error: %s", u.String(), err)
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}
	res, err := s.do(req)
	if err != nil {
		return fmt.Errorf("cannot connect to OpenFaaS on URL: %s, error: %s", s.GatewayURL, err)

	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusAccepted, http.StatusOK, http.StatusCreated:
		break

	case http.StatusNotFound:
		return fmt.Errorf("function %s not found", functionName)

	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized action, please setup authentication for this server")

	default:
		var err error
		bytesOut, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("server returned unexpected status code %d, message: %q", res.StatusCode, string(bytesOut))
	}
	return nil
}

// DeleteFunction deletes a function
func (s *Client) DeleteFunction(ctx context.Context, functionName, namespace string) error {

	delReq := types.DeleteFunctionRequest{
		FunctionName: functionName,
		Namespace:    namespace,
	}

	var err error

	bodyBytes, _ := json.Marshal(delReq)
	bodyReader := bytes.NewReader(bodyBytes)

	u, _ := url.Parse(s.GatewayURL.String())
	u.Path = "/system/functions"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("cannot connect to OpenFaaS on URL: %s, error: %s", u.String(), err)
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}
	res, err := s.do(req)
	if err != nil {
		return fmt.Errorf("cannot connect to OpenFaaS on URL: %s, error: %s", s.GatewayURL, err)

	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusAccepted, http.StatusOK, http.StatusCreated:
		break

	case http.StatusNotFound:
		return fmt.Errorf("function %s not found", functionName)

	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized action, please setup authentication for this server")

	default:
		var err error
		bytesOut, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("server returned unexpected status code %d, message: %q", res.StatusCode, string(bytesOut))
	}
	return nil
}

// GetSecrets list all secrets
func (s *Client) GetSecrets(ctx context.Context, namespace string) ([]types.Secret, error) {
	u, _ := url.Parse(s.GatewayURL.String())
	u.Path = "/system/secrets"

	if len(namespace) > 0 {
		query := u.Query()
		query.Set("namespace", namespace)
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return []types.Secret{}, fmt.Errorf("unable to create request for %s, error: %w", u.String(), err)
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return []types.Secret{}, fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}

	res, err := s.do(req)
	if err != nil {
		return []types.Secret{}, fmt.Errorf("unable to make HTTP request: %w", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, _ := io.ReadAll(res.Body)

	secrets := []types.Secret{}
	if err := json.Unmarshal(body, &secrets); err != nil {
		return []types.Secret{},
			fmt.Errorf("unable to unmarshal value: %q, error: %w", string(body), err)
	}

	return secrets, nil
}

// CreateSecret creates a secret
func (s *Client) CreateSecret(ctx context.Context, spec types.Secret) (int, error) {

	bodyBytes, err := json.Marshal(spec)
	if err != nil {
		return http.StatusBadRequest, err
	}

	bodyReader := bytes.NewReader(bodyBytes)

	u, _ := url.Parse(s.GatewayURL.String())
	u.Path = "/system/secrets"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bodyReader)
	if err != nil {
		return http.StatusBadGateway, err
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}

	res, err := s.do(req)
	if err != nil {
		return http.StatusBadGateway, err
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	switch res.StatusCode {
	case http.StatusAccepted, http.StatusOK, http.StatusCreated:
		return res.StatusCode, nil

	case http.StatusUnauthorized:
		return res.StatusCode, fmt.Errorf("unauthorized action, please setup authentication for this server")

	default:
		return res.StatusCode, fmt.Errorf("unexpected status code: %d, message: %q", res.StatusCode, string(body))
	}
}

// UpdateSecret updates a secret
func (s *Client) UpdateSecret(ctx context.Context, spec types.Secret) (int, error) {

	bodyBytes, err := json.Marshal(spec)
	if err != nil {
		return http.StatusBadRequest, err
	}

	bodyReader := bytes.NewReader(bodyBytes)

	u, _ := url.Parse(s.GatewayURL.String())
	u.Path = "/system/secrets"

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), bodyReader)
	if err != nil {
		return http.StatusBadGateway, err
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}

	res, err := s.do(req)
	if err != nil {
		return http.StatusBadGateway, err
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	switch res.StatusCode {
	case http.StatusAccepted, http.StatusOK, http.StatusCreated:
		return res.StatusCode, nil

	case http.StatusNotFound:
		return res.StatusCode, fmt.Errorf("secret %s not found", spec.Name)

	case http.StatusUnauthorized:
		return res.StatusCode, fmt.Errorf("unauthorized action, please setup authentication for this server")

	default:
		return res.StatusCode, fmt.Errorf("unexpected status code: %d, message: %q", res.StatusCode, string(body))
	}
}

// DeleteSecret deletes a secret
func (s *Client) DeleteSecret(ctx context.Context, secretName, namespace string) error {

	delReq := types.Secret{
		Name:      secretName,
		Namespace: namespace,
	}

	var err error

	bodyBytes, _ := json.Marshal(delReq)
	bodyReader := bytes.NewReader(bodyBytes)

	u, _ := url.Parse(s.GatewayURL.String())
	u.Path = "/system/secrets"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("cannot connect to OpenFaaS on URL: %s, error: %s", u.String(), err)
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}
	res, err := s.do(req)
	if err != nil {
		return fmt.Errorf("cannot connect to OpenFaaS on URL: %s, error: %s", s.GatewayURL, err)

	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusAccepted, http.StatusOK, http.StatusCreated:
		break

	case http.StatusNotFound:
		return fmt.Errorf("secret %s not found", secretName)

	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized action, please setup authentication for this server")

	default:
		var err error
		bytesOut, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("server returned unexpected status code %d, message: %q", res.StatusCode, string(bytesOut))
	}
	return nil
}

func generateLogRequest(functionName, namespace string, follow bool, tail int, since *time.Time) url.Values {
	query := url.Values{}
	query.Add("name", functionName)
	if len(namespace) > 0 {
		query.Add("namespace", namespace)
	}

	if follow {
		query.Add("follow", "1")
	} else {
		query.Add("follow", "0")
	}

	if since != nil {
		query.Add("since", since.Format(time.RFC3339))
	}

	if tail != 0 {
		query.Add("tail", strconv.Itoa(tail))
	}

	return query
}

func (s *Client) GetLogs(ctx context.Context, functionName, namespace string, follow bool, tail int, since *time.Time) (<-chan logs.Message, error) {

	var err error

	u, _ := url.Parse(s.GatewayURL.String())
	u.Path = "/system/logs"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s, error: %s", u.String(), err)
	}

	req.URL.RawQuery = generateLogRequest(functionName, namespace, follow, tail, since).Encode()
	log.Printf("%s", req.URL.RawQuery)

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return nil, fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}

	res, err := s.do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS on URL: %s, error: %s", s.GatewayURL, err)

	}

	logStream := make(chan logs.Message, 1000)

	switch res.StatusCode {
	case http.StatusOK:
		go func() {
			defer func() {
				close(logStream)
			}()

			if res.Body != nil {
				defer res.Body.Close()
			}

			decoder := json.NewDecoder(res.Body)

			for decoder.More() {
				msg := logs.Message{}
				err := decoder.Decode(&msg)
				if err != nil {
					log.Printf("cannot parse log results: %s", err.Error())
					return
				}
				logStream <- msg
			}
		}()

	case http.StatusNotFound:
		return nil, fmt.Errorf("function: %s not found", functionName)

	case http.StatusUnauthorized:
		return nil, fmt.Errorf("unauthorized action, please setup authentication for this server")

	default:
		bytesOut, err := io.ReadAll(res.Body)
		if err == nil {
			return nil, fmt.Errorf("unexpected status code: %d, message: %q", res.StatusCode, string(bytesOut))
		}
	}
	return logStream, nil
}

func dumpRequest(req *http.Request) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s %s\n", req.Method, req.URL.String()))
	for k, v := range req.Header {
		if k == "Authorization" {
			auth := "[REDACTED]"
			if len(v) == 0 {
				auth = "[NOT_SET]"
			} else {
				l, _, ok := strings.Cut(v[0], " ")
				if ok && (l == "Basic" || l == "Bearer") {
					auth = l + " [REDACTED]"
				}
			}
			sb.WriteString(fmt.Sprintf("%s: %s\n", k, auth))

		} else {
			sb.WriteString(fmt.Sprintf("%s: %s\n", k, v))
		}
	}

	if req.Body != nil {
		r := io.NopCloser(req.Body)
		buf := new(strings.Builder)
		_, err := io.Copy(buf, r)
		if err != nil {
			return "", err
		}
		bodyDebug := buf.String()
		if len(bodyDebug) > 0 {
			sb.WriteString(fmt.Sprintf("%s\n", bodyDebug))

		}
		req.Body = io.NopCloser(strings.NewReader(buf.String()))
	}

	return sb.String(), nil
}
