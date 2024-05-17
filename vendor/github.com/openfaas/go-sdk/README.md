## go-sdk

A lightweight Go SDK for use within OpenFaaS functions and to control the OpenFaaS gateway.

For use within any Go code (not just OpenFaaS Functions):

* Client - A client for the OpenFaaS REST API

For use within functions:

* ReadSecret() - Read a named secret from within an OpenFaaS Function
* ReadSecrets() - Read all available secrets returning a queryable map

Authentication helpers (See: [Authentication with IAM](#authentication-with-iam)):

* ServiceAccountTokenSource - An implementation of the TokenSource interface to get an ID token by reading a Kubernetes projected service account token from `/var/secrets/tokens/openfaas-token` or the path set by the `token_mount_path` environment
variable.

## Usage

```go
import "github.com/openfaas/go-sdk"
```

Construct a new OpenFaaS client and use it to access the OpenFaaS gateway API.

```go
gatewayURL, _ := url.Parse("http://127.0.0.1:8080")
auth := &sdk.BasicAuth{
    Username: username,
    Password: password,
}

client := sdk.NewClient(gatewayURL, auth, http.DefaultClient)

namespace, err := client.GetNamespaces(context.Background())
```

### Authentication with IAM

To authenticate with an OpenFaaS deployment that has [Identity and Access Management (IAM)](https://docs.openfaas.com/openfaas-pro/iam/overview/) enabled, the client needs to exchange an ID token for an OpenFaaS ID token.

To get a token that can be exchanged for an OpenFaaS token you need to implement the `TokenSource` interface.

This is an example of a token source that gets a service account token mounted into a pod with [ServiceAccount token volume projection](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#serviceaccount-token-volume-projection).

```go
type ServiceAccountTokenSource struct{}

func (ts *ServiceAccountTokenSource) Token() (string, error) {
	tokenMountPath := getEnv("token_mount_path", "/var/secrets/tokens")
	if len(tokenMountPath) == 0 {
		return "", fmt.Errorf("invalid token_mount_path specified for reading the service account token")
	}

	idTokenPath := path.Join(tokenMountPath, "openfaas-token")
	idToken, err := os.ReadFile(idTokenPath)
	if err != nil {
		return "", fmt.Errorf("unable to load service account token: %s", err)
	}

	return string(idToken), nil
}
```

The service account token returned by the `TokenSource` is automatically exchanged for an OpenFaaS token that is then used in the Authorization header for all requests made to the API.

If the OpenFaaS token is expired the `TokenSource` is asked for a token and the token exchange will run again.

```go
gatewayURL, _ := url.Parse("https://gw.openfaas.example.com")

auth := &sdk.TokenAuth{
    TokenURL "https://gw.openfaas.example.com/oauth/token",
    TokenSource: &ServiceAccountTokenSource{}
}

client := sdk.NewClient(gatewayURL, auth, http.DefaultClient)
```

### Authentication with Federated Gateway

```go
func Test_ClientCredentials(t *testing.T) {
	clientID := ""
	clientSecret := ""
	tokenURL := "https://keycloak.example.com/realms/openfaas/protocol/openid-connect/token"
	scope := "email"
	grantType := "client_credentials"

	audience = "" // Optional

	auth := NewClientCredentialsTokenSource(clientID, clientSecret, tokenURL, scope, grantType, audience)

	token, err := auth.Token()
	if err != nil {
		t.Fatal(err)
	}

	if token == "" {
		t.Fatal("token is empty")
	}

	u, _ := url.Parse("https://fed-gw.example.com")

	client := NewClient(u, &ClientCredentialsAuth{tokenSource: auth}, http.DefaultClient)

	fns, err := client.GetFunctions(context.Background(), "openfaas-fn")
	if err != nil {
		t.Fatal(err)
	}

	if len(fns) == 0 {
		t.Fatal("no functions found")
	}
}
```

## Deploy Function
```go

status, err := client.Deploy(context.Background(), types.FunctionDeployment{
	Service:    "env-store-test",
	Image:      "ghcr.io/openfaas/alpine:latest",
	Namespace:  "openfaas-fn",
	EnvProcess: "env",
})

// non 200 status value will have some error
if err != nil {
	log.Printf("Deploy Failed: %s", err)
}
```

## Delete Function
```go

err := client.DeleteFunction(context.Background(),"env-store-test", "openfaas-fn")
if err != nil {
	log.Printf("Deletion Failed: %s", err)
}
```

Please refer [examples](https://github.com/openfaas/go-sdk/tree/master/examples) folder for code examples of each operation

## Invoke functions

```go
body := strings.NewReader("OpenFaaS")
req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, "/", body)
if err != nil {
	panic(err)
}

req.Header.Set("Content-Type", "text/plain")

async := false
authenticate := false

// Make a POST request to a figlet function in the openfaas-fn namespace
res, err := client.InvokeFunction(context.Background(), "figlet", "openfaas-fn", async, authenticate, req)
if err != nil {
	log.Printf("Failed to invoke function: %s", err)
	return
}

if res.Body != nil {
	defer res.Body.Close()
}

// Read the response body
body, err := io.ReadAll(res.Body)
if err != nil {
	log.Printf("Error reading response body: %s", err)
	return
}

// Print the response
fmt.Printf("Response status code: %s\n", res.Status)
fmt.Printf("Response body: %s\n", string(body))
```

### Authenticate function invocations

The SDK supports invoking functions if you are using OpenFaaS IAM with [built-in authentication for functions](https://www.openfaas.com/blog/built-in-function-authentication/).

Set the `auth` argument to `true` when calling `InvokeFunction` to authenticate the request with an OpenFaaS function access token.

The `Client` needs a `TokenSource` to get an ID token that can be exchanged for a function access token to make authenticated function invocations. By default the `TokenAuth` provider that was set when constructing a new `Client` is used.

It is also possible to provide a custom `TokenSource` for the function token exchange:

```go
ts := sdk.NewClientCredentialsTokenSource(clientID, clientSecret, tokenURL, scope, grantType, audience)

client := sdk.NewClientWithOpts(gatewayURL, http.DefaultClient, sdk.WithFunctionTokenSource(ts))
```

Optionally a `TokenCache` can be configured to cache function access tokens and prevent the client from having to do a token exchange each time a function is invoked.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

fnTokenCache := sdk.NewMemoryTokenCache()
// Start garbage collection to remove expired tokens from the cache.
go fnTokenCache.StartGC(ctx, time.Second*10)

client := sdk.NewClientWithOpts(
    gatewayUrl,
    httpClient,
    sdk.WithAuthentication(auth),
    sdk.WithFunctionTokenCache(fnTokenCache),
)
```

## License

License: MIT
