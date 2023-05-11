// Package main ...
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/lusis/prototokens"
	"github.com/lusis/prototokens/managers/ed25519url"
	tokenpb "github.com/lusis/prototokens/proto/gen/go/prototokens/v1"
)

func main() {
	// build a token
	token, err := prototokens.New(5 * time.Hour)
	// or fully-customized
	// import tokenpb "github.com/lusis/prototokens/proto/gen/go/prototokens/v1"
	// token, err := prototokens.New(
	// 	5*time.Hour,
	// 	prototokens.WithID(ksuid.New().String()),
	// 	prototokens.WithSID(ksuid.New().String()),
	// 	prototokens.WithVendor([]byte("my-custom-data")),
	// 	prototokens.WithUsages(tokenpb.TokenUsages_TOKEN_USAGES_EXCHANGE),
	// )
	if err != nil {
		panic(err)
	}
	fmt.Printf("original token: %+v\n", token)

	// now we sign the token
	// for that we need a token manager and a valid 32byte key seed for ed25519
	manager, err := ed25519url.New(keyfunc)
	if err != nil {
		panic(err)
	}
	signedToken, err := manager.Sign(context.TODO(), token)
	if err != nil {
		panic(err)
	}
	fmt.Printf("signed token: %+v\n", signedToken)

	if manager.Validate(context.TODO(), signedToken) != nil {
		fmt.Println("token not valid")
	}
	// get a url-safe string of the signed token
	encoded, err := manager.Encode(context.TODO(), signedToken)
	if err != nil {
		panic(err)
	}
	fmt.Printf("encoded signed token: %s\n", encoded)

	// decode the token
	decoded, err := manager.Decode(context.TODO(), encoded)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded signed token: %+v\n", decoded)

	// now we can get the validated token
	vt, err := manager.GetValidatedToken(context.TODO(), decoded)
	if err != nil {
		panic(err)
	}
	fmt.Printf("validated token: %+v\n", vt)

	// you can also check if a token has a specific usage allowed
	if err := manager.ValidFor(context.TODO(), decoded, tokenpb.TokenUsages_TOKEN_USAGES_ROTATION); err != nil {
		fmt.Println(err.Error())
	}
}

func keyfunc(_ context.Context) []byte {
	// random string I generated. if this func returns a different value each time, validation will fail
	return []byte("rcuTQ6KzLQ5G5lcAzzXqcs0cdem5i8zy")
}
