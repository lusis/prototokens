package prototokens

import (
	"fmt"
	"testing"
	"time"

	tokenpb "github.com/lusis/prototokens/proto/gen/go/prototokens/v1"
	"github.com/stretchr/testify/require"
)

func failingtokenopt() TokenOpt {
	return func(pt *tokenpb.ProtoToken) error {
		return fmt.Errorf("snarf!")
	}
}
func TestTokenOpts(t *testing.T) {
	testCases := map[string]struct {
		err  bool
		opts []TokenOpt
	}{
		"all-opts": {
			opts: []TokenOpt{
				WithID(t.Name()),
				WithUsages(tokenpb.TokenUsages_TOKEN_USAGES_HUMAN, tokenpb.TokenUsages_TOKEN_USAGES_EXCHANGE),
				WithSID(t.Name() + "_sid"),
				WithVendor([]byte("vendor data")),
			},
		},
		"invalid-id": {
			err: true,
			opts: []TokenOpt{
				WithID(""),
			},
		},
		"invalid-sid": {
			err: true,
			opts: []TokenOpt{
				WithSID(""),
			},
		},
		"invalid-usages": {
			err: true,
			opts: []TokenOpt{
				WithUsages(),
			},
		},
		"invalid-vendor": {
			err: true,
			opts: []TokenOpt{
				WithVendor(nil),
			},
		},
		"custom-option": {
			err: true,
			opts: []TokenOpt{
				failingtokenopt(),
			},
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			pt, err := New(5*time.Minute, tc.opts...)
			if tc.err {
				require.Error(t, err, "should error")
				require.Nil(t, pt, "token should be nil")
			} else {
				require.NoError(t, err, "should not error")
				require.NotNil(t, pt, "token should not be nil")
			}
		})
	}
}

func TestTokenOptsOverwrite(t *testing.T) {
	testCases := map[string]TokenOpt{
		"id":     WithID(t.Name()),
		"sid":    WithSID(t.Name()),
		"usages": WithUsages(tokenpb.TokenUsages_TOKEN_USAGES_HUMAN),
		"vendor": WithVendor([]byte(t.Name())),
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			// bascially we attempt to pass in the same param twice which should trigger an overwrite error
			// we don't care if it's the same or not, we shouldn't allow it
			pt, err := New(5*time.Minute, tc, tc)
			require.ErrorIs(t, err, ErrOverwrite, "should be an overwrite error")
			require.Contains(t, err.Error(), n, "field should be in error message")
			require.Nil(t, pt, "token should be nil")
		})
	}
}
