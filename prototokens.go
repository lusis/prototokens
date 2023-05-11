// Package prototokens contains code for working with [tokenpb.ProtoToken]
package prototokens

import (
	"time"

	tokenpb "github.com/lusis/prototokens/proto/gen/go/prototokens/v1"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/segmentio/ksuid"
)

// New returns a new prototoken valid for the provided duration
func New(duration time.Duration, opts ...TokenOpt) (*tokenpb.ProtoToken, error) {
	if duration == 0 {
		// no you can't shoot yourself in the foot sorry
		// provide a really long duration if you want a long-lived token
		panic("duration MUST be provided")
	}
	nvb := time.Now().UTC()
	nva := nvb.Add(duration)
	tok := &tokenpb.ProtoToken{
		Timestamps: &tokenpb.Timestamps{
			NotValidBefore: timestamppb.New(nvb),
			NotValidAfter:  timestamppb.New(nva),
		},
	}
	for _, opt := range opts {
		if err := opt(tok); err != nil {
			return nil, err
		}
	}
	if tok.GetId() == "" {
		tok.Id = ksuid.New().String()
	}
	return tok, nil
}
