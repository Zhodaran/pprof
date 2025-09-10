package http

import (
	"errors"
	"net/http"
	"strings"

	"studentgit.kata.academy/Zhodaran/go-kata/core/entity"
)

func TokenAuthMiddleware(resp entity.Responder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				resp.ErrorUnauthorized(w, errors.New("missing authorization token"))
				return
			}

			token = strings.TrimPrefix(token, "Bearer ")

			_, err := entity.TokenAuth.Decode(token)
			if err != nil {
				resp.ErrorUnauthorized(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
