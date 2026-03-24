package builder

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/openfaas/go-sdk/seal"
)

// BuildSecretsFileName is the conventional filename for sealed build secrets
// within a build tar, placed alongside com.openfaas.docker.config.
const BuildSecretsFileName = "com.openfaas.secrets"

type sealConfig struct {
	PublicKey []byte
}

// WithBuildSecretsKey configures the public key used to seal per-build secrets
// into the build tar. The key must be a valid base64-encoded 32-byte Curve25519 public key.
func WithBuildSecretsKey(publicKey []byte) BuilderOption {
	return func(b *FunctionBuilder) {
		raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(publicKey)))
		if err != nil || len(raw) != 32 {
			b.buildSecretsErr = fmt.Errorf("invalid build secrets public key: expected base64-encoded 32-byte Curve25519 key")
			return
		}

		b.sealConfig = sealConfig{
			PublicKey: publicKey,
		}
	}
}

func sealBuildSecrets(cfg sealConfig, secrets map[string]string) ([]byte, error) {
	values := make(map[string][]byte, len(secrets))
	for k, v := range secrets {
		values[k] = []byte(v)
	}

	return seal.Seal(cfg.PublicKey, values)
}
