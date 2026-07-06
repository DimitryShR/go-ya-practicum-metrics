package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// Находим корень проекта
	projectRoot, err := findProjectRoot()
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	// Создание директории для сертификатов
	certsDir := filepath.Join(projectRoot, ".certs")
	if err := os.MkdirAll(certsDir, 0755); err != nil {
		log.Fatalf("Failed to create .certs directory: %v\n", err)
	}

	// Генерация RSA-ключа 4096 бит
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v\n", err)
	}

	// Создание X.509 сертификата
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"Yandex.Praktikum"},
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
		log.Fatalf("Failed to create certificate: %v\n", err)
	}

	// Сохранение сертификата (публичный ключ) в cert.pem
	certFile, err := os.Create(filepath.Join(certsDir, "cert.pem"))
	if err != nil {
		log.Fatalf("Failed to create cert.pem: %v\n", err)
	}
	defer certFile.Close()

	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err != nil {
		log.Fatalf("Failed to encode certificate: %v\n", err)
	}
	log.Println("Created cert.pem (X.509 certificate with public key)")

	// Сохранение приватного ключа (PKCS1) в private.pem
	keyFile, err := os.Create(filepath.Join(certsDir, "private.pem"))
	if err != nil {
		log.Fatalf("Failed to create private.pem: %v\n", err)
	}
	defer keyFile.Close()

	err = pem.Encode(keyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		log.Fatalf("Failed to encode private key: %v\n", err)
	}
	log.Println("Created private.pem (RSA private key, PKCS1)")
}

// findProjectRoot находит корень проекта.
func findProjectRoot() (string, error) {
	// Получаем текущую рабочую директорию
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Поднимаемся по дереву директорий, пока не найдем go.mod
	for {
		goModPath := filepath.Join(wd, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return wd, nil
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			// Достигли корня файловой системы
			break
		}
		wd = parent
	}

	return "", fmt.Errorf("go.mod not found in any parent directory")
}
