// This package is generate hashing bcrypt
package pkg

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

func NewHashingPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err.Error())
	}
	return string(hashedPassword)
}

func ComparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
