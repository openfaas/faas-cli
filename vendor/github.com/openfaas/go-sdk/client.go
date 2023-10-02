package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/openfaas/faas-provider/types"
)

// Client is used to manage OpenFaaS functions
type Client struct {
	GatewayURL *url.URL
	Client     *http.Client
	ClientAuth ClientAuth
}

// ClientAuth an interface for client authentication.
// to add authentication to the client implement this interface
type ClientAuth interface {
	Set(req *http.Request) error
}

// NewClient creates an Client for managing OpenFaaS
func NewClient(gatewayURL *url.URL, auth ClientAuth, client *http.Client) *Client {
	return &Client{
		GatewayURL: gatewayURL,
		Client:     client,
		ClientAuth: auth,
	}
}

// GetNamespaces get openfaas namespaces
func (s *Client) GetNamespaces(ctx context.Context) ([]string, error) {
	u := s.GatewayURL
	namespaces := []string{}
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

	res, err := s.Client.Do(req)
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
	u := s.GatewayURL
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

	res, err := s.Client.Do(req)
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

	u := s.GatewayURL
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

	res, err := s.Client.Do(req)
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

	u := s.GatewayURL
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

	res, err := s.Client.Do(req)
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

	u := s.GatewayURL
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
	res, err := http.DefaultClient.Do(req)
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
	u := s.GatewayURL

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

	res, err := s.Client.Do(req)
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
	u := s.GatewayURL

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

	res, err := s.Client.Do(req)
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
	u := s.GatewayURL

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

	res, err := s.Client.Do(req)
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

	u := s.GatewayURL
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

	res, err := s.Client.Do(req)
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

	u := s.GatewayURL

	functionPath := filepath.Join("/system/scale-function", functionName)

	u.Path = functionPath

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("cannot connect to OpenFaaS on URL: %s, error: %s", u.String(), err)
	}

	if s.ClientAuth != nil {
		if err := s.ClientAuth.Set(req); err != nil {
			return fmt.Errorf("unable to set Authorization header: %w", err)
		}
	}
	res, err := http.DefaultClient.Do(req)
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

	u := s.GatewayURL
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
	res, err := http.DefaultClient.Do(req)
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
