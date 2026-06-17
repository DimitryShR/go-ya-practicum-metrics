// Package sign предоставляет утилиты для HMAC-SHA256 подписи и верификации данных.
package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Signer предоставляет методы для подписи и проверки подписи данных с использованием HMAC-SHA256.
type Signer struct {
	key []byte
}

// NewSigner возвращает новый экземпляр *Signer с заданным ключом.
// Если ключ пустой, подпись/верификация не выполняется.
func NewSigner(key string) *Signer {
	return &Signer{
		key: []byte(key),
	}
}

// Sign подписывает данные src с использованием HMAC-SHA256.
// Возвращает подпись в виде шестнадцатеричной строки.
func (s *Signer) Sign(src []byte) string {
	h := hmac.New(sha256.New, s.key)
	h.Write(src)
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

// Verify проверяет подпись sign для данных src.
// Ожидает sign в виде шестнадцатеричной строки, декодирует её и сравнивает с вычисленной подписью.
func (s *Signer) Verify(sign string, src []byte) (bool, error) {
	bSign, err := hex.DecodeString(sign)
	if err != nil {
		return false, err
	}

	h := hmac.New(sha256.New, s.key)
	h.Write(src)
	srcSign := h.Sum(nil)

	return hmac.Equal(bSign, srcSign), nil
}
