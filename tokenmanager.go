package prototokens

import (
	"context"

	tokenpb "github.com/lusis/prototokens/proto/gen/go/prototokens/v1"
)

// TokenManager is something that can work with [tokenpb.ProtoToken] and [tokenpb.SignedToken]
type TokenManager interface {
	// Sign signs the token
	Sign(context.Context, *tokenpb.ProtoToken) (*tokenpb.SignedToken, error)
	// Decode decodes a signed token from a string representation
	Decode(context.Context, string) (*tokenpb.SignedToken, error)
	// Encode encodes a signed token as a string
	Encode(context.Context, *tokenpb.SignedToken) (string, error)
	// Validate validates the provided [tokenpb.SignedToken]
	Validate(context.Context, *tokenpb.SignedToken) error
	// ValidFor validates if the token can be used for the provided usages
	ValidFor(context.Context, *tokenpb.SignedToken, tokenpb.TokenUsages) error
	// GetValidatedToken turns a [tokenpb.SignedToken] into a [tokenpb.ProtoToken] after validation
	GetValidatedToken(context.Context, *tokenpb.SignedToken) (*tokenpb.ProtoToken, error)
	// RevokeToken revokes a token
	RevokeToken(context.Context, *tokenpb.ProtoToken) error
}

// UnimplementedTokenManager is a TokenManager implementation designed to be
// used for testing and embedding in other implementations to maintain compatibility
type UnimplementedTokenManager struct{}

// Sign signs the token
func (up *UnimplementedTokenManager) Sign(_ context.Context, _ *tokenpb.ProtoToken) (*tokenpb.SignedToken, error) {
	return nil, ErrUnimplemented
}

// Decode decodes a signed token from a string representation
func (up *UnimplementedTokenManager) Decode(_ context.Context, _ string) (*tokenpb.SignedToken, error) {
	return nil, ErrUnimplemented
}

// Encode encodes a signed token as a string
func (up *UnimplementedTokenManager) Encode(_ context.Context, _ *tokenpb.SignedToken) (string, error) {
	return "", ErrUnimplemented
}

// Validate validates the provided [tokenpb.SignedToken]
func (up *UnimplementedTokenManager) Validate(_ context.Context, _ *tokenpb.SignedToken) error {
	return ErrUnimplemented
}

// ValidFor validates if the token can be used for the provided usages
func (up *UnimplementedTokenManager) ValidFor(_ context.Context, _ *tokenpb.SignedToken, _ tokenpb.TokenUsages) error {
	return ErrUnimplemented
}

// GetValidatedToken turns a [tokenpb.SignedToken] into a [tokenpb.ProtoToken] after validation
func (up *UnimplementedTokenManager) GetValidatedToken(_ context.Context, _ *tokenpb.SignedToken) (*tokenpb.ProtoToken, error) {
	return nil, ErrUnimplemented
}

// RevokeToken revokes a token
func (up *UnimplementedTokenManager) RevokeToken(_ context.Context, _ *tokenpb.ProtoToken) error {
	return ErrUnimplemented
}
