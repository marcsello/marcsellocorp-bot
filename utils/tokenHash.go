package utils

import "crypto/sha256"

func TokenHash(token string) []byte {
	h := sha256.New()
	h.Write([]byte(token))
	return h.Sum(nil)
}
