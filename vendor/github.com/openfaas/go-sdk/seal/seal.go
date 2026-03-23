// Package seal provides authenticated public-key encryption using NaCl box
// (Curve25519 + XSalsa20-Poly1305). It is format-agnostic: the plaintext is
// opaque bytes that can hold YAML secrets, a CA certificate, or any other
// sensitive data.
//
// Sealed envelopes are serialised as YAML so they can be stored on disk,
// committed to git, or transferred over the wire.
package seal

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/nacl/box"
	"gopkg.in/yaml.v3"
)

const (
	// Version is the current sealed envelope format version.
	Version = "v1"

	// Algorithm is the encryption algorithm identifier.
	Algorithm = "nacl/box"
)

// Envelope is the per-value sealed format. Key names are visible,
// values are each independently encrypted as base64(nonce || ciphertext).
type Envelope struct {
	Version   string            `yaml:"version"`
	Algorithm string            `yaml:"algorithm"`
	KeyID     string            `yaml:"key_id,omitempty"`
	PublicKey string            `yaml:"public_key"`
	Secrets   map[string]string `yaml:"secrets"`
}

// Seal encrypts each value independently using a shared keypair.
// Each sealed value is base64(24-byte nonce || ciphertext).
// Key names are stored in cleartext for auditability and git diffs.
func Seal(publicKey []byte, values map[string][]byte, keyID string) ([]byte, error) {
	recipient, err := decodeKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating keypair: %w", err)
	}

	secrets := make(map[string]string, len(values))
	for k, v := range values {
		var nonce [24]byte
		if _, err := rand.Read(nonce[:]); err != nil {
			return nil, fmt.Errorf("generating nonce for %q: %w", k, err)
		}

		ciphertext := box.Seal(nil, v, &nonce, recipient, priv)

		// nonce || ciphertext, single base64 value
		combined := make([]byte, 24+len(ciphertext))
		copy(combined[:24], nonce[:])
		copy(combined[24:], ciphertext)
		secrets[k] = base64.StdEncoding.EncodeToString(combined)
	}

	env := Envelope{
		Version:   Version,
		Algorithm: Algorithm,
		KeyID:     keyID,
		PublicKey: base64.StdEncoding.EncodeToString(pub[:]),
		Secrets:   secrets,
	}

	return yaml.Marshal(env)
}

// Unseal decrypts a YAML-encoded Envelope, returning all values.
func Unseal(privateKey []byte, envelope []byte) (map[string][]byte, error) {
	env, pub, priv, err := parseEnvelope(privateKey, envelope)
	if err != nil {
		return nil, err
	}

	values := make(map[string][]byte, len(env.Secrets))
	for k, encoded := range env.Secrets {
		plaintext, err := unsealValue(encoded, k, pub, priv)
		if err != nil {
			return nil, err
		}
		values[k] = plaintext
	}

	return values, nil
}

// UnsealKey decrypts a single key from a YAML-encoded Envelope.
func UnsealKey(privateKey []byte, envelope []byte, key string) ([]byte, error) {
	env, pub, priv, err := parseEnvelope(privateKey, envelope)
	if err != nil {
		return nil, err
	}

	encoded, ok := env.Secrets[key]
	if !ok {
		return nil, fmt.Errorf("key %q not found in sealed envelope", key)
	}

	return unsealValue(encoded, key, pub, priv)
}

// KeyID extracts the key_id from a YAML-encoded sealed envelope
// without decrypting it.
func KeyID(envelope []byte) (string, error) {
	var env struct {
		KeyID string `yaml:"key_id"`
	}
	if err := yaml.Unmarshal(envelope, &env); err != nil {
		return "", fmt.Errorf("invalid sealed envelope: %w", err)
	}
	return env.KeyID, nil
}

// GenerateKeyPair generates a new Curve25519 keypair.
// Both keys are returned as base64-encoded bytes, matching the format
// expected by Seal and Unseal.
func GenerateKeyPair() (publicKey, privateKey []byte, err error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	return []byte(base64.StdEncoding.EncodeToString(pub[:])),
		[]byte(base64.StdEncoding.EncodeToString(priv[:])),
		nil
}

// parseEnvelope validates and decodes a YAML envelope, returning the
// parsed envelope, the sender's public key, and the recipient's private key.
func parseEnvelope(privateKey []byte, envelope []byte) (*Envelope, *[32]byte, *[32]byte, error) {
	priv, err := decodeKey(privateKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid private key: %w", err)
	}

	var env Envelope
	if err := yaml.Unmarshal(envelope, &env); err != nil {
		return nil, nil, nil, fmt.Errorf("invalid sealed envelope: %w", err)
	}

	if env.Version != Version {
		return nil, nil, nil, fmt.Errorf("unsupported envelope version: %q, expected %q", env.Version, Version)
	}

	if env.Algorithm != Algorithm {
		return nil, nil, nil, fmt.Errorf("unsupported algorithm: %q, expected %q", env.Algorithm, Algorithm)
	}

	pub, err := decodeKey([]byte(env.PublicKey))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid envelope public key: %w", err)
	}

	return &env, pub, priv, nil
}

// unsealValue decodes and decrypts a single base64(nonce || ciphertext) value.
func unsealValue(encoded string, key string, pub *[32]byte, priv *[32]byte) ([]byte, error) {
	combined, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return nil, fmt.Errorf("invalid base64 for key %q: %w", key, err)
	}

	if len(combined) < 24 {
		return nil, fmt.Errorf("sealed value for key %q too short: expected at least 24 bytes, got %d", key, len(combined))
	}

	var nonce [24]byte
	copy(nonce[:], combined[:24])
	ciphertext := combined[24:]

	plaintext, ok := box.Open(nil, ciphertext, &nonce, pub, priv)
	if !ok {
		return nil, fmt.Errorf("decryption failed for key %q: invalid key or corrupted data", key)
	}

	return plaintext, nil
}

func decodeKey(key []byte) (*[32]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(key)))
	if err != nil {
		return nil, err
	}
	if len(raw) != 32 {
		return nil, fmt.Errorf("expected 32-byte key, got %d", len(raw))
	}
	out := new([32]byte)
	copy(out[:], raw)
	return out, nil
}
