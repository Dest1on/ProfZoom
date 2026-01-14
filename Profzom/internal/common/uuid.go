package common

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type UUID string

func NewUUID() UUID {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return UUID(fmt.Sprintf("fallback-%d", time.Now().UnixNano()))
	}
	// версия 4
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return UUID(fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(b[0:4]), hex.EncodeToString(b[4:6]), hex.EncodeToString(b[6:8]), hex.EncodeToString(b[8:10]), hex.EncodeToString(b[10:16])))
}

func (u UUID) String() string {
	return string(u)
}

func ParseUUID(value string) (UUID, error) {
	if len(value) != 36 {
		return "", fmt.Errorf("invalid uuid")
	}
	for i, r := range value {
		switch i {
		case 8, 13, 18, 23:
			if r != '-' {
				return "", fmt.Errorf("invalid uuid")
			}
			continue
		}
		if !isHexDigit(r) {
			return "", fmt.Errorf("invalid uuid")
		}
	}
	return UUID(value), nil
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') ||
		(r >= 'a' && r <= 'f') ||
		(r >= 'A' && r <= 'F')
}
