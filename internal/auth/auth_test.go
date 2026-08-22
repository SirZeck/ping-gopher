package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecurePassword123!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if !CheckPasswordHash(password, hash) {
		t.Fatalf("Password verification failed for correct password")
	}

	if CheckPasswordHash("WrongPassword!", hash) {
		t.Fatalf("Password verification succeeded for wrong password")
	}
}

func TestJWTTokenGenerationAndValidation(t *testing.T) {
	userID := uuid.New()
	email := "user@pinggopher.com"
	secret := "test-secret-key"

	token, err := GenerateJWTToken(userID, email, secret, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	claims, err := ValidateJWTToken(token, secret)
	if err != nil {
		t.Fatalf("Failed to validate JWT token: %v", err)
	}

	if claims.UserID != userID.String() {
		t.Fatalf("Expected UserID %s, got %s", userID.String(), claims.UserID)
	}
	if claims.Email != email {
		t.Fatalf("Expected Email %s, got %s", email, claims.Email)
	}
}

func TestExpiredJWTTokenValidation(t *testing.T) {
	userID := uuid.New()
	email := "user@pinggopher.com"
	secret := "test-secret-key"

	// Token expires immediately (-1s)
	token, err := GenerateJWTToken(userID, email, secret, -1*time.Second)
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	_, err = ValidateJWTToken(token, secret)
	if err == nil {
		t.Fatalf("Expected validation error for expired token, got nil")
	}
}
