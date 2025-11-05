// File: utils/keygen.go

// This file contains utility functions for generating unique activation codes and API keys.

package utils

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
)

// GenerateActivationCode generates TWTH-XXXX-YYYY format
func GenerateActivationCode() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	code := "TWTH-" +
		strings.ToUpper(hex.EncodeToString(b[0:2])) + "-" +
		strings.ToUpper(hex.EncodeToString(b[2:4]))
	return code, nil
}

// GenerateApiKey generates a secure 16-byte random string base64-url encoded
func GenerateApiKey() (string, error) {
	b := make([]byte, 16) // 128 bits
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	apiKey := "twth_" + base64.RawURLEncoding.EncodeToString(b)
	return apiKey, nil
}

// GenerateUniqueApiKey ensures generated API key is unique
func GenerateUniqueApiKey(conn *sql.DB) (string, error) {
	for range 5 {
		apiKey, err := GenerateApiKey()
		if err != nil {
			return "", err
		}
		var exists bool
		err = conn.QueryRow(`SELECT EXISTS(SELECT 1 FROM accounts WHERE api_key = $1)`, apiKey).Scan(&exists)
		if err != nil {
			return "", err
		}
		if !exists {
			return apiKey, nil
		}
	}
	return "", errors.New("could not generate unique API key")
}
