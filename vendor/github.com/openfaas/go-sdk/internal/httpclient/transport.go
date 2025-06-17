package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
)

// FaasTransport is an http.RoundTripper that adds default headers and request logging capabilities.
// Requests will be logged to the console if the FAAS_DEBUG environment variable is set to 1.
type FaasTransport struct {
	// Transport is the underlying HTTP transport to use when making requests.
	// It will default to http.DefaultTransport if nil.
	Transport http.RoundTripper
}

func (t *FaasTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add default user-agent header
	if len(req.Header.Get("User-Agent")) == 0 {
		req.Header.Set("User-Agent", "openfaas-go-sdk")
	}

	// If the FAAS_DEBUG environment variable is set to 1, dump the request to the console
	if os.Getenv("FAAS_DEBUG") == "1" {
		dump, err := DumpRequest(req)
		if err != nil {
			return nil, err
		}

		fmt.Println(dump)
	}

	// Call the underlying transport
	return t.transport().RoundTrip(req)
}

func (t *FaasTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

// WithFaasTransport clones the http.Client and wraps the Transport with a FaasTransport.
// If the provided client is nil, the http.DefaultClient is used.
func WithFaasTransport(client *http.Client) *http.Client {
	if client == nil {
		return &http.Client{
			Transport: &FaasTransport{},
		}
	}

	decoratedClient := &http.Client{}
	decoratedClient.Transport = &FaasTransport{
		Transport: client.Transport,
	}
	decoratedClient.CheckRedirect = client.CheckRedirect
	decoratedClient.Jar = client.Jar
	decoratedClient.Timeout = client.Timeout

	return decoratedClient
}

func DumpRequest(req *http.Request) (string, error) {
	var sb strings.Builder

	// Get all header keys and sort them
	keys := make([]string, 0, len(req.Header))
	for k := range req.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sb.WriteString(fmt.Sprintf("%s %s\n", req.Method, req.URL.String()))
	for _, k := range keys {
		v := req.Header[k]
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

	contentType := req.Header.Get("Content-Type")
	if req.Body != nil && isPrintableContentType(contentType) {
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
	}

	return sb.String(), nil
}

func isPrintableContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)

	if strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/plain") ||
		strings.Contains(contentType, "application/x-www-form-urlencoded") {
		return true
	}

	return false
}
