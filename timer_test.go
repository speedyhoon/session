package session

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/speedyhoon/frm"
)

func TestTimer(t *testing.T) {
	ExpiryTime(3)
	w := httptest.NewRecorder()

	const mx = 10000

	for i := 1; i <= mx; i++ {
		Set(w, frm.Form{Action: uint8(i), Fields: []frm.Field{{Name: "n", Required: true}}})
		assert.Len(t, cache.store, i)
	}

	time.Sleep(expiryTime - time.Second)
	assert.Len(t, cache.store, mx)

	time.Sleep(2 * time.Second)
	assert.Len(t, cache.store, 0)
}
