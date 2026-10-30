package auth

import (
	"testing"
	"time"
)

func TestPasswordHashAndCheck(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(hash, "correct-horse-battery") {
		t.Fatal("expected password to match")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Fatal("expected wrong password to fail")
	}
}

func TestTokenRoundTrip(t *testing.T) {
	manager := NewTokenManager("test-secret", time.Hour)
	token, err := manager.Generate(7, "student@example.com", "student")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := manager.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 7 || claims.Role != "student" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}
