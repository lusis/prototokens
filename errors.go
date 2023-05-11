package prototokens

import "fmt"

var (
	// ErrMarshal is the error from marshalling a token
	ErrMarshal = fmt.Errorf("error marshalling")
	// ErrUnmarshal is the error from unmarshaling a token
	ErrUnmarshal = fmt.Errorf("error unmarshalling")
	// ErrNotValid is the error when a token is invalid
	ErrNotValid = fmt.Errorf("token is invalid")
	// ErrNotValidForUsage is the error when a token is not valid for a specific usage
	ErrNotValidForUsage = fmt.Errorf("token is not valid for provided usages")
	// ErrNotYetValid is the error when a token is not yet valid
	ErrNotYetValid = fmt.Errorf("token is not yet valid")
	// ErrNoLongerValid is the error when a token is not yet valid
	ErrNoLongerValid = fmt.Errorf("token is no longer valid")
	// ErrSign is the error when there is an issue signing a token
	ErrSign = fmt.Errorf("unable to sign token")
	// ErrInvalidSignature is the error when the signature is invalid
	ErrInvalidSignature = fmt.Errorf("signature is invalid")
	// ErrTamper is the error when a signature does not match the token indicating someone tampered with the token bytes
	ErrTamper = fmt.Errorf("token appears to be tampered with")
	// ErrEncode is the error when there is an issue encoding a signed token
	ErrEncode = fmt.Errorf("unable to encode signed token")
	// ErrDecode is the error when there is an issue decoding a signed token
	ErrDecode = fmt.Errorf("unable to decode signed token")
	// ErrKeyData is the error when the private key data is invalid in some way
	ErrKeyData = fmt.Errorf("key data is invalid")
	// ErrOverwrite is the error when you attempt to overwrite a a token's properties after they've been set
	ErrOverwrite = fmt.Errorf("attempted overwrite of field")
	// ErrUnimplemented is the error when a [prototokens.TokenManager] has yet to implement the interface fully
	ErrUnimplemented = fmt.Errorf("functionality not yet implemented")
	// ErrTokenRevoked is the error when a [tokenpb.ProtoToken] has been revoked
	ErrTokenRevoked = fmt.Errorf("token has been revoked")
)
