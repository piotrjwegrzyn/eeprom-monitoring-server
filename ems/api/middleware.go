package api

import (
	"context"
	"log/slog"
	"net/http"

	oapi "pi-wegrzyn/ems/api/oapi/generated"

	strictnethttp "github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
)

const (
	GetSignInOperation     = "GetSignin"
	PostSignInOperation    = "PostSignin"
	GetStyleCSSOperation   = "GetStaticStyleCss"
	GetFaviconIcoOperation = "GetStaticFaviconIco"
)

type SessionChecker interface {
	IsSignedIn(r *http.Request) bool
}

func NewAuthMiddleware(cfg Config, cookies SessionChecker) strictnethttp.StrictHTTPMiddlewareFunc {
	return func(f strictnethttp.StrictHTTPHandlerFunc, operationID string) strictnethttp.StrictHTTPHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (response any, err error) {
			if (operationID == GetSignInOperation || operationID == PostSignInOperation) && cookies.IsSignedIn(r) {
				return oapi.PageRedirectResponse{
					Headers: oapi.PageRedirectResponseHeaders{
						Location: "/",
					},
				}, nil
			}
			if operationID == GetSignInOperation ||
				operationID == PostSignInOperation ||
				operationID == GetStyleCSSOperation ||
				operationID == GetFaviconIcoOperation ||
				cookies.IsSignedIn(r) {
				return f(ctx, w, r, request)
			}

			return oapi.PageRedirectResponse{
				Headers: oapi.PageRedirectResponseHeaders{
					Location: "/signin",
				},
			}, nil
		}
	}
}

func NewLoggerMiddleware() strictnethttp.StrictHTTPMiddlewareFunc {
	return func(f strictnethttp.StrictHTTPHandlerFunc, operationID string) strictnethttp.StrictHTTPHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (response any, err error) {
			if operationID != GetFaviconIcoOperation && operationID != GetStyleCSSOperation {
				slog.InfoContext(ctx, "request",
					slog.Any("operationID", operationID),
					slog.Any("protocol", r.Proto),
					slog.Any("url", r.URL),
				)
			}

			return f(ctx, w, r, request)
		}
	}
}
