package security

import "golang.org/x/crypto/bcrypt"

// PasswordHash generates a hash of the password
func PasswordHash(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

// PasswordVerify validates the password against a given hash
func PasswordVerify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
