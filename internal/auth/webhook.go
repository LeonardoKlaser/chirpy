package auth

import (
	"net/http"
	"strings"
	"errors"
)

func GetAPIKey(headers *http.Header) (string, error){
	apiheader := headers.Get("Authorization")
	if len(strings.Split(apiheader, " ") ) < 2{
		return "", errors.New("invalid Authorization header format")
	}

	tokenString := strings.TrimSpace(strings.Split(apiheader, " ")[1])
	if tokenString == "" {
		return "", errors.New("token is empty")
	}
	return tokenString, nil
}
