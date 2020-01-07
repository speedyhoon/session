package session

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/speedyhoon/frm"
)

const (
	token      = "s"
	expiryTime = time.Minute * 2
)

type session struct {
	Expiry time.Time
	Form   frm.Form
}

var sessionCache = struct {
	sync.RWMutex
	m map[string]session
}{m: make(map[string]session)}

func init() {
	//periodically delete expired sessions
	go func() {
		for range time.NewTicker(expiryTime).C {
			//Can't directly change global variables in a go routine, so call an external function.
			purge()
		}
	}()
}

//Set attaches a newly generated session ID to the HTTP headers & saves the form for future retrieval.
func Set(w http.ResponseWriter, f frm.Form) {
	id := generateID()

	//Start mutex write lock.
	sessionCache.Lock()
	for {
		if _, ok := sessionCache.m[id]; !ok {
			//Assign the session ID if it isn't already assigned
			sessionCache.m[id] = session{Form: f, Expiry: time.Now().UTC().Add(expiryTime)}
			break
		}
		//Else sessionID is already assigned, so regenerate a different session ID
		id = generateID()
	}
	sessionCache.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     token,
		Value:    id,
		HttpOnly: true, //HttpOnly means the cookie can't be accessed by JavaScript
		MaxAge:   int(expiryTime.Seconds()),
	})
}

//Get retrieves a slice of forms and clears the Set-Cookie HTTP header.
func Get(w http.ResponseWriter, r *http.Request, getFields func(uint8) []frm.Field, formIDs ...uint8) (fs map[uint8]frm.Form, action uint8) {
	action--

	//Get users session id from request cookie header
	cookie, err := r.Cookie(token)
	if err != nil || cookie == nil || cookie.Value == "" {
		//No session found. Return default forms.
		return getForms(getFields, formIDs...), action
	}

	//Remove client session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     token,
		HttpOnly: true, //HttpOnly means the cookie can't be accessed by JavaScript
		MaxAge:   0,
	})

	//Start a read lock to prevent concurrent reads while other parts are executing a write.
	sessionCache.Lock()
	contents, ok := sessionCache.m[cookie.Value]
	if ok {
		//Clear the session contents as it has been returned to the user.
		delete(sessionCache.m, cookie.Value)
	}
	sessionCache.Unlock()
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
		//Get form fields because they are not populated for successful requests that passed validation
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

//generateID generates a new random session ID string 24 ASCII characters long
func generateID() string {
	const (
		//string generated from validCookieValueByte golang source code net/http/cookie.go
		letterBytes   = "!#$%&'()*+,-./0123456789:<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[]^_`abcdefghijklmnopqrstuvwxyz{|}~"
		n             = 24                   //Session ID length is recommended to be at least 16 characters long.
		letterIdxBits = 6                    //6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 //All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   //# of letter indices fitting in 63 bits
	)
	src := rand.NewSource(time.Now().UnixNano())
	//credit: icza, stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
	b := make([]byte, n)
	//A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

//purge deletes unused sessions when their expiry datetime lapses.
func purge() {
	now := time.Now()
	sessionCache.Lock()
	for sessionID := range sessionCache.m {
		if sessionCache.m[sessionID].Expiry.Before(now) {
			delete(sessionCache.m, sessionID)
		}
	}
	sessionCache.Unlock()
}
