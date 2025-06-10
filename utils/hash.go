package utils

import (
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes the given password using bcrypt
func HashPassword(password string) (string, error) {
	// Step 1: Create bcrypt hash (returns bytes)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Step 2: Encode to base64 string
	encodedHash := base64.StdEncoding.EncodeToString(hashedBytes)

	// For debugging
	fmt.Printf("HashPassword: Original: %q, Hashed length: %d, Encoded: %q\n",
		password, len(hashedBytes), encodedHash)

	return encodedHash, nil
}

// ComparePasswords compares a hashed password with a plaintext password
// func ComparePasswords(hashedPassword, password string) bool {
// 	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
// 	if err != nil {
// 		fmt.Println("Password mismatch error:", err) // Debugging
// 	}
// 	return err == nil
// }

func ComparePasswords(encodedHash, password string) bool {
	// Step 1: Decode from base64
	hashedBytes, err := base64.StdEncoding.DecodeString(encodedHash)
	if err != nil {
		fmt.Printf("ComparePasswords: Base64 decode error: %v\n", err)
		return false
	}

	// Step 2: Compare with bcrypt
	err = bcrypt.CompareHashAndPassword(hashedBytes, []byte(password))

	// For debugging
	fmt.Printf("ComparePasswords: Password: %q, Decoded hash length: %d, Match: %v\n",
		password, len(hashedBytes), err == nil)
	if err != nil {
		fmt.Printf("ComparePasswords: Error: %v\n", err)
	}

	return err == nil
}
