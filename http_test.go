package session

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/speedyhoon/frm"
)

// TestSetupMissing tests a panic will be caused by frm.GetFields not being set before session.Get is called.
func TestSetupMissing(t *testing.T) {
	const frmOne = 1

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	frm.GetFields = nil // Resets if other tests have assigned a function.

	assert.Panics(t, func() {
		_, _ = Get(w, r, frmOne)
	})
}

// TestGetSetupIncomplete tests session.Get returns a form with no fields.
func TestSetupIncomplete(t *testing.T) {
	const frmOne = 1

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	frm.GetFields = func(uint8) []frm.Field {
		return []frm.Field{}
	}

	form, action := Get(w, r, frmOne)
	assert.Equal(t, uint8(255), action)

	expected := frm.Form{Action: frmOne, Fields: []frm.Field{}}
	assert.Equal(t, expected, form[frmOne])
}

func TestGetOneForm(t *testing.T) {
	const (
		frmOne = 1
	)
	frm.GetFields = setupOneForm
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	form, action := Get(w, r, frmOne)
	assert.Equal(t, uint8(255), action)
	assert.Len(t, form, 1)

	expected := frm.Form{Action: frmOne, Fields: []frm.Field{{Name: "one"}}}
	assert.Equal(t, expected, form[frmOne])
}

func TestGetFourForms(t *testing.T) {
	const (
		token  = "s"
		frmOne = 1
	)
	frm.GetFields = func(formID uint8) []frm.Field {
		switch formID {
		case 1:
			return []frm.Field{{Name: "one"}}
		case 2:
			return []frm.Field{{Name: "two"}}
		case 3:
			return []frm.Field{{Name: "three"}}
		case 4:
			return []frm.Field{{Name: "four"}}
		}
		return []frm.Field{}
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	Set(w, frm.Form{Action: frmOne, Fields: []frm.Field{{Name: "dog", Focus: true}}})

	r = httptest.NewRequest(http.MethodPost, "/", nil)
	copyCookie(t, w, r)

	form, action := Get(w, r, frmOne, 2, 3, 4)
	assert.Equal(t, uint8(1), action)
	assert.Len(t, form, 4)

	expected := map[uint8]frm.Form{
		frmOne: {Action: frmOne, Fields: []frm.Field{{Name: "dog", Focus: true}}},
		2:      {Action: 2, Fields: []frm.Field{{Name: "two"}}},
		3:      {Action: 3, Fields: []frm.Field{{Name: "three"}}},
		4:      {Action: 4, Fields: []frm.Field{{Name: "four"}}},
	}
	assert.Equal(t, expected, form)
}

func TestGetAndSet(t *testing.T) {
	expected := frm.Field{Name: "two", Required: true, Focus: true, Disable: true}
	const (
		token  = "s"
		frmOne = 1
		frmTwo = 2
	)
	frm.GetFields = setupOneForm
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	Set(w, frm.Form{Action: frmTwo, Fields: []frm.Field{expected}})

	r = httptest.NewRequest(http.MethodPost, "/", nil)
	copyCookie(t, w, r)
	form, action := Get(w, r, frmOne, frmTwo)

	assert.Equal(t, uint8(2), action)
	assert.Len(t, form, 2)
	assert.Equal(t, map[uint8]frm.Form{
		frmOne: {Action: frmOne, Fields: []frm.Field{{Name: "one"}}},
		frmTwo: {Action: frmTwo, Fields: []frm.Field{expected}},
	}, form)
}

func TestGetPurged(t *testing.T) {
	const (
		token  = "s"
		frmOne = 1
		frmTwo = 2
	)
	frm.GetFields = setupOneForm
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.AddCookie(&http.Cookie{Name: token, Value: "!#$%&'()*+,-./0123456789"})

	form, action := Get(w, r, frmOne, frmTwo)
	assert.Equal(t, uint8(255), action)
	assert.Len(t, form, 2)
	assert.Equal(t, map[uint8]frm.Form{
		frmOne: {Action: frmOne, Fields: []frm.Field{{Name: "one"}}},
		frmTwo: {Action: frmTwo, Fields: []frm.Field{}},
	}, form)
}

func TestGetMissingForm(t *testing.T) {
	const (
		token  = "s"
		frmTwo = 2
	)
	frm.GetFields = setupOneForm

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	form, action := Get(w, r, frmTwo)
	assert.Equal(t, uint8(255), action)
	assert.Len(t, form, 1)

	cookie, err := r.Cookie(token)
	assert.Equal(t, http.ErrNoCookie, err)
	assert.Nil(t, cookie)

	assert.Equal(t, frm.Form{Action: frmTwo, Fields: []frm.Field{}}, form[frmTwo])
}

func setupOneForm(formID uint8) []frm.Field {
	switch formID {
	case 1:
		return []frm.Field{
			{Name: "one"},
		}
	}
	return []frm.Field{}
}

func copyCookie(t *testing.T, w *httptest.ResponseRecorder, r *http.Request) {
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == token {
			r.AddCookie(cookie)
			return
		}
	}
	t.Errorf("cookie `%s` not found in response", token)
}
