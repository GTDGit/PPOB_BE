package hash

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = 12
)

// HashPIN hashes a PIN using bcrypt
func HashPIN(pin string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pin), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPIN verifies a PIN against its hash
func VerifyPIN(pin, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pin))
	return err == nil
}

// HashToken creates a SHA256 hash of a token (for refresh token storage)
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// IsWeakPIN checks if a PIN is too weak
func IsWeakPIN(pin string) bool {
	weakPINs := []string{
		"000000", "111111", "222222", "333333", "444444",
		"555555", "666666", "777777", "888888", "999999",
		"123456", "654321", "123123", "112233", "121212",
		"123321", "696969", "000001", "100000",
	}

	for _, weak := range weakPINs {
		if pin == weak {
			return true
		}
	}

	// Check for sequential patterns
	if isSequential(pin) {
		return true
	}

	// Check for repeated patterns
	if isRepeated(pin) {
		return true
	}

	return false
}

// isSequential checks if the PIN is a sequential number
func isSequential(pin string) bool {
	if len(pin) < 3 {
		return false
	}

	// Check ascending
	ascending := true
	for i := 1; i < len(pin); i++ {
		if pin[i] != pin[i-1]+1 {
			ascending = false
			break
		}
	}
	if ascending {
		return true
	}

	// Check descending
	descending := true
	for i := 1; i < len(pin); i++ {
		if pin[i] != pin[i-1]-1 {
			descending = false
			break
		}
	}

	return descending
}

// isRepeated checks if the PIN has too many repeated digits
func isRepeated(pin string) bool {
	if len(pin) < 3 {
		return false
	}

	// Check if all digits are the same
	allSame := true
	for i := 1; i < len(pin); i++ {
		if pin[i] != pin[0] {
			allSame = false
			break
		}
	}

	return allSame
}
