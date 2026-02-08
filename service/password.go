package service

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string, cost int) (string, error) {
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
