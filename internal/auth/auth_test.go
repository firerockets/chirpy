package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test secret"
	expiresIn := time.Duration(2) * time.Second

	token, err := MakeJWT(userID, secret, expiresIn)

	if err != nil {
		t.Errorf("Error creating token: %s", err)
	}

	validatedID, err := ValidadeJWT(token, secret)

	if err != nil {
		t.Errorf("Error validating token: %s", err)
	}

	if validatedID != userID {
		t.Error("User ID does not match with validated ID")
	}
}

func TestMakeJWTWhenTokenExpired(t *testing.T) {
	userID := uuid.New()
	secret := "test secret"
	expiresIn := time.Duration(-1_000_000)

	token, err := MakeJWT(userID, secret, expiresIn)

	if err != nil {
		t.Errorf("Error creating token: %s", err)
	}

	validatedID, err := ValidadeJWT(token, secret)

	if err == nil {
		t.Errorf("Token should be invalid")
	}

	if validatedID == userID {
		t.Error("User ID should not match since token is expired")
	}
}
