package server

import (
	"github.com/sirupsen/logrus"
	"net/http"
)

const (
	authHeader = "Authorization"
)

func simpleAuth(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get(authHeader)
		if auth != secret {
			logrus.Errorf("request [%s %s] from %s is not authorized", r.Method, r.URL, r.Host)
			logrus.Debugf("headers: %+v", r.Header)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
