package auth

import(
	"github.com/google/uuid"
	"fmt"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"errors"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error){
	mySigningKey := []byte(tokenSecret)
	idString := userID.String()
	claims := jwt.RegisteredClaims{
		ExpiresAt : jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		IssuedAt : jwt.NewNumericDate(time.Now().UTC()),
		Issuer : "chirpy",
		Subject: idString,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)
	if err != nil {
		return "", nil
	}

	return ss, nil

}


func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error){
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error){
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unknown signin type, cannot proceed: " + token.Header["alg"].(string))
		}
		return []byte(tokenSecret), nil
	}, jwt.WithLeeway(5*time.Second))
	if err != nil {
		return uuid.Nil, err
	}
	if !token.Valid{
		return uuid.Nil, errors.New("Invalid token")
	}

	if claims.Subject == ""{
		return uuid.Nil, errors.New("Subject token not found")
	}

	userId, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, errors.New(fmt.Sprintf("Error to convert Subject into uuid: %v", err))
	}

	return userId, nil
}