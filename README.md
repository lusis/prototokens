# prototokens
This is an implementation of Protobuf tokens as described in the [fly.io blog post on API tokens/keys](https://fly.io/blog/api-tokens-a-tedious-survey/)

# Motivation
I've implemented various api key strategies multiple times (including JWTs and a couple of variations based on protobufs) over my career (including tokens based on protobufs)

I generally enjoy working with protobufs and wanted to see if I could build a reusable implementation of the idea as described and give myself something opensource to use again in the future.

# Usage

For the most part you don't have to care about protocol buffers at all.

My goals are hopefully that:
- no one should need to pull in any third-party repos explicitly to use it out of the box
- it shouldn't let you do something "bad"

When working with the tokens, you'll want to use `prototokens.New` to create a `ProtoToken` (though nothing stops you from just creating one yourself from the generated code). Generally you'll be passing around a `SignedToken` and extracting properties from a `ProtoToken` contained in that `SignedToken`

You'll also need a `TokenManager` implementation. The only shipped implementation uses ed25519 as described in the blog post.

## `ProtoToken` and `SignedToken`
The two protobuf types we're working with are as follows:

```proto3
message SignedToken {
    bytes signature = 1;
    bytes prototoken = 2;
}

message ProtoToken {
    // id is used for revocation and other purposes
    // tokens without ids cannot be checked for revocation
    string id = 1;
    // secondary id such as a primary group id of some kind
    string sid = 2;
    // opaque data to be passed across the token if any
    bytes vendor = 3;
    // some canned usages for tokens if desired
    repeated TokenUsages usages = 4;
    // timestamp data
    Timestamps timestamps = 15;
}
```

The highlevel idea is that you create a `ProtoToken` and sign it. Marshal the signature and the original token to bytes and create the `SignedToken`. In most cases you'll be passing around a `SignedToken` to the `TokenManager` interface.

As with proto3 in general, no fields are required but an empty token won't get you much. Working with these types is described below.

## Imports

If you just want to use what the repo ships with, you can import the root and the shipped implemenation:

```go
import (
    "github.com/lusis/prototokens"
    "github.com/lusis/prototokens/managers/ed25519url"
)
```

If you're building your own implementation (or creating tokens with usage restrictions), you'll have to pull in the generated code but you won't need the `ed25519url` implementation:

```go
import (
    "github.com/lusis/prototokens"
    tokenpb "github.com/lusis/prototokens/proto/gen/go/prototokens/v1"
)

type MyCustomTokenManager struct {
    // compatibility embedding
    *prototokens.UnimplementedTokenManager
}
```

## Creating a new token
When creating a token, you are REQUIRED to pass in a valid `time.Duration`.
The length is not checked so you could make it for 5 years but that's your call

```go
token, err := prototokens.New(1 * time.Hour)
```

or fully customize everything

```go
import tokenpb "github.com/lusis/prototokens/proto/gen/go/prototokens/v1"
token, err := prototokens.New(
    5*time.Hour,
    prototokens.WithID("mycustomid"),
    prototokens.WithSID("mycustomsubid"),
    prototokens.WithVendor([]byte("my-custom-data")),
    prototokens.WithUsages(tokenpb.TokenUsages_TOKEN_USAGES_EXCHANGE),
)
```

## Creating a token manager

```go
manager, err := ed25519url.New(keyfunc)
```

where keyfunc is a `func(context.Context) []byte`.

Now that you have a `TokenManager` you can do most of the "fun" stuff

## Signing a token
*(Signing and encoding are two different steps)*

```go
signedToken, err := manager.Sign(ctx, token)
```

## Validating a token
There are a few different ways to validate the token based on what decision you need to make:

### Is the token valid?
You might not actually care about the semantics of a prototoken under the covers. Maybe you don't need usages or anything. You can just check if the token is valid. This is a handy mechanism for passing around a trusted temporary string

```go
if manager.Validate(ctx, signedToken) != nil {
    // forbidden
}
```

### Getting the validated `ProtoToken`
If you don't use the usages concept, you can call `GetValidatedToken` and a trusted `ProtoToken` back,
Because we're working with protobufs, you should generally use the getters provided in the generated code to avoid accidental panics:

```go
vt, err := manager.GetValidatedToken(ctx, signedToken)
id := vt.GetId() // note the capitalization of GetId and GetSid - this is how protoc-gen-go generates getters as opposed to GetID() and GetSID() which is more idiomatic
sid := vt.GetSid()
usages := vt.GetUsages()
vendorData := vt.GetVendor()
```

### Checking if a usage is valid
*usages are optional*

```go
err := manager.ValidFor(ctx, signedToken, tokenpb.TokenUsages_TOKEN_USAGES_ROTATION)
```

## Encoding/Decoding a token
Encoding allows you to convert the signed token to a scary string representation for use as an api key.

```go
encoded, err := manager.Encode(ctx, signedToken)
// CkANUi9wA2rOQkCXrkcf3GhB4K7yjk-jXPdyrAiJkZRK_eJBB1PXJg5TcQXK-qTsYVZJSja9UVVeYkahwCgy72gHEjsKGzJQZjBNYnhOdGJBdTNvVVN3eFRKSmQzZVJEbnocCgwIzf_0ogYQ7eDqmwISDAidjPaiBhDt4OqbAg
```

```go
decoded, err := manager.Decode(ctx, encoded)
```

# Revocation
I've provided an interface for a revocation storer though not provided an implementation. I want to add a couple of basic implementations for common datastores (redis/mysql/pgsql/sqlite) but I'm not ready to support those just yet.

In generally revocation should be baked in to the `TokenManager` implementation such that a call to `GetValidatedToken` ensures that whatever identifier is used in the `RevocationStorer` is able to be calculated or extracted from a `SignedToken`. You could store a hash of the encoded `SignedToken` or the signature but you probably don't want to store the actual encoded `SignedToken` itself.

I plan on adding revocation to the `TokenManager` interface once I'm settled a bit more on the ergonomics of revocation. Using the `RevocationStorer` interface which is why I'm including it.

# Other implementations
The only implementation I found of the same idea outside of the blog post was here:

- https://github.com/ThatsMrTalbot/prototoken

but it seems unmaintained. My implementation largly follows the same pattern mainly because the operations needed are similar across the board.

# Testing
```go
type testTokenManager struct {
    *prototokens.UnimplementedTokenManager
    validateErr error
    signFunc func() (*tokenpb.SignedToken, error)
}

// Sign implements our own signing for tests
func (ttm *testTokenManager) Sign(_ context.Context, _ *tokenpb.ProtoToken) (*tokenpb.SignedToken, error) {
    return ttm.signFunc()
}

// Validate implements our own validation for tests
func (ttm *testTokenManager) Validate(_ context.Context, _ *tokenpb.SignedToken) error {
    return ttm.validateErr
}

func TestMyCode(t *testing.T) {

    testmanager := &testTokenManager{
        validateErr: prototokens.ErrNotValid,
        signFunc: func() (*tokenpb.SignedToken, error) {
            return myprecomputedsignedtoken, nil
        }
    }

    // myservice is something that needs to sign and validate tokens
    myservice := NewMyService(testmanager)
}
```

# Design Decisions

## Usages what?
My experience with tokens/apikeys of any kind is that they generally can be used for very specific things. Think scopes associated with an oauth token.

Usages are my semantics for scopes in prototokens. You don't need to use them but I find them useful for doing something like so:

```go
// generate a token that is only valid for exchange within a 2 minute window
tok, _ := prototokens.New(120*time.Second, prototokens.WithUsages(tokenpb.TokenUsages_TOKEN_USAGES_EXCHANGE))
st, _ := manager.Sign(ctx, tok)
key, _ := manager.Encode(ctx, st)
```

We can give this to token out and require it to be exchanged for a longer lived token:

```go
decoded, _ := manager.Decode(ctx, key)
err := manager.ValidFor(ctx, decoded, tokenpb.TokenUsages_TOKEN_USAGES_EXCHANGE)
if err != nil {
   // return a forbidden 
}
// generate a long-lived token for the same token
longTok, _ := prototokens.New(604800 * time.Second, prototokens.WithID(decoded.GetId()), prototokens.WithUsages(tokenpb.TokenUsages_TOKEN_USAGES_HUMAN))
// sign, encode and return to user
```

## Why are encoding and signing different steps? Why is encoding included at all?
Encoding/decoding is included for convienience and to ensure you shouldn't need to generally pull in any external protobuf deps. Using the wrong proto package can easily happen accidentally or you might want to use your OWN encoding/decoding scheme so the interface allows it.

## Why a keydata func?
I'm paranoid. I honestly didn't want to keep the actual key data in memory myself and risk an issue because of that.

Also using a function type instead of a byte slice directly allows pulling the keydata from an external source at runtime.