package session

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

func New() *scs.SessionManager {
	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Name = "MANIMATIC_S"
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = true
	return sessionManager
}
