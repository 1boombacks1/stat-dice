package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/models"
	"github.com/1boombacks1/stat_dice/server/httpErr"
	"github.com/go-chi/render"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appctx.FromContext(r.Context())

		token, err := extractToken(r)
		if err != nil {
			render.Render(w, r, httpErr.HTTPUnauthorized(fmt.Errorf("failed to get auth token: %w", err)))
			return
		}

		user, err := models.GetUserByJWT(ctx, token)
		if err != nil {
			render.Render(w, r, httpErr.HTTPUnauthorized(fmt.Errorf("failed to get user: %w", err)))
			return
		}

		r = r.WithContext(user.WithContext(ctx))
		next.ServeHTTP(w, r)
	})
}

func extractToken(r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return "", errors.New("missing authorization token")
	}

	splitToken := strings.Split(token, "Bearer")
	if len(splitToken) != 2 {
		return "", errors.New("invalid authorization token format")
	}

	token = strings.TrimSpace(splitToken[1])
	if token == "" {
		return "", errors.New("invalid authorization token format")
	}

	return token, nil
}
