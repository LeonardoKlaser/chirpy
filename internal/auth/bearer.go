package auth

import (
	"net/http"
	"strings"
	"errors"
)

func GetBearerToken(headers http.Header) (string, error){
	authHeader := headers.Get("Authorization")
	if len(strings.Split(authHeader, " ") ) < 2{
		return "", errors.New("invalid Authorization header format")
	}


	tokenString := strings.TrimSpace(strings.Split(authHeader, " ")[1])
	if tokenString == "" {
		return "", errors.New("bearer token is empty")
	}
	return tokenString, nil

}