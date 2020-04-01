package constants

import (
	"strings"
)

// EncryptionCiphers supported encryption chipers
var EncryptionCiphers = []string{
	"aes",
	"rsa",
	"pgp",
}

// ChiperToInt cipter to int
func ChiperToInt(c string) int32 {
	c = strings.ToLower(c)
	for i, ec := range EncryptionCiphers {
		if c == strings.ToLower(ec) {
			return int32(i) + 1
		}
	}

	return 0
}

// EncryptionIValid return true if encryption i is valid
func EncryptionIValid(i int32) bool {
	if i-1 < 0 || i-1 >= int32(len(EncryptionCiphers)) {
		return false
	}

	return true
}

// ChiperToString cipter to int
func ChiperToString(i int32) string {
	if !EncryptionIValid(i) {
		return ""
	}

	return EncryptionCiphers[i-1]
}

// IsValidCipher return true if given cipher is valid
func IsValidCipher(c string) bool {
	c = strings.ToLower(c)
	for _, ec := range EncryptionCiphers {
		if strings.ToLower(ec) == c {
			return true
		}
	}

	return false
}
