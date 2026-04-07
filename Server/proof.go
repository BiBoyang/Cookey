package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidRequestSecret = errors.New("invalid request secret")
	ErrInvalidSessionProof  = errors.New("session proof verification failed")
)

func ComputeEnvelopeProof(rid string, envelope EncryptedSession, requestSecret string) (string, error) {
	secretBytes, err := decodeRequestSecret(requestSecret)
	if err != nil {
		return "", err
	}

	message := strings.Join([]string{
		"cookey-session-v1",
		rid,
		envelope.Algorithm,
		envelope.EphemeralPublicKey,
		envelope.Nonce,
		envelope.Ciphertext,
		formatProofTime(envelope.CapturedAt.Time),
		strconv.Itoa(envelope.Version),
	}, "\n")

	return computeProof(secretBytes, message), nil
}

func VerifyEnvelopeProof(rid string, envelope EncryptedSession, requestSecret string) error {
	actual, err := ComputeEnvelopeProof(rid, envelope, requestSecret)
	if err != nil {
		return err
	}
	if !hmac.Equal([]byte(actual), []byte(envelope.RequestSignature)) {
		return ErrInvalidSessionProof
	}
	return nil
}

func decodeRequestSecret(requestSecret string) ([]byte, error) {
	secretBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(requestSecret))
	if err != nil || len(secretBytes) < 16 {
		return nil, ErrInvalidRequestSecret
	}
	return secretBytes, nil
}

func computeProof(secret []byte, message string) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func formatProofTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}
