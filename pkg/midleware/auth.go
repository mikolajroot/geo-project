package midleware

import (
	apperrors "geo-project/pkg/errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

func JWTAuthMiddleware(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return apperrors.NewAppError("UNAUTHORIZED", "missing authorization header")
			}

			const bearer = "Bearer "
			if len(auth) <= len(bearer) || auth[:len(bearer)] != bearer {
				return apperrors.NewAppError("UNAUTHORIZED", "invalid authorization header")
			}

			tokenStr := auth[len(bearer):]

			type claims struct {
				UserID int32  `json:"user_id"`
				Login  string `json:"login"`
				jwt.RegisteredClaims
			}

			var cl claims
			t, err := jwt.ParseWithClaims(tokenStr, &cl, func(t *jwt.Token) (any, error) {
				return []byte(jwtSecret), nil
			})
			if err != nil || !t.Valid {
				return apperrors.NewAppError("UNAUTHORIZED", "invalid token")
			}

			if cl.UserID <= 0 {
				return apperrors.NewAppError("UNAUTHORIZED", "invalid token claims")
			}
			if cl.Login == "" {
				return apperrors.NewAppError("UNAUTHORIZED", "invalid token claims")
			}

			newCtx := WithClaims(c.Request().Context(), Claims{UserID: cl.UserID, Login: cl.Login})
			req := c.Request()
			req = req.WithContext(newCtx)
			c.SetRequest(req)

			return next(c)
		}
	}
}
