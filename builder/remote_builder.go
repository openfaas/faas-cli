package builder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	sdkbuilder "github.com/openfaas/go-sdk/builder"
)

const naclBoxAlgorithm = "nacl/box"

type remoteBuilderPublicKeyResponse struct {
	KeyID     string `json:"key_id"`
	Algorithm string `json:"algorithm"`
	PublicKey string `json:"public_key"`
}

func runRemoteBuild(builderURL *url.URL, tarPath, payloadSecretPath, builderPublicKeyPath, builderKeyID string, buildSecrets map[string]string, quietBuild bool, functionName, imageName string) error {
	payloadSecret, err := os.ReadFile(payloadSecretPath)
	if err != nil {
		return fmt.Errorf("failed to read payload secret: %w", err)
	}
	payloadSecret = bytes.TrimSpace(payloadSecret)

	opts := []sdkbuilder.BuilderOption{
		sdkbuilder.WithHmacAuth(string(payloadSecret)),
	}

	if len(buildSecrets) > 0 {
		publicKey, err := resolveRemoteBuilderPublicKey(builderURL, builderPublicKeyPath, builderKeyID)
		if err != nil {
			return err
		}
		opts = append(opts, sdkbuilder.WithBuildSecretsKey(publicKey.KeyID, []byte(publicKey.PublicKey)))
	}

	b := sdkbuilder.NewFunctionBuilder(builderURL, http.DefaultClient, opts...)

	var stream *sdkbuilder.BuildResultStream
	if len(buildSecrets) > 0 {
		stream, err = b.BuildWithSecretsStream(tarPath, buildSecrets)
	} else {
		stream, err = b.BuildWithStream(tarPath)
	}
	if err != nil {
		return err
	}
	defer stream.Close()

	return consumeBuildStream(stream, quietBuild, functionName, imageName)
}

func resolveRemoteBuilderPublicKey(builderURL *url.URL, builderPublicKeyPath, builderKeyID string) (*remoteBuilderPublicKeyResponse, error) {
	if builderPublicKeyPath == "" {
		return fetchRemoteBuilderPublicKey(builderURL)
	}

	publicKeyData, err := readBuilderPublicKeyInput(builderPublicKeyPath)
	if err != nil {
		return nil, err
	}

	publicKey, err := parseRemoteBuilderPublicKey(publicKeyData)
	if err != nil {
		return nil, err
	}

	if builderKeyID != "" {
		publicKey.KeyID = builderKeyID
	}

	if publicKey.KeyID == "" {
		return nil, fmt.Errorf("builder key id is required when using a pinned builder public key")
	}

	return publicKey, nil
}

func readBuilderPublicKeyInput(value string) ([]byte, error) {
	info, err := os.Stat(value)
	if err == nil {
		if info.IsDir() {
			return nil, fmt.Errorf("builder public key path %q is a directory", value)
		}

		data, readErr := os.ReadFile(value)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read builder public key: %w", readErr)
		}

		return data, nil
	}

	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to stat builder public key %q: %w", value, err)
	}

	return []byte(value), nil
}

func fetchRemoteBuilderPublicKey(builderURL *url.URL) (*remoteBuilderPublicKeyResponse, error) {
	reqURL := builderURL.JoinPath("/publickey")

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed to fetch builder public key, status code %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	publicKey := remoteBuilderPublicKeyResponse{}
	if err := json.NewDecoder(res.Body).Decode(&publicKey); err != nil {
		return nil, err
	}

	algorithm := publicKey.Algorithm
	if algorithm == "" {
		algorithm = naclBoxAlgorithm
	}
	if algorithm != naclBoxAlgorithm {
		return nil, fmt.Errorf("unsupported encrypted build secrets algorithm: %s", publicKey.Algorithm)
	}
	if publicKey.PublicKey == "" {
		return nil, fmt.Errorf("builder public key response did not include a public key")
	}

	return &publicKey, nil
}

func parseRemoteBuilderPublicKey(data []byte) (*remoteBuilderPublicKeyResponse, error) {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil, fmt.Errorf("builder public key file is empty")
	}

	if strings.HasPrefix(trimmed, "{") {
		publicKey := remoteBuilderPublicKeyResponse{}
		if err := json.Unmarshal([]byte(trimmed), &publicKey); err != nil {
			return nil, fmt.Errorf("failed to parse builder public key JSON: %w", err)
		}

		algorithm := publicKey.Algorithm
		if algorithm == "" {
			algorithm = naclBoxAlgorithm
		}
		if algorithm != naclBoxAlgorithm {
			return nil, fmt.Errorf("unsupported encrypted build secrets algorithm: %s", publicKey.Algorithm)
		}
		if publicKey.PublicKey == "" {
			return nil, fmt.Errorf("builder public key JSON did not include a public key")
		}

		return &publicKey, nil
	}

	return &remoteBuilderPublicKeyResponse{
		PublicKey: trimmed,
	}, nil
}

func consumeBuildStream(stream *sdkbuilder.BuildResultStream, quietBuild bool, functionName, imageName string) error {
	for result, err := range stream.Results() {
		if err != nil {
			return err
		}
		if !quietBuild {
			for _, logMsg := range result.Log {
				fmt.Printf("%s\n", logMsg)
			}
		}

		switch result.Status {
		case sdkbuilder.BuildSuccess:
			log.Printf("%s success building and pushing image: %s", functionName, result.Image)
		case sdkbuilder.BuildFailed:
			return fmt.Errorf("%s failure while building or pushing image %s: %s", functionName, imageName, result.Error)
		}
	}
	return nil
}
