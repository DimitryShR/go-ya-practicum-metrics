package rsacrypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
)

const (
	// rsaBlockSize — размер блока открытого текста для RSA-4096 с OAEP-SHA256.
	// Для RSA-4096 размер ключа 512 байт, OAEP-SHA256 добавляет 66 байт служебной информации,
	// поэтому максимальный размер данных на блок: 512 - 2*32 - 2 = 446.
	rsaBlockSize = 446

	// encryptedBlockSize — размер зашифрованного блока для RSA-4096.
	encryptedBlockSize = 512
)

// encryptBlock шифрует один блок данных с использованием RSA-OAEP-SHA256.
func encryptBlock(pub *rsa.PublicKey, block []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, block, nil)
}

// Encrypt разбивает plaintext на блоки по 446 байт, шифрует каждый блок
// с помощью RSA-OAEP-SHA256 и возвращает конкатенацию 512-байтовых зашифрованных блоков.
func Encrypt(pub *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	if pub == nil {
		return nil, fmt.Errorf("public key is nil")
	}

	if len(plaintext) == 0 {
		return nil, nil
	}

	// Вычисляем количество блоков
	numBlocks := len(plaintext) / rsaBlockSize
	if len(plaintext)%rsaBlockSize != 0 {
		numBlocks++
	}

	ciphertext := make([]byte, 0, numBlocks*encryptedBlockSize)

	for i := 0; i < numBlocks; i++ {
		start := i * rsaBlockSize
		end := start + rsaBlockSize
		if end > len(plaintext) {
			end = len(plaintext)
		}

		encrypted, err := encryptBlock(pub, plaintext[start:end])
		if err != nil {
			return nil, fmt.Errorf("encrypt block %d: %w", i, err)
		}

		ciphertext = append(ciphertext, encrypted...)
	}

	return ciphertext, nil
}
