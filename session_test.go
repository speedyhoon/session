package session

import (
	"math"
	"math/bits"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCharset(t *testing.T) {
	var b byte
	var validChars string

	for b = 0; b < math.MaxUint8; b++ {
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

func TestGenerateID(t *testing.T) {
	const iterations = math.MaxUint16

	list := make(map[string]struct{}, iterations)
	for i := 0; i < iterations; i++ {
		id := generateID()
		assert.Len(t, id, idLength)

		_, ok := list[id]
		assert.False(t, ok, "duplicate session id generated", id)
		t.Log(id)

		list[id] = struct{}{}
	}
}

// TestLetterIdxBits tests the constant value is correct.
func TestLetterIdxBits(t *testing.T) {
	assert.Equal(t, letterIdxBits, bits.Len(uint(charsetSize)))
}

// TestMaxAge tests the constant value is correct.
func TestMaxAge(t *testing.T) {
	assert.Equal(t, maxAge, int(ExpiryTime.Seconds()))
}

// TestAllCharsUsed tests all runes within charset are utilised.
func TestAllCharsUsed(t *testing.T) {
	const passWithin = 49
	charsUsed := map[rune]struct{}{}

	for i := 0; i < passWithin; i++ {
		s := generateID()
		for _, c := range s {
			charsUsed[c] = struct{}{}
		}
		if int64(len(charsUsed)) == charsetSize {
			return
		}
	}
	assert.Len(t, charsUsed, int(charsetSize))
}
