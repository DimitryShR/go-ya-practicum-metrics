package rsacrypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"testing"
	"time"
)

// generateTestKeyPair создаёт тестовую RSA-4096 пару ключей и возвращает
// пути к временным файлам: cert.pem и private.pem.
func generateTestKeyPair(t *testing.T) (certPath, keyPath string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	// Создаём самоподписанный сертификат
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"Test"},
			Country:      []string{"RU"},
		},
		IPAddresses: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP("::1"),
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	// Сохраняем сертификат
	certFile, err := os.CreateTemp(t.TempDir(), "cert*.pem")
	if err != nil {
		t.Fatalf("failed to create temp cert file: %v", err)
	}
	defer certFile.Close()

	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err != nil {
		t.Fatalf("failed to encode cert: %v", err)
	}

	// Сохраняем приватный ключ (PKCS1)
	keyFile, err := os.CreateTemp(t.TempDir(), "private*.pem")
	if err != nil {
		t.Fatalf("failed to create temp key file: %v", err)
	}
	defer keyFile.Close()

	err = pem.Encode(keyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		t.Fatalf("failed to encode private key: %v", err)
	}

	return certFile.Name(), keyFile.Name()
}

func TestLoadPublicKey(t *testing.T) {
	certPath, _ := generateTestKeyPair(t)

	pubKey, err := LoadPublicKey(certPath)
	if err != nil {
		t.Fatalf("LoadPublicKey failed: %v", err)
	}
	if pubKey == nil {
		t.Fatal("LoadPublicKey returned nil key")
	}
}

func TestLoadPublicKey_InvalidFile(t *testing.T) {
	_, err := LoadPublicKey("/nonexistent/file.pem")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLoadPublicKey_InvalidPEM(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "invalid*.pem")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer tmpFile.Close()
	tmpFile.WriteString("not a pem file")
	tmpFile.Close()

	_, err = LoadPublicKey(tmpFile.Name())
	if err == nil {
		t.Fatal("expected error for invalid PEM")
	}
}

func TestLoadPrivateKey(t *testing.T) {
	_, keyPath := generateTestKeyPair(t)

	privKey, err := LoadPrivateKey(keyPath)
	if err != nil {
		t.Fatalf("LoadPrivateKey failed: %v", err)
	}
	if privKey == nil {
		t.Fatal("LoadPrivateKey returned nil key")
	}
}

func TestLoadPrivateKey_InvalidFile(t *testing.T) {
	_, err := LoadPrivateKey("/nonexistent/file.pem")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	certPath, keyPath := generateTestKeyPair(t)

	pubKey, err := LoadPublicKey(certPath)
	if err != nil {
		t.Fatalf("LoadPublicKey failed: %v", err)
	}

	privKey, err := LoadPrivateKey(keyPath)
	if err != nil {
		t.Fatalf("LoadPrivateKey failed: %v", err)
	}

	tests := []struct {
		name    string
		payload []byte
	}{
		{"empty data", []byte{}},
		{"single block", []byte("hello world")},
		{"exact block size", bytes.Repeat([]byte("A"), rsaBlockSize)},
		{"multiple blocks", bytes.Repeat([]byte("B"), rsaBlockSize*3+100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := Encrypt(pubKey, tt.payload)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			decrypted, err := Decrypt(privKey, encrypted)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if !bytes.Equal(decrypted, tt.payload) {
				t.Fatalf("roundtrip mismatch: got %d bytes, want %d bytes", len(decrypted), len(tt.payload))
			}
		})
	}
}

func TestDecrypt_InvalidData(t *testing.T) {
	_, keyPath := generateTestKeyPair(t)

	privKey, err := LoadPrivateKey(keyPath)
	if err != nil {
		t.Fatalf("LoadPrivateKey failed: %v", err)
	}

	// Неверный размер ciphertext
	_, err = Decrypt(privKey, []byte{1, 2, 3})
	if err == nil {
		t.Fatal("expected error for invalid ciphertext size")
	}

	// Корректный размер, но мусорные данные
	_, err = Decrypt(privKey, make([]byte, encryptedBlockSize))
	if err == nil {
		t.Fatal("expected error for garbage ciphertext")
	}
}

func TestEncrypt_NilKey(t *testing.T) {
	_, err := Encrypt(nil, []byte("test"))
	if err == nil {
		t.Fatal("expected error for nil public key")
	}
}

func TestDecrypt_NilKey(t *testing.T) {
	_, err := Decrypt(nil, []byte("test"))
	if err == nil {
		t.Fatal("expected error for nil private key")
	}
}

func TestEncrypt_NilKeyEmptyData(t *testing.T) {
	_, err := Encrypt(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil public key")
	}
}
