package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/speedyhoon/frm"
)

func TestPurge(t *testing.T) {
	ExpiryTime(3)
	w := httptest.NewRecorder()

	Set(w, frm.Form{Action: 0})
	assert.Len(t, cache.store, 1)

	req, err := http.NewRequest(http.MethodPost, "/", nil)
	assert.NoError(t, err)

	copyCookie(t, w, req)
	frm.GetFields = func(id uint8) []frm.Field {
		return []frm.Field{
			{Name: "n"},
			{Name: "C"},
		}
	}

	Get(w, req, 0)
	assert.Len(t, cache.store, 0)

	Set(w, frm.Form{Action: 5})
	assert.Len(t, cache.store, 1)

	time.Sleep(expiryTime - time.Second)
	assert.Len(t, cache.store, 1)

	time.Sleep(2 * time.Second)
	assert.Len(t, cache.store, 0)
}
