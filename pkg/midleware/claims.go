package midleware

import (
	"context"

	"github.com/labstack/echo/v5"
)

type contextKey string

const claimsContextKey contextKey = "jwt_claims"

type Claims struct {
	UserID int32
	Login  string
}

func WithClaims(ctx context.Context, c Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, c)
}

func GetClaims(e *echo.Context) (Claims, bool) {
	v := e.Request().Context().Value(claimsContextKey)
	if v == nil {
		return Claims{}, false
	}
	cl, ok := v.(Claims)
	return cl, ok
}
