package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
)

func getSearchPepper() string {
	if s := os.Getenv("SEARCH_PEPPER"); s != "" {
		return s
	}
	return "lovecheck_default_pepper_2026"
}

// GetEvidencePepper returns a pepper derived from SEARCH_PEPPER for evidence token signing.
func GetEvidencePepper() string {
	return "evd:" + getSearchPepper()
}

// DeterministicHash computes HMAC-SHA256(data, pepper) and returns a hex string.
// Unlike bcrypt, the same input always produces the same output, enabling O(1)
// indexed lookups instead of O(n) full-table scans with per-row comparison.
func DeterministicHash(data string) string {
	mac := hmac.New(sha256.New, []byte(getSearchPepper()))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// MaskPhoneNumber masks a phone number for privacy display (e.g., 13812345678 -> 138****5678).
func MaskPhoneNumber(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// MaskName masks a real name for privacy display.
func MaskName(name string) string {
	runes := []rune(name)
	if len(runes) <= 1 {
		return name
	} else if len(runes) == 2 {
		return string(runes[0]) + "*"
	}
	return string(runes[0]) + "*" + string(runes[len(runes)-1])
}
