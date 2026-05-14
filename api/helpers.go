package api

import "golang.org/x/crypto/bcrypt"

func hashPassword(s string) (string, error) {
	hashByte, err := bcrypt.GenerateFromPassword([]byte(s), 10)
	if err != nil {
		return "", err
	}
	hashedPassword := string(hashByte)
	return hashedPassword, nil
}
