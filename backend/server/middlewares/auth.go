package middlewares

import "net/http"

func Auth(next http.Handler) http.Handler {
	return func(w http.ResponseWriter, r *http.Request) {

		next
	}
}
