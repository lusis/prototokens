package prototokens

import (
	"context"
)

// RevocationStorer is an interface for storing tokens that have been revoked
type RevocationStorer interface {
	// Revoke revokes a [tokenpb.ProtoToken] by its identifier
	// revocationID does not have to be the token's id or sid even
	// you can just use the signature if you want
	Revoke(ctx context.Context, revocationID string) error
	// CheckRevocation checks the store to see if the token has been revoked
	CheckRevocation(ctx context.Context, revocatonID string) error
}

// UnimplementedRevocationStorer is an implementation for testing and compatibility
type UnimplementedRevocationStorer struct{}

// Revoke revokes a [tokenpb.ProtoToken] by its identifier
// revocationID does not have to be the token's id or sid even
// you can just use the signature if you want
func (urs *UnimplementedRevocationStorer) Revoke(_ context.Context, _ string) error {
	return ErrUnimplemented
}

// CheckRevocation checks the store to see if the token has been revoked
func (urs *UnimplementedRevocationStorer) CheckRevocation(_ context.Context, _ string) error {
	return ErrUnimplemented
}
