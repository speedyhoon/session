package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCharset(t *testing.T) {
	var b byte
	var validChars string

	for b = 0; b < 255; b++ {
		if validCookieValueByte(b) {
			validChars += string(b)
		}
	}

	assert.Equal(t, validChars, charset)
}

// From net/http/cookie.go
func validCookieValueByte(b byte) bool {
	return 0x20 <= b && b < 0x7f && b != '"' && b != ';' && b != '\\'
}
