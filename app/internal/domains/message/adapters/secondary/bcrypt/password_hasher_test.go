package bcrypt

import (
	"context"
	"strings"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var _ secondary.PasswordHasherPort = NewPasswordHasher(bcrypt.MinCost)

func TestNewPasswordHasher_CostClamping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		inputCost    int
		expectedCost int
	}{
		{
			name:         "cost below MinCost is clamped to DefaultCost",
			inputCost:    bcrypt.MinCost - 1,
			expectedCost: bcrypt.DefaultCost,
		},
		{
			name:         "cost above MaxCost is clamped to DefaultCost",
			inputCost:    bcrypt.MaxCost + 1,
			expectedCost: bcrypt.DefaultCost,
		},
		{
			name:         "MinCost is accepted as-is",
			inputCost:    bcrypt.MinCost,
			expectedCost: bcrypt.MinCost,
		},
		{
			// MinCost+1 avoids DefaultCost/MaxCost values that would add seconds of wall time.
			name:         "in-range cost above MinCost is accepted as-is",
			inputCost:    bcrypt.MinCost + 1,
			expectedCost: bcrypt.MinCost + 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := NewPasswordHasher(tc.inputCost)
			require.NotNil(t, h)

			// The observable contract: hashing a password should succeed, and the
			// resulting hash should have been generated with the expected cost.
			// We verify the stored cost indirectly through bcrypt.Cost().
			hash, err := h.Hash(context.Background(), "probe-password")
			require.NoError(t, err)

			// bcrypt.Cost is the only observable way to verify the cost was clamped.
			actualCost, err := bcrypt.Cost([]byte(hash))
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCost, actualCost)
		})
	}
}

func TestHash_And_Verify_RoundTrip(t *testing.T) {
	t.Parallel()

	h := NewPasswordHasher(bcrypt.MinCost)
	ctx := context.Background()

	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "simple alphanumeric password",
			password: "correct-horse-battery-staple",
		},
		{
			name:     "password with special characters",
			password: "P@$$w0rd!#%^&*()",
		},
		{
			name:     "password with leading and trailing spaces",
			password: " password with spaces ",
		},
		{
			name:     "single character password",
			password: "x",
		},
		{
			name:     "unicode password",
			password: "pässwörð🔑",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hash, err := h.Hash(ctx, tc.password)
			require.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tc.password, hash)

			ok, err := h.Verify(ctx, tc.password, hash)
			require.NoError(t, err)
			assert.True(t, ok)
		})
	}
}

func TestVerify_WrongPassword(t *testing.T) {
	t.Parallel()

	h := NewPasswordHasher(bcrypt.MinCost)
	ctx := context.Background()

	tests := []struct {
		name           string
		hashPassword   string
		verifyPassword string
	}{
		{
			name:           "completely different password",
			hashPassword:   "correct-password",
			verifyPassword: "wrong-password",
		},
		{
			name:           "same password with different casing",
			hashPassword:   "Password",
			verifyPassword: "password",
		},
		{
			name:           "empty string against a real hash",
			hashPassword:   "some-password",
			verifyPassword: "",
		},
		{
			name:           "password with trailing space vs without",
			hashPassword:   "password",
			verifyPassword: "password ",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hash, err := h.Hash(ctx, tc.hashPassword)
			require.NoError(t, err)

			ok, err := h.Verify(ctx, tc.verifyPassword, hash)
			assert.NoError(t, err)
			assert.False(t, ok)
		})
	}
}

func TestHash_RejectsEmptyPassword(t *testing.T) {
	t.Parallel()

	h := NewPasswordHasher(bcrypt.MinCost)
	ctx := context.Background()

	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "empty string",
			password: "",
		},
		{
			name:     "single space",
			password: " ",
		},
		{
			name:     "multiple spaces",
			password: "   ",
		},
		{
			name:     "tab character",
			password: "\t",
		},
		{
			name:     "newline character",
			password: "\n",
		},
		{
			name:     "mixed whitespace",
			password: " \t\n\r ",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hash, err := h.Hash(ctx, tc.password)
			assert.Error(t, err)
			assert.Empty(t, hash)
		})
	}
}

// TestVerify_EmptyHash covers the "no password required" semantic: an empty or
// whitespace-only hash means the resource is unprotected, so Verify returns (true, nil).
func TestVerify_EmptyHash(t *testing.T) {
	t.Parallel()

	h := NewPasswordHasher(bcrypt.MinCost)
	ctx := context.Background()

	tests := []struct {
		name     string
		hash     string
		password string
	}{
		{
			name:     "empty hash with non-empty password",
			hash:     "",
			password: "some-password",
		},
		{
			name:     "empty hash with empty password",
			hash:     "",
			password: "",
		},
		{
			name:     "whitespace-only hash",
			hash:     "   ",
			password: "any-password",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ok, err := h.Verify(ctx, tc.password, tc.hash)
			assert.NoError(t, err)
			assert.True(t, ok)
		})
	}
}

func TestVerify_MalformedHash(t *testing.T) {
	t.Parallel()

	h := NewPasswordHasher(bcrypt.MinCost)
	ctx := context.Background()

	tests := []struct {
		name     string
		hash     string
		password string
	}{
		{
			name:     "random garbage string",
			hash:     "this-is-not-a-bcrypt-hash",
			password: "password",
		},
		{
			name:     "truncated bcrypt prefix only",
			hash:     "$2a$",
			password: "password",
		},
		{
			name:     "base64 but not bcrypt",
			hash:     "aGVsbG8gd29ybGQ=",
			password: "password",
		},
		{
			name:     "bcrypt prefix with wrong cost field",
			hash:     "$2a$99$notavalidhashstring",
			password: "password",
		},
		{
			name:     "looks like bcrypt but truncated body",
			hash:     "$2a$04$abc",
			password: "password",
		},
		{
			name:     "single non-whitespace character",
			hash:     "x",
			password: "password",
		},
		{
			name:     "long random ASCII string that is not a bcrypt hash",
			hash:     strings.Repeat("a", 60),
			password: "password",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ok, err := h.Verify(ctx, tc.password, tc.hash)
			assert.Error(t, err)
			assert.False(t, ok)
		})
	}
}

func TestHash_Uniqueness(t *testing.T) {
	t.Parallel()

	h := NewPasswordHasher(bcrypt.MinCost)
	ctx := context.Background()
	password := "same-password-twice"

	hash1, err := h.Hash(ctx, password)
	require.NoError(t, err)

	hash2, err := h.Hash(ctx, password)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)

	ok1, err := h.Verify(ctx, password, hash1)
	require.NoError(t, err)
	assert.True(t, ok1)

	ok2, err := h.Verify(ctx, password, hash2)
	require.NoError(t, err)
	assert.True(t, ok2)
}
