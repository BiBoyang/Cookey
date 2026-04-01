package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/nacl/box"

	"cookey/internal/models"
)

func TestGenerateKeypair(t *testing.T) {
	keypair, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if keypair.Version != 1 {
		t.Fatalf("Version = %d, want 1", keypair.Version)
	}
	if keypair.Algorithm != "ed25519" {
		t.Fatalf("Algorithm = %q, want ed25519", keypair.Algorithm)
	}

	publicKey, err := base64.StdEncoding.DecodeString(keypair.PublicKey)
	if err != nil || len(publicKey) != 32 {
		t.Fatalf("public key decode failed: %v len=%d", err, len(publicKey))
	}

	privateKey, err := base64.StdEncoding.DecodeString(keypair.PrivateKey)
	if err != nil || len(privateKey) != 32 {
		t.Fatalf("private key decode failed: %v len=%d", err, len(privateKey))
	}
}

func TestX25519Derivation(t *testing.T) {
	keypair, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	first, err := DeriveX25519PrivateKey(keypair)
	if err != nil {
		t.Fatalf("DeriveX25519PrivateKey() error = %v", err)
	}
	second, err := DeriveX25519PrivateKey(keypair)
	if err != nil {
		t.Fatalf("DeriveX25519PrivateKey() error = %v", err)
	}

	if first != second {
		t.Fatalf("derivation not deterministic")
	}

	publicKey, err := X25519PublicKeyBase64(keypair)
	if err != nil {
		t.Fatalf("X25519PublicKeyBase64() error = %v", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil || len(decoded) != 32 {
		t.Fatalf("public key decode failed: %v len=%d", err, len(decoded))
	}
}

func TestDecryptRoundTrip(t *testing.T) {
	keypair, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	recipientPublicKey, err := X25519PublicKeyBase64(keypair)
	if err != nil {
		t.Fatalf("X25519PublicKeyBase64() error = %v", err)
	}

	recipientPublicKeyBytes, err := base64.StdEncoding.DecodeString(recipientPublicKey)
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}

	var recipientPublicKey32 [32]byte
	copy(recipientPublicKey32[:], recipientPublicKeyBytes)

	ephemeralPublicKey, ephemeralPrivateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("box.GenerateKey() error = %v", err)
	}

	var nonce [24]byte
	copy(nonce[:], mustRandomBytes(t, 24))

	plaintext := []byte(`{"cookies":[{"name":"session","value":"abc123","domain":"example.com","path":"/","expires":-1,"httpOnly":true,"secure":true,"sameSite":"Lax"}],"origins":[{"origin":"https://example.com","localStorage":[{"name":"token","value":"secret"}]}]}`)
	ciphertext := box.Seal(nil, plaintext, &nonce, &recipientPublicKey32, ephemeralPrivateKey)

	envelope := models.EncryptedSessionEnvelope{
		Version:            1,
		Algorithm:          models.SessionEncryptionAlgorithmX25519XSalsa20Poly1305,
		EphemeralPublicKey: base64.StdEncoding.EncodeToString(ephemeralPublicKey[:]),
		Nonce:              base64.StdEncoding.EncodeToString(nonce[:]),
		Ciphertext:         base64.StdEncoding.EncodeToString(ciphertext),
		CapturedAt:         models.NewISO8601Time(time.Now()),
	}

	decrypted, err := DecryptSessionEnvelope(envelope, keypair)
	if err != nil {
		t.Fatalf("DecryptSessionEnvelope() error = %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("plaintext mismatch")
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	keypair, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	recipientPublicKey, err := X25519PublicKeyBase64(keypair)
	if err != nil {
		t.Fatalf("X25519PublicKeyBase64() error = %v", err)
	}

	plaintext := []byte(`{"cookies":[{"name":"seed","value":"xyz","domain":"example.com","path":"/","expires":-1,"httpOnly":false,"secure":true,"sameSite":"Lax"}],"origins":[{"origin":"https://example.com","localStorage":[{"name":"mode","value":"refresh"}]}]}`)
	envelope, err := EncryptSessionEnvelope(plaintext, recipientPublicKey)
	if err != nil {
		t.Fatalf("EncryptSessionEnvelope() error = %v", err)
	}

	decrypted, err := DecryptSessionEnvelope(envelope, keypair)
	if err != nil {
		t.Fatalf("DecryptSessionEnvelope() error = %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("plaintext mismatch")
	}
}

func TestDecryptGoldenFixture(t *testing.T) {
	fixturePath := filepath.Join("testdata", "swift_fixture.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", fixturePath, err)
	}

	var fixture struct {
		Keypair              models.KeypairFile              `json:"keypair"`
		ExpectedX25519Public string                          `json:"expected_x25519_public"`
		Envelope             models.EncryptedSessionEnvelope `json:"envelope"`
		Plaintext            models.SessionFile              `json:"plaintext"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	publicKey, err := X25519PublicKeyBase64(fixture.Keypair)
	if err != nil {
		t.Fatalf("X25519PublicKeyBase64() error = %v", err)
	}
	if publicKey != fixture.ExpectedX25519Public {
		t.Fatalf("X25519 public mismatch: got %q want %q", publicKey, fixture.ExpectedX25519Public)
	}

	decrypted, err := DecryptSessionEnvelope(fixture.Envelope, fixture.Keypair)
	if err != nil {
		t.Fatalf("DecryptSessionEnvelope() error = %v", err)
	}

	var session models.SessionFile
	if err := json.Unmarshal(decrypted, &session); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	expectedPlaintext, err := json.Marshal(fixture.Plaintext)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	actualPlaintext, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	if string(actualPlaintext) != string(expectedPlaintext) {
		t.Fatalf("plaintext mismatch")
	}
}

func TestRequestIDFormat(t *testing.T) {
	rid, err := GenerateRequestID()
	if err != nil {
		t.Fatalf("GenerateRequestID() error = %v", err)
	}

	if !regexp.MustCompile(`^r_[A-Za-z0-9_-]{22}$`).MatchString(rid) {
		t.Fatalf("unexpected request ID: %q", rid)
	}
}

func TestBase64URLNopadding(t *testing.T) {
	for i := 0; i < 16; i++ {
		rid, err := GenerateRequestID()
		if err != nil {
			t.Fatalf("GenerateRequestID() error = %v", err)
		}
		if strings.ContainsAny(rid, "=+/") {
			t.Fatalf("request ID should be raw base64url: %q", rid)
		}
	}
}

func mustRandomBytes(t *testing.T, length int) []byte {
	t.Helper()

	bytes, err := RandomBytes(length)
	if err != nil {
		t.Fatalf("RandomBytes() error = %v", err)
	}
	return bytes
}
