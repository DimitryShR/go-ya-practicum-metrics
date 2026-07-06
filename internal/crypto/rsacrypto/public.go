// Package rsacrypto предоставляет функции для загрузки RSA-ключей,
// а также для блочного шифрования и расшифровки данных с использованием RSA-OAEP-SHA256.
package rsacrypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// LoadPublicKey загружает публичный RSA-ключ из X.509 сертификата в PEM-формате.
// Файл должен содержать PEM-блок "CERTIFICATE".
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	pemData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read public key file: %w", err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in file: %s", path)
	}

	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("unexpected PEM block type: %s, expected CERTIFICATE", block.Type)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse certificate: %w", err)
	}

	pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("certificate public key is not RSA")
	}

	return pubKey, nil
}
