package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/models"
	httpErrors "github.com/1boombacks1/stat_dice/server/http_errors"
	"github.com/go-chi/render"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appctx.FromContext(r.Context())

		cookie, err := r.Cookie("token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				w.Header().Set("HX-Redirect", "/login")
				// http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
				return
			}
			render.Render(w, r, httpErrors.ErrUnauthorized(fmt.Errorf("failed to get cookie: %w", err)))
			return
		}

		token := cookie.Value
		user, err := models.GetUserByJWT(ctx, token)
		if err != nil {
			render.Render(w, r, httpErrors.ErrUnauthorized(fmt.Errorf("failet to get user by JWT: %w", err)))
			return
		}
		if user == nil {
			render.Render(w, r, httpErrors.ErrUnauthorized(errors.New("invalid authorization token: user with this token not found")))
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
