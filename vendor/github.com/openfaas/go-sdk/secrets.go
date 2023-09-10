package sdk

import (
	"fmt"
	"os"
	"path"
	"strings"
)

// ReadSecrets reads a single secrets from /var/openfaas/secrets or from
// the environment "secret_mount_path" if set.
func ReadSecret(key string) (string, error) {

	readPath := getPath(key)
	secretBytes, readErr := os.ReadFile(readPath)
	if readErr != nil {
		return "", fmt.Errorf("unable to read secret: %s, error: %s", readPath, readErr)
	}
	val := strings.TrimSpace(string(secretBytes))
	return val, nil
}

// ReadSecrets reads all secrets from /var/openfaas/secrets or from
// the environment "secret_mount_path" if set.
// The results are returned in a map of key/value pairs.
func ReadSecrets() (SecretMap, error) {

	values := map[string]string{}
	secretMap := newSecretMap(values)
	base := getPath("")

	files, err := os.ReadDir(base)
	if err != nil {
		return secretMap, err
	}

	for _, file := range files {
		val, err := ReadSecret(file.Name())
		if err != nil {
			return secretMap, err
		}
		values[file.Name()] = val
	}

	return secretMap, nil
}

func newSecretMap(values map[string]string) SecretMap {
	return SecretMap{
		values: values,
	}
}

func getPath(key string) string {
	basePath := "/var/openfaas/secrets/"
	if len(os.Getenv("secret_mount_path")) > 0 {
		basePath = os.Getenv("secret_mount_path")
	}
	return path.Join(basePath, key)
}

type SecretMap struct {
	values map[string]string
}

func (s *SecretMap) Get(key string) (string, error) {
	val, ok := s.values[key]
	if !ok {
		return "", fmt.Errorf("secret %s not found", key)
	}
	return val, nil
}

func (s *SecretMap) Exists(key string) bool {
	_, ok := s.values[key]
	return ok
}
