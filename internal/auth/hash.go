package auth

import(
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error){

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil{
		return "", err
	}
	return string(hashPassword), nil

}

func CheckPasswordHash(hash, password string) (error){
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}
	return nil

}