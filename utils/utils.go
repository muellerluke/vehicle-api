package utils

import (
	"math/rand"
	"time"
)

func ValidatePassword(password string) bool {
	//verify password is at least 8 characters, has a number, a capital letter, and has a special character
	if len(password) < 8 {
		return false
	}

	hasNumber := false
	hasCapital := false
	hasSpecial := false

	for _, char := range password {
		if char >= '0' && char <= '9' {
			hasNumber = true
		}

		if char >= 'A' && char <= 'Z' {
			hasCapital = true
		}

		if char == '!' || char == '@' || char == '#' || char == '$' || char == '%' || char == '^' || char == '&' || char == '*' {
			hasSpecial = true
		}
	}

	if !hasNumber || !hasCapital || !hasSpecial {
		return false
	}

	return true
}

func GenerateRandomString(length int) string {
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
