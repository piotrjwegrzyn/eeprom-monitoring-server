package cookies

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Store struct {
	expiration time.Duration
	cookies    map[string]cookie
}

func NewStore(expiration time.Duration) *Store {
	return &Store{
		expiration: expiration,
		cookies:    make(map[string]cookie),
	}
}

func (cs *Store) Create(ctx context.Context, w http.ResponseWriter, username string) {
	cs.pruneExpired()

	token := uuid.NewString()
	expiration := time.Now().Add(cs.expiration)

	cs.cookies[token] = cookie{
		Login:   username,
		Expires: expiration,
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   token,
		Expires: expiration,
	})

	slog.InfoContext(ctx, "created cookie for user", slog.Any("username", username), slog.Any("expiresAt", expiration))
}

func (cs *Store) Delete(ctx context.Context, w http.ResponseWriter, token *string) {
	cs.pruneExpired()

	if token == nil {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   *token,
		Expires: time.Now(),
	})

	if session, exists := cs.cookies[*token]; exists {
		slog.InfoContext(ctx, "deleting cookie for user", slog.Any("username", session.Login))
		delete(cs.cookies, *token)
	}
}

func (cs *Store) IsSignedIn(r *http.Request) bool {
	cs.pruneExpired()

	cookie, err := r.Cookie("session_token")
	if err != nil {
		return false
	}

	session, exists := cs.cookies[cookie.Value]
	if !exists || session.isExpired() {
		delete(cs.cookies, cookie.Value)
		return false
	}

	return true
}

func (cs *Store) pruneExpired() {
	for k, v := range cs.cookies {
		if v.isExpired() {
			delete(cs.cookies, k)
		}
	}
}

type cookie struct {
	Login   string
	Expires time.Time
}

func (ac cookie) isExpired() bool {
	return ac.Expires.Before(time.Now())
}
