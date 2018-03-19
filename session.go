package session

import (
	"math/rand"
	"net/http"
	"sync"
	"time"
	"github.com/speedyhoon/forms"
)

const (
	token             = "s"
	expiryTime = time.Minute * 2
)

type session struct {
	Expiry time.Time
	Form forms.Form
}

var globalSessions = struct {
	sync.RWMutex
	m map[string]session
}{m: make(map[string]session)}

func init() {
	go Upkeep()
}

//Upkeep periodically deletes expired sessions
func Upkeep() {
	for range time.NewTicker(expiryTime).C {
		//Can't directly change global variables in a go routine, so call an external function.
		Purge()
	}
}

//Purge sessions where the expiry datetime has already lapsed.
func Purge() {
	globalSessions.RLock()
	qty := len(globalSessions.m)
	globalSessions.RUnlock()
	if qty == 0 {
		return
	}

	now := time.Now()
	globalSessions.Lock()
	for sessionID := range globalSessions.m {
		if globalSessions.m[sessionID].Expiry.Before(now) {
			delete(globalSessions.m, sessionID)
		}
	}
	globalSessions.Unlock()
}

func Set(w http.ResponseWriter, f forms.Form) {
	s := session{Form: f, Expiry: time.Now().Add(expiryTime)}
	var ok bool
	var id string

	//Start mutex write lock.
	globalSessions.Lock()
	for {
		id = generateID()
		_, ok = globalSessions.m[id]

		if !ok {
			//Assign the session ID if it isn't already assigned
			globalSessions.m[id] = s
			break
		}
		//else if sessionID is already assigned then regenerate a different session ID
	}
	globalSessions.Unlock()

	cookie := http.Cookie{
		Name:     token,
		Value:    id,
		HttpOnly: true,
		Expires:  s.Expiry,
	}
	http.SetCookie(w, &cookie)
}

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

//Forms retrieves a slice of forms, including any errors (if any)
func Forms(w http.ResponseWriter, r *http.Request, getForm func(uint8)forms.Form, formIDs ...uint8) (uint8, []forms.Form) {
	const noAction = 255

	//Get users session id from request cookie header
	cookie, err := r.Cookie(token)
	if err != nil || cookie == nil || cookie.Value == "" {
		//No session found. Return default forms.
		return noAction, fetch(getForm, formIDs...)
	}

	//Remove client session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     token,
		Value:    "",                                       //Remove cookie by setting it to nothing (empty string).
		HttpOnly: true,                                     //HttpOnly means the cookie can't be accessed by JavaScript
		Expires:  time.Now().UTC().Add(-expiryTime), //Using minus expiryTime so the session expires time is set to the past
	})

	//Start a read lock to prevent concurrent reads while other parts are executing a write.
	globalSessions.RLock()
	contents, ok := globalSessions.m[cookie.Value]
	globalSessions.RUnlock()
	if !ok {
		return noAction, fetch(getForm, formIDs...)
	}

	//Clear the session contents as it has been returned to the user.
	globalSessions.Lock()
	delete(globalSessions.m, cookie.Value)
	globalSessions.Unlock()

	var f []forms.Form
	for _, id := range formIDs {
		if contents.Form.Action == id {
			f = append(f, contents.Form)
		} else {
			f = append(f, getForm(id))
		}
	}
	return contents.Form.Action, f
}

func fetch(getForm func(uint8)forms.Form, formIDs ...uint8) (f []forms.Form) {
	for _, id := range formIDs {
		f = append(f, getForm(id))
	}
	return f
}
