package utils

import (
	"crypto/rand"
	"encoding/hex"
	uuid "github.com/satori/go.uuid"
	"io"
	"regexp"
)

func IsEmailValid(email string) bool {
	if len(email) < 3 && len(email) > 254 {
		return false
	}
	match, err := regexp.MatchString("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$", email)
	return match && err == nil
}

// NewId generates unique Id
// use this function for Id generation
func NewId() string {
	return uuid.NewV4().String()
}

// UUID generates UUID
func UUID(size int) string {
	u := make([]byte, size)
	_, _ = io.ReadFull(rand.Reader, u)
	return hex.EncodeToString(u)
}

// Nil returns nil UUID
func Nil() string {
	return uuid.Nil.String()
}
