package util

import (
	"crypto/sha1"
	"encoding/hex"
)

func GetIDFromString(str *string) string {
	hasher := sha1.New()
	hasher.Write([]byte(*str))

	return hex.EncodeToString(hasher.Sum(nil))
}
