package auth

import (
	"fmt"
	"net/http"
)

var ApiAuthKey string

func Require(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", ApiAuthKey) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
