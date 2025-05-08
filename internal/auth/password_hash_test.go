package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	pass := "123qwe"

	hashedPass, err := HashPassword(pass)

	if err != nil {
		t.Errorf("Something went wrong with hashing: %s\n", err)
	}

	err = CheckPasswordHash(hashedPass, pass)

	if err != nil {
		t.Errorf("Hashed password doesn't match: %s\n", err)
	}

	err = CheckPasswordHash(hashedPass, "lorem ipsum")

	if err == nil {
		t.Errorf("Hashed pass should not match with lorem ipsum: %s\n", err)
	}
}
