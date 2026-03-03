package session

import (
	"github.com/go-openapi/testify/v2/assert"
	"github.com/speedyhoon/frm"
	"testing"
	"time"
)

func TestPurge(t *testing.T) {
	cache.store[generateID()] = session{
		Expiry: time.Now(),
		Form:   frm.Form{},
	}
	assert.Len(t, cache.store, 1)

	purge()
	assert.Len(t, cache.store, 0)

	cache.store[generateID()] = session{
		Expiry: time.Now().Add(time.Second),
		Form:   frm.Form{},
	}
	assert.Len(t, cache.store, 1)

	purge()
	assert.Len(t, cache.store, 1)

	time.Sleep(time.Second)
	purge()
	assert.Len(t, cache.store, 0)
}
