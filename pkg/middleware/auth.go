package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"historymap-microservices/pkg/tools"
	"net/http"
)

var auth_json = "Authorization"
func Auth(unAuthchain http.HandlerFunc) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get(auth_json) == ""  {
				if unAuthchain != nil {
					unAuthchain(w,r)
				}
			} else {
				if validate(w, r) {
					h(w, r)
				}
			}
		}
	}
}

func validate(w http.ResponseWriter, r *http.Request) bool {
	jwtString := r.Header.Get(auth_json)
	token, err := jwt.Parse(jwtString, func (token *jwt.Token) (interface{}, error){
		return []byte(tools.JwtSecretKey), nil
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return false
	}

	if !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}

	return true
}

