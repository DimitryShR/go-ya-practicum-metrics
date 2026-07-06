package rsacrypto

import (
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
)

// Decrypt разбивает ciphertext на блоки по 512 байт, расшифровывает каждый блок
// с помощью RSA-OAEP-SHA256 и возвращает конкатенацию расшифрованных блоков.
func Decrypt(priv *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	if priv == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	if len(ciphertext) == 0 {
		return nil, nil
	}

	if len(ciphertext)%encryptedBlockSize != 0 {
		return nil, fmt.Errorf("ciphertext length (%d) is not a multiple of block size (%d)",
			len(ciphertext), encryptedBlockSize)
	}

	numBlocks := len(ciphertext) / encryptedBlockSize
	plaintext := make([]byte, 0, numBlocks*rsaBlockSize)

	for i := 0; i < numBlocks; i++ {
		start := i * encryptedBlockSize
		end := start + encryptedBlockSize

		decrypted, err := rsa.DecryptOAEP(sha256.New(), nil, priv, ciphertext[start:end], nil)
		if err != nil {
			return nil, fmt.Errorf("decrypt block %d: %w", i, err)
		}

		plaintext = append(plaintext, decrypted...)
	}

	return plaintext, nil
}
