package cmds

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type cookie struct {
	Login   string
	Expires time.Time
}

func (ac cookie) isExpired() bool {
	return ac.Expires.Before(time.Now())
}

// cookies key represents a session token
type cookies map[string]cookie

func (ac cookies) createCookie(w http.ResponseWriter, username string) {
	token := uuid.NewString()
	expiration := time.Now().Add(15 * time.Minute)

	ac[token] = cookie{
		Login:   username,
		Expires: expiration,
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   token,
		Expires: expiration,
	})

	log.Printf("Created cookie for user: %s (cookie expires on: %s)\n", username, expiration.Format(time.RFC3339))
}

func (ac cookies) deleteCookie(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("Deleting cookie for user: %s\n", ac[cookie.Value].Login)
	delete(ac, cookie.Value)

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Now(),
	})
}
