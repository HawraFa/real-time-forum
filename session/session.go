package session

import (
	"net/http"
	//"log"
	"github.com/gorilla/sessions"
)

// Single global session store
var Store *sessions.CookieStore

func InitSessionStore(secretKey string) {
	Store = sessions.NewCookieStore([]byte(secretKey))
	Store.Options = &sessions.Options{
		Path:     "/",
		//MaxAge:   3600 * 8, // 8 hours
		MaxAge: 0,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}
}

func GetUserIDFromSession(r *http.Request) (int, error) {
	if Store == nil {
		return 0, http.ErrNoCookie
	}

	session, err := Store.Get(r, "forum-session")
	if err != nil {
		return 0, err
	}

	userID, ok := session.Values["user_id"].(int)
	if !ok {
		return 0, http.ErrNoCookie
	}

	return userID, nil
}