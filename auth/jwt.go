package auth

import (
	"net/http"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/dgrijalva/jwt-go"
)

var defaultSecret = []byte("this$is#my(secret)314")
var signingMethod = jwt.SigningMethodHS256

type authmiddleware func(http.HandlerFunc) http.HandlerFunc

func GenerateToken(username, secret string) (string, error) {

	token := jwt.New(signingMethod)

	claims := token.Claims.(jwt.MapClaims)

	claims["name"] = username
	claims["exp"] = time.Now().Add(time.Minute * 10).Unix()

	return token.SignedString(getSecret(secret))
}

func getSecret(secret string) []byte {
	if len(secret) > 0 {
		return []byte(secret)
	}

	return defaultSecret
}

func WithJWT(secret string, securityEnabled bool, next http.HandlerFunc) http.HandlerFunc {
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return getSecret(secret), nil
		},
		SigningMethod: signingMethod,
	})
	return func(w http.ResponseWriter, r *http.Request) {
		if securityEnabled {
			err := jwtMiddleware.CheckJWT(w, r)

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

		}
		next.ServeHTTP(w, r)
	}
}
