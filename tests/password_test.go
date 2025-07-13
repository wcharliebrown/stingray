package tests

import (
	"strings"
	"testing"
	"stingray/auth"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	
	// Test basic password hashing
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	
	// Check that hash is in correct format
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("Hash should start with $argon2id$, got: %s", hash)
	}
	
	// Check that hash contains all required parts
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		t.Errorf("Hash should have 6 parts, got %d: %v", len(parts), parts)
	}
	
	// Verify the password
	valid, err := auth.CheckPassword(password, hash)
	if err != nil {
		t.Fatalf("Failed to check password: %v", err)
	}
	if !valid {
		t.Error("Password verification failed")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "correctpassword"
	wrongPassword := "wrongpassword"
	
	// Hash the password
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	
	// Test correct password
	valid, err := auth.CheckPassword(password, hash)
	if err != nil {
		t.Fatalf("Failed to check correct password: %v", err)
	}
	if !valid {
		t.Error("Correct password should be valid")
	}
	
	// Test wrong password
	valid, err = auth.CheckPassword(wrongPassword, hash)
	if err != nil {
		t.Fatalf("Failed to check wrong password: %v", err)
	}
	if valid {
		t.Error("Wrong password should be invalid")
	}
}

func TestHashUniqueness(t *testing.T) {
	password := "samepassword"
	
	// Hash the same password multiple times
	hash1, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password first time: %v", err)
	}
	
	hash2, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password second time: %v", err)
	}
	
	// Hashes should be different due to random salt
	if hash1 == hash2 {
		t.Error("Hashes should be different due to random salt")
	}
	
	// Both hashes should verify the same password
	valid1, err := auth.CheckPassword(password, hash1)
	if err != nil {
		t.Fatalf("Failed to check password with hash1: %v", err)
	}
	if !valid1 {
		t.Error("Hash1 should verify the password")
	}
	
	valid2, err := auth.CheckPassword(password, hash2)
	if err != nil {
		t.Fatalf("Failed to check password with hash2: %v", err)
	}
	if !valid2 {
		t.Error("Hash2 should verify the password")
	}
}

func TestIsHashFormat(t *testing.T) {
	// Test valid hash format
	password := "testpassword"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	
	if !auth.IsHashFormat(hash) {
		t.Error("Valid hash should be recognized as hash format")
	}
	
	// Test invalid formats
	invalidFormats := []string{
		"plaintext",
		"$bcrypt$",
		"$argon2i$",
		"$argon2d$",
		"",
		"notahash",
	}
	
	for _, invalid := range invalidFormats {
		if auth.IsHashFormat(invalid) {
			t.Errorf("Invalid format should not be recognized as hash: %s", invalid)
		}
	}
}

func TestHashWithCustomParams(t *testing.T) {
	password := "testpassword"
	
	// Test with custom parameters
	params := &auth.Argon2Params{
		Memory:      32 * 1024, // 32 MB
		Iterations:  2,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}
	
	hash, err := auth.HashPasswordWithParams(password, params)
	if err != nil {
		t.Fatalf("Failed to hash password with custom params: %v", err)
	}
	
	// Verify the password
	valid, err := auth.CheckPassword(password, hash)
	if err != nil {
		t.Fatalf("Failed to check password with custom params: %v", err)
	}
	if !valid {
		t.Error("Password verification failed with custom params")
	}
}

func TestMigratePlainTextPassword(t *testing.T) {
	plainPassword := "plaintextpassword"
	
	// Test migration function
	hash, err := auth.MigratePlainTextPassword(plainPassword)
	if err != nil {
		t.Fatalf("Failed to migrate plain text password: %v", err)
	}
	
	// Check that it's a valid hash
	if !auth.IsHashFormat(hash) {
		t.Error("Migrated password should be in hash format")
	}
	
	// Verify the password
	valid, err := auth.CheckPassword(plainPassword, hash)
	if err != nil {
		t.Fatalf("Failed to check migrated password: %v", err)
	}
	if !valid {
		t.Error("Migrated password should verify correctly")
	}
}

func TestPasswordSecurity(t *testing.T) {
	// Test with various password types
	testCases := []string{
		"simple",
		"Complex123!",
		"verylongpasswordwithspecialchars!@#$%^&*()",
		"123456789",
		"",
		"password with spaces",
		"Unicode测试密码",
	}
	
	for _, password := range testCases {
		t.Run("password_"+password, func(t *testing.T) {
			hash, err := auth.HashPassword(password)
			if err != nil {
				t.Fatalf("Failed to hash password '%s': %v", password, err)
			}
			
			valid, err := auth.CheckPassword(password, hash)
			if err != nil {
				t.Fatalf("Failed to check password '%s': %v", password, err)
			}
			if !valid {
				t.Errorf("Password '%s' should verify correctly", password)
			}
		})
	}
} 