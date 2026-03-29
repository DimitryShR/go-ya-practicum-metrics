package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type Signer struct {
	key []byte
}

func NewSigner(key string) *Signer {
	return &Signer{
		key: []byte(key),
	}
}

// Подпись для данных src, возвращаемая в виде строки в шестнадцатеричном формате
func (s *Signer) Sign(src []byte) string {
	h := hmac.New(sha256.New, s.key)
	h.Write(src)
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

// Проверка подписи sign для данных src.
// Ожидается sign в шестнадцатеричном формата строки,
// который будет декодирован в байты для сравнения с вычисленной подписью.
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
