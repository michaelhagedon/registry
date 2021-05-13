package common

import (
	"crypto/rand"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Cost is the cost of the bcrypt digest. This is hard-coded at 10 to match
// the value that Rails/Devise uses. We're porting a Rails database and we
// want to be able to use existing passwords instead of forcing users to
// reset them.
const Cost = 10

// EncryptedTokenPrefix is the prefix that should appear on all our
// encrypted passwords and tokens in the database. Note that it depends
// on common.Cost being set to 10.
const EncryptedTokenPrefix = "$2a$10$"

// EncryptPassword returns the password run through bcrypt 2a with the
// specified salt and cost.
func EncryptPassword(password string) (string, error) {
	digest, err := bcrypt.GenerateFromPassword([]byte(password), Cost)
	if err != nil {
		return "", err
	}
	return string(digest), nil
}

// ComparePasswords compares a plaintext password against a hashed password,
// returning true if they match, false otherwise.
func ComparePasswords(hashed, plaintext string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plaintext)) == nil
}

// RandomToken returns a string of random hex digits suitable for use
// as a secure token.
func RandomToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// LooksEncrypted returns true if string s looks like it's been through our
// EncryptePassword fuction. We use LooksEncrypted on some pgmodels to ensure
// we're not saving unencrypted passwords or tokens.
func LooksEncrypted(s string) bool {
	return strings.HasPrefix(s, EncryptedTokenPrefix)
}
