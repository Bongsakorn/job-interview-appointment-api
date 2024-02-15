package user

import (
	"fmt"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHash(t *testing.T) {
	password := "password"
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)

	fmt.Println(string(bytes))
}
