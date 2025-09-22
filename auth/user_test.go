package auth

import (
	"testing"

	"github.com/EggsyOnCode/velho-exchange/internals"
	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	// Test with a new private key
	user := NewUser(nil, 100.0)
	assert.NotNil(t, user, "New user should not be nil")
	assert.NotNil(t, user.PrivateKey, "Private key should not be nil")
	assert.Equal(t, 100.0, user.USD, "User USD balance should be initialized correctly")

	// Test with an existing private key
	privKey := internals.GenerateNewPrivateKey()
	user2 := NewUser(privKey, 50.0)
	assert.NotNil(t, user2, "New user should not be nil")
	assert.Equal(t, privKey, user2.PrivateKey, "Private keys should match")
	assert.Equal(t, 50.0, user2.USD, "User USD balance should be initialized correctly")
}

func TestGenerateUsers(t *testing.T) {
	users := GenerateUsers()
	assert.Len(t, users, 3, "Should generate 3 users")
	for _, user := range users {
		assert.NotNil(t, user.PrivateKey, "Private key should not be nil")
		assert.NotNil(t, user.ID, "User ID should not be nil")
	}
}

func TestGenerateMM(t *testing.T) {
	users, privKeys := GenerateMM()
	assert.Len(t, users, 3, "Should generate 3 users")
	assert.Len(t, privKeys, 3, "Should generate 3 private keys")
	for _, user := range users {
		assert.NotNil(t, user.PrivateKey, "Private key should not be nil")
		assert.NotNil(t, user.ID, "User ID should not be nil")
	}
}
