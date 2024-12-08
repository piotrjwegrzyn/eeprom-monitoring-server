package api

import (
	"context"
	"net/http"

	strictnethttp "github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
)

const (
	GetSignInOperation     = "GetSignin"
	PostSignInOperation    = "PostSignin"
	GetStyleCSSOperation   = "GetStyleCss"
	GetFaviconIcoOperation = "GetFaviconIco"
)

func NewAuthMiddleware(cfg Config, cookies *map[string]Cookie) strictnethttp.StrictHTTPMiddlewareFunc {
	return func(f strictnethttp.StrictHTTPHandlerFunc, operationID string) strictnethttp.StrictHTTPHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
			if (operationID == GetSignInOperation || operationID == PostSignInOperation) &&
				isSignedIn(r, cookies) {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return nil, nil
			}
			if operationID == GetSignInOperation ||
				operationID == PostSignInOperation ||
				operationID == GetStyleCSSOperation ||
				operationID == GetFaviconIcoOperation ||
				isSignedIn(r, cookies) {
				return f(ctx, w, r, request)
			}

			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return nil, nil
		}
	}
}

func isSignedIn(r *http.Request, cookies *map[string]Cookie) bool {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return false
	}

	session, exists := (*cookies)[cookie.Value]
	if !exists || session.isExpired() {
		delete(*cookies, cookie.Value)
		return false
	}

	return true
}
