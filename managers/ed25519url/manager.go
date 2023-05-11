package ed25519url

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/lusis/prototokens"
	"github.com/lusis/prototokens/internal"
	tokenpb "github.com/lusis/prototokens/proto/gen/go/prototokens/v1"

	"google.golang.org/protobuf/proto"
)

// Manager is an implementation of [prototokens.TokenManager] that:
// - signs tokens with ed25519 with [keyDataFunc] returning the seed that will be passed to [ed25519.NewFromSeed]
// - encodes/decodes with [base64.RawURLEncoding.EncodeToString]
type Manager struct {
	*prototokens.UnimplementedTokenManager
	keyDataFunc KeyDataFunc
}

// KeyDataFunc is a func that can return the seed passed to [ed25519.NewFromSeed] and optional error
type KeyDataFunc func(context.Context) []byte

// New returns a new [sharedkey.Manager]
func New(keyDataFunc KeyDataFunc) (*Manager, error) {
	// get the seed an make sure it's valid
	seed := keyDataFunc(context.Background())
	seedlen := len(seed)
	if seedlen != ed25519.SeedSize {
		return nil, fmt.Errorf("%w: invalid seed size returned (want: %d have: %d)", prototokens.ErrKeyData, ed25519.SeedSize, seedlen)
	}

	return &Manager{keyDataFunc: keyDataFunc}, nil
}

// GetValidatedToken turns a [tokenpb.SignedToken] into a [tokenpb.ProtoToken] after validation
func (skm *Manager) GetValidatedToken(ctx context.Context, token *tokenpb.SignedToken) (*tokenpb.ProtoToken, error) {
	ctx, span := internal.StartSpan(ctx, "GetValidatedToken")
	defer span.End()
	if err := skm.Validate(ctx, token); err != nil {
		return nil, fmt.Errorf("%w: %w", prototokens.ErrNotValid, err)
	}
	pt := &tokenpb.ProtoToken{}
	if err := proto.Unmarshal(token.GetPrototoken(), pt); err != nil {
		return nil, fmt.Errorf("%w: %w", prototokens.ErrUnmarshal, err)
	}
	return pt, nil
}

// Sign signs the token
func (skm *Manager) Sign(ctx context.Context, pt *tokenpb.ProtoToken) (*tokenpb.SignedToken, error) {
	ctx, span := internal.StartSpan(ctx, "Sign")
	defer span.End()
	span.AddEvent("marshal start")
	b, err := proto.Marshal(pt)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", prototokens.ErrMarshal, err)
	}
	span.AddEvent("marshal end")

	sig, err := skm.sign(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", prototokens.ErrSign, err)
	}
	st := &tokenpb.SignedToken{
		Signature:  sig,
		Prototoken: b,
	}
	return st, nil
}

// ValidFor checks if a token is valid for a specific usage
func (skm *Manager) ValidFor(ctx context.Context, st *tokenpb.SignedToken, usage tokenpb.TokenUsages) error {
	ctx, span := internal.StartSpan(ctx, "ValidFor")
	defer span.End()
	tok, err := skm.GetValidatedToken(ctx, st)
	if err != nil {
		return err
	}
	for _, u := range tok.GetUsages() {
		if u == usage {
			// got a hit
			return nil
		}
	}
	return prototokens.ErrNotValidForUsage
}

// Validate checks if the token is valid
// we do the validation in layers based on how expensive it is to validate
// - unmarshal the token bytes. we need to do that for the later checks. failure means its not valid
// - validate the signature to ensure the message hasn't been tampered with
// - check timestamps in the token now that we know we can trust it
func (skm *Manager) Validate(ctx context.Context, st *tokenpb.SignedToken) error {
	ctx, span := internal.StartSpan(ctx, "Validate")
	defer span.End()
	now := time.Now().UTC()
	tok := &tokenpb.ProtoToken{}
	if err := proto.Unmarshal(st.GetPrototoken(), tok); err != nil {
		return fmt.Errorf("%w: %w", prototokens.ErrUnmarshal, err)
	}
	// validate the signature
	if err := skm.verify(ctx, st.GetSignature(), st.GetPrototoken()); err != nil {
		return fmt.Errorf("%w: %w", prototokens.ErrInvalidSignature, err)
	}

	// we know the token is valid so we can do our other checks
	nvb := tok.GetTimestamps().GetNotValidBefore().AsTime().UTC()
	if now.Before(nvb) {
		return prototokens.ErrNotYetValid
	}

	nva := tok.GetTimestamps().GetNotValidAfter().AsTime().UTC()
	if now.After(nva) {
		return prototokens.ErrNoLongerValid
	}

	return nil
}

// Encode encodes a signed token as a url-safe string
func (skm *Manager) Encode(ctx context.Context, st *tokenpb.SignedToken) (string, error) {
	_, span := internal.StartSpan(ctx, "Encode")
	defer span.End()

	// marshal first
	b, err := proto.Marshal(st)
	if err != nil {
		return "", fmt.Errorf("%w: %w", prototokens.ErrMarshal, err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// Decode decodes a signed token from a url-safe string representation
func (skm *Manager) Decode(ctx context.Context, s string) (*tokenpb.SignedToken, error) {
	_, span := internal.StartSpan(ctx, "Decode")
	defer span.End()

	st := &tokenpb.SignedToken{}
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", prototokens.ErrDecode, err)
	}
	if err := proto.Unmarshal(b, st); err != nil {
		return nil, fmt.Errorf("%w: %w", prototokens.ErrUnmarshal, err)
	}
	return st, nil
}

func (skm *Manager) sign(ctx context.Context, data []byte) ([]byte, error) { // nolint: unparam
	seed := skm.keyDataFunc(ctx)

	priv := ed25519.NewKeyFromSeed(seed)
	sig := ed25519.Sign(priv, data)
	return sig, nil
}

func (skm *Manager) verify(ctx context.Context, sig []byte, data []byte) error {
	seed := skm.keyDataFunc(ctx)

	pub := ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey)
	if verified := ed25519.Verify(pub, data, sig); !verified {
		return prototokens.ErrTamper
	}
	return nil
}
