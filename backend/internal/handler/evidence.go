package handler

import (
	"crypto/rand"
	"encoding/hex"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var evidenceObjectNameSanitizer = strings.NewReplacer(
	"/", "_",
	"\\", "_",
	" ", "_",
)

func buildUniqueEvidenceObjectName(prefix, label, originalFilename string) string {
	safeLabel := sanitizeEvidenceObjectSegment(label, "unknown")
	safeFilename := sanitizeEvidenceObjectSegment(filepath.Base(strings.TrimSpace(originalFilename)), "evidence")
	return prefix + "_" + safeLabel + "_" + uniqueEvidenceObjectSuffix() + "_" + safeFilename
}

func sanitizeEvidenceObjectSegment(value, fallback string) string {
	value = evidenceObjectNameSanitizer.Replace(strings.TrimSpace(value))
	value = strings.Trim(value, "._")
	if value == "" {
		return fallback
	}
	return value
}

func uniqueEvidenceObjectSuffix() string {
	var token [8]byte
	if _, err := rand.Read(token[:]); err != nil {
		return strconv.FormatInt(time.Now().UTC().UnixNano(), 36)
	}
	return strconv.FormatInt(time.Now().UTC().UnixNano(), 36) + "_" + hex.EncodeToString(token[:])
}
