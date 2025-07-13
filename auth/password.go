package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"golang.org/x/crypto/argon2"
)

// Argon2Params defines the parameters for Argon2 hashing
type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultArgon2Params returns recommended parameters for Argon2
func DefaultArgon2Params() *Argon2Params {
	return &Argon2Params{
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// HashPassword creates an Argon2 hash of the password
func HashPassword(password string) (string, error) {
	return HashPasswordWithParams(password, DefaultArgon2Params())
}

// HashPasswordWithParams creates an Argon2 hash with custom parameters
func HashPasswordWithParams(password string, params *Argon2Params) (string, error) {
	// Generate a cryptographically secure random salt
	salt := make([]byte, params.SaltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate the hash using Argon2id (recommended variant)
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	// Encode the hash and salt
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)

	// Return the encoded hash in the format: $argon2id$v=19$m=65536,t=3,p=2$salt$hash
	encodedParams := fmt.Sprintf("m=%d,t=%d,p=%d", params.Memory, params.Iterations, params.Parallelism)
	return fmt.Sprintf("$argon2id$v=19$%s$%s$%s", encodedParams, encodedSalt, encodedHash), nil
}

// CheckPassword verifies a password against its hash
func CheckPassword(password, encodedHash string) (bool, error) {
	// Parse the encoded hash
	params, salt, hash, err := parseEncodedHash(encodedHash)
	if err != nil {
		return false, fmt.Errorf("failed to parse hash: %w", err)
	}

	// Generate hash with the same parameters
	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	// Compare hashes using constant-time comparison
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

// parseEncodedHash parses the encoded hash string to extract parameters, salt, and hash
func parseEncodedHash(encodedHash string) (*Argon2Params, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, fmt.Errorf("invalid hash format")
	}

	// Check algorithm
	if parts[1] != "argon2id" {
		return nil, nil, nil, fmt.Errorf("unsupported algorithm: %s", parts[1])
	}

	// Check version
	if parts[2] != "v=19" {
		return nil, nil, nil, fmt.Errorf("unsupported version: %s", parts[2])
	}

	// Parse parameters
	paramParts := strings.Split(parts[3], ",")
	if len(paramParts) != 3 {
		return nil, nil, nil, fmt.Errorf("invalid parameters format")
	}

	params := &Argon2Params{}
	for _, param := range paramParts {
		keyValue := strings.Split(param, "=")
		if len(keyValue) != 2 {
			return nil, nil, nil, fmt.Errorf("invalid parameter format: %s", param)
		}

		switch keyValue[0] {
		case "m":
			if _, err := fmt.Sscanf(keyValue[1], "%d", &params.Memory); err != nil {
				return nil, nil, nil, fmt.Errorf("invalid memory parameter: %w", err)
			}
		case "t":
			if _, err := fmt.Sscanf(keyValue[1], "%d", &params.Iterations); err != nil {
				return nil, nil, nil, fmt.Errorf("invalid iterations parameter: %w", err)
			}
		case "p":
			if _, err := fmt.Sscanf(keyValue[1], "%d", &params.Parallelism); err != nil {
				return nil, nil, nil, fmt.Errorf("invalid parallelism parameter: %w", err)
			}
		default:
			return nil, nil, nil, fmt.Errorf("unknown parameter: %s", keyValue[0])
		}
	}

	// Decode salt and hash
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode hash: %w", err)
	}

	params.SaltLength = uint32(len(salt))
	params.KeyLength = uint32(len(hash))

	return params, salt, hash, nil
}

// IsHashFormat checks if a string is in the expected Argon2 hash format
func IsHashFormat(hash string) bool {
	return strings.HasPrefix(hash, "$argon2id$")
}

// MigratePlainTextPassword migrates a plain text password to hashed format
func MigratePlainTextPassword(plainPassword string) (string, error) {
	return HashPassword(plainPassword)
} 