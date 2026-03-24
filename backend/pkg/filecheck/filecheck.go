package filecheck

import (
	"io"
)

// magicSignatures maps MIME types to their file header magic bytes.
var magicSignatures = map[string][]byte{
	"image/jpeg": {0xFF, 0xD8, 0xFF},
	"image/png":  {0x89, 0x50, 0x4E, 0x47},
	"image/gif":  {0x47, 0x49, 0x46},
	"image/webp": {0x52, 0x49, 0x46, 0x46}, // RIFF header; WebP has "WEBP" at offset 8
	"image/bmp":  {0x42, 0x4D},
}

// ValidateMagicBytes reads the first few bytes of a file and checks if they
// match the expected magic bytes for the claimed Content-Type.
// Returns true if the file header matches the claimed type.
func ValidateMagicBytes(reader io.Reader, claimedType string) bool {
	expected, ok := magicSignatures[claimedType]
	if !ok {
		return false
	}

	header := make([]byte, len(expected))
	n, err := io.ReadFull(reader, header)
	if err != nil || n < len(expected) {
		return false
	}

	for i, b := range expected {
		if header[i] != b {
			return false
		}
	}

	// Extra check for WebP: bytes 8-11 must be "WEBP"
	if claimedType == "image/webp" {
		extra := make([]byte, 8) // read bytes 4-11
		n, err := io.ReadFull(reader, extra)
		if err != nil || n < 8 {
			return false
		}
		if string(extra[4:8]) != "WEBP" {
			return false
		}
	}

	return true
}
