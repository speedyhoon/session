package session

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/speedyhoon/frm"
)

const (
	ExpiryTime = time.Minute * 2
	maxAge     = 120 // ExpiryTime in seconds.
	PurgeEvery = time.Second * 15

	token = "s"
	// String generated from validCookieValueByte Go source code net/http/cookie.go
	charset       = " !#$%&'()*+,-./0123456789:<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[]^_`abcdefghijklmnopqrstuvwxyz{|}~"
	charsetSize   = int64(len(charset))
	idLength      = 24                   // Session ID length is recommended to be at least 16 characters long.
	letterIdxBits = 5                    // Bits needed to represent a letter index.
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, same byte length as letterIdxBits.
	letterIdxMax  = 63 / letterIdxBits   // No. of letter indices fitting in 63 bits
)

type session struct {
	Expiry time.Time
	Form   frm.Form
}

var cache = struct {
	sync.RWMutex
	store map[string]session
}{store: make(map[string]session)}

func init() {
	// Periodically delete expired sessions.
	go func() {
		for range time.NewTicker(PurgeEvery).C {
			// Can't directly change global variables in a go routine, so call an external function.
			purge()
		}
	}()
}

//Set attaches a newly generated session ID to the HTTP headers & saves the form for future retrieval.
func Set(w http.ResponseWriter, f frm.Form) {
	// Generate the first ID before the cache is locked to reduce lock time
	id := generateID()

	// Start mutex write lock.
	cache.Lock()
	for {
		if _, ok := cache.store[id]; !ok {
			// Assign the session ID if it isn't already assigned
			cache.store[id] = session{Form: f, Expiry: time.Now().UTC().Add(ExpiryTime)}
			break
		}
		// Else sessionID is already assigned, so regenerate a different session ID
		id = generateID()
	}
	cache.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     token,
		Value:    id,
		HttpOnly: true, //HttpOnly means the cookie can't be accessed by JavaScript
		MaxAge:   maxAge,
	})
}

// Get retrieves a slice of forms and clears the Set-Cookie HTTP header.
func Get(w http.ResponseWriter, r *http.Request, getFields func(uint8) []frm.Field, formIDs ...uint8) (fs map[uint8]frm.Form, action uint8) {
	// Set default value to 255
	action--

	//Get users session id from request cookie header
	cookie, err := r.Cookie(token)
	if err != nil || cookie == nil || cookie.Value == "" {
		//No session found. Return default forms.
		return getForms(getFields, formIDs...), action
	}

	// Remove client session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     token,
		HttpOnly: true, // HttpOnly means the cookie can't be accessed by JavaScript
		MaxAge:   -1,   // MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
	})

	// Start a lock to prevent concurrent reads while other parts are executing a write.
	cache.Lock()
	contents, ok := cache.store[cookie.Value]
	if ok {
		// Clear the session contents because it has been returned to the user.
		delete(cache.store, cookie.Value)
	}
	cache.Unlock()

	if !ok || len(formIDs) == 0 {
		return getForms(getFields, formIDs...), action
	}

	fs = make(map[uint8]frm.Form, len(formIDs))
	for _, id := range formIDs {
		if contents.Form.Action == id {
			action = id
			if len(contents.Form.Fields) > 0 {
				fs[id] = contents.Form
				continue
			}
		}
		// Get form fields because they are not populated for successful requests that passed validation
		fs[id] = frm.Form{Action: id, Fields: getFields(id)}
	}

	return
}

func getForms(getFields func(uint8) []frm.Field, formIDs ...uint8) (f map[uint8]frm.Form) {
	f = make(map[uint8]frm.Form, len(formIDs))
	for _, id := range formIDs {
		f[id] = frm.Form{Action: id, Fields: getFields(id)}
	}

	return
}

// generateID generates a new random session ID string 24 ASCII characters long
func generateID() string {
	b := make([]byte, idLength)
	src := rand.NewSource(time.Now().UnixNano())

	// credit: icza, stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
	// Int63() generates 63 random bits, enough for letterIdxMax characters.
	for i, cache, remain := idLength-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < charsetSize {
			b[i] = charset[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// purge deletes unused sessions when their expiry datetime lapses.
func purge() {
	now := time.Now()
	cache.Lock()
	for sessionID := range cache.store {
		if cache.store[sessionID].Expiry.Before(now) {
			delete(cache.store, sessionID)
		}
	}
	cache.Unlock()
}
