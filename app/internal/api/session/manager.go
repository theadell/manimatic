package session

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
)

func New() *scs.SessionManager {
	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Name = "MANIMATIC_SS"
	sessionManager.Cookie.Persist = true
	sessionManager.Store = memstore.New()
	sessionManager.Cookie.SameSite = http.SameSiteNoneMode
	sessionManager.Cookie.Secure = true
	return sessionManager
}
