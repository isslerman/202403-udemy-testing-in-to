package main

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

func getSession() *scs.SessionManager {
	session := scs.New()
	// time to session last
	session.Lifetime = 24 * time.Hour
	// if the cookie will persist - for here it's ok.
	// in production, more problably using redis or other solution
	session.Cookie.Persist = true
	// compatibilty with old browsers
	session.Cookie.SameSite = http.SameSiteLaxMode
	// crypto cookies
	session.Cookie.Secure = true

	return session
}
