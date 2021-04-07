# session
[![Build Status](https://travis-ci.org/speedyhoon/session.svg?branch=master)](https://travis-ci.org/speedyhoon/session)
[![Go Report Card](https://goreportcard.com/badge/github.com/speedyhoon/session)](https://goreportcard.com/report/github.com/speedyhoon/session)

Go session handling for temporarily storing form data submitted via HTTP Get or Post methods. An easy way to add simple server side form validation.
This implementation is designed to destroy the session contents after 2 minutes and isn't suitable for login authorisation or maintaining sessions across servers or after the process is terminated.

```go
package main

import (
    "net/http"

    "github.com/speedyhoon/frm"
    "github.com/speedyhoon/session"
	"github.com/speedyhoon/vl"
)

const frmUpdate = 7

func updateFooBar(w http.ResponseWriter, r *http.Request) {
	// Retrieve form contents if available, else return default form values specified in getFields().
	f, _ := session.Get(w, r, getDefaultFields, frmUpdate)
	
	// Insert code here to validate input data, insert or update database etc.
	// For example: Here we are modifying the form data returned. 
	f[frmUpdate].Fields[0].Value = "foo bar"
	
	// Store the modified/updated form data. This will regenerate the session ID to prevent a CSRF attack. 
	session.Set(w, f[frmUpdate])

	//Return response or redirect to different URL
}

func getDefaultFields(formID uint8) []frm.Field {
	switch formID {
	case frmUpdate:
		return []frm.Field{
			{Name: "foo", V8: vl.Str, Required: true, Value: getDBDefaultValue()},
		}
	}
	return []frm.Field{}
}
```