package ed25519url

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"testing"
	"time"

	"github.com/lusis/prototokens"
	tokenpb "github.com/lusis/prototokens/proto/gen/go/prototokens/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/stretchr/testify/require"
)

func TestImplements(t *testing.T) {
	require.Implements(t, (*prototokens.TokenManager)(nil), &Manager{}, "should implement the interface")
}

func TestNotValidBefore(t *testing.T) {
	// we're going to use the helper to generate our base values
	// and then resign the token after changing the timestamps
	data, err := setupTest(t.Name(), nil)
	require.NoError(t, err)
	_, pt, m := data.st, data.pt, data.m

	cloned := proto.Clone(pt).(*tokenpb.ProtoToken)
	fiveMin := timestamppb.New(
		time.Now().Add(5 * time.Minute).UTC(),
	)
	cloned.Timestamps.NotValidBefore = fiveMin
	st, err := m.Sign(context.Background(), cloned)
	require.NoError(t, err, "should sign")

	err = m.Validate(context.Background(), st)
	require.ErrorIs(t, err, prototokens.ErrNotYetValid)
}

func TestNoLongerValid(t *testing.T) {
	// we're going to use the helper to generate our base values
	// and then resign the token after changing the timestamps
	data, err := setupTest(t.Name(), nil)
	require.NoError(t, err)
	_, pt, m := data.st, data.pt, data.m

	cloned := proto.Clone(pt).(*tokenpb.ProtoToken)
	fiveMinAgo := timestamppb.New(
		time.Now().Add(-5 * time.Minute).UTC(),
	)
	cloned.Timestamps.NotValidAfter = fiveMinAgo
	st, err := m.Sign(context.Background(), cloned)

	err = m.Validate(context.Background(), st)
	require.ErrorIs(t, err, prototokens.ErrNoLongerValid)
}

func TestHappyPath(t *testing.T) {
	data, err := setupTest(t.Name(), nil)
	require.NoError(t, err)

	vt, err := data.m.GetValidatedToken(context.Background(), data.st)
	require.NoError(t, err, "should not error")
	require.NotNil(t, vt, "validated token should not be nil")
	// this is technicaly just retesting the validation logic but whatevs
	require.Equal(t, data.pt.GetId(), vt.GetId(), "id should match")
	require.Equal(t, data.pt.GetSid(), vt.GetSid(), "sid should match")
	require.True(t, bytes.Equal(data.pt.GetVendor(), vt.GetVendor()), "vendor should match")
	require.Equal(t, data.pt.GetTimestamps().GetNotValidAfter().String(), vt.GetTimestamps().GetNotValidAfter().String(), "not valid after timestamps should match")
	require.Equal(t, data.pt.GetTimestamps().GetNotValidBefore().String(), vt.GetTimestamps().GetNotValidBefore().String(), "not valid before timestamps should match")
	require.Equal(t, data.pt.GetUsages(), vt.GetUsages())
	require.NoError(t, data.m.ValidFor(context.Background(), data.st, tokenpb.TokenUsages_TOKEN_USAGES_HUMAN))
}

func TestEncodeDecode(t *testing.T) {
	data, err := setupTest(t.Name(), nil)
	require.NoError(t, err)
	st, pt, m := data.st, data.pt, data.m
	enc, err := m.Encode(context.Background(), st)
	require.NoError(t, err, "should encode")
	require.NotEmpty(t, enc, "encoded value should not be empty")
	t.Logf("encoded token: %s", enc)

	dec, err := m.Decode(context.Background(), enc)
	require.NoError(t, err, "should decode")
	require.NotNil(t, dec, "signed token should not be nil")
	vtok, err := m.GetValidatedToken(context.Background(), dec)
	require.NoError(t, err, "should be able to get validated token from decoded value")
	require.NotNil(t, vtok, "validated token should not be nil")

	require.True(t, proto.Equal(st, dec), "original and decoded signed token should be the same")
	require.True(t, proto.Equal(vtok, pt), "original and decoded token should be the same")
}
func TestValidation(t *testing.T) {
	data, err := setupTest(t.Name(), nil)
	require.NoError(t, err)
	st, pt, m := data.st, data.pt, data.m

	validationTestCases := map[string]struct {
		expectedError error
		errText       string
		token         *tokenpb.ProtoToken
	}{
		"tamper-id": {
			expectedError: prototokens.ErrTamper,
			token: func() *tokenpb.ProtoToken {
				cloned := proto.Clone(pt).(*tokenpb.ProtoToken)
				cloned.Id = "mismatch"
				return cloned
			}(),
		},
		"tamper-sid": {
			expectedError: prototokens.ErrTamper,
			token: func() *tokenpb.ProtoToken {
				cloned := proto.Clone(pt).(*tokenpb.ProtoToken)
				cloned.Sid = "mismatch"
				return cloned
			}(),
		},
		"tamper-vendor": {
			expectedError: prototokens.ErrTamper,
			token: func() *tokenpb.ProtoToken {
				cloned := proto.Clone(pt).(*tokenpb.ProtoToken)
				cloned.Vendor = []byte("mismatch")
				return cloned
			}(),
		},
		"tamper-usages": {
			expectedError: prototokens.ErrTamper,
			token: func() *tokenpb.ProtoToken {
				cloned := proto.Clone(pt).(*tokenpb.ProtoToken)
				cloned.Usages = append(cloned.Usages, tokenpb.TokenUsages_TOKEN_USAGES_ROTATION)
				return cloned
			}(),
		},
		"tamper-ts-before": {
			expectedError: prototokens.ErrTamper,
			token: func() *tokenpb.ProtoToken {
				cloned := proto.Clone(pt).(*tokenpb.ProtoToken)
				cloned.Timestamps.NotValidBefore = timestamppb.New(time.Now())
				return cloned
			}(),
		},
		"tamper-ts-error": {
			expectedError: prototokens.ErrTamper,
			token: func() *tokenpb.ProtoToken {
				cloned := proto.Clone(pt).(*tokenpb.ProtoToken)
				cloned.Timestamps.NotValidAfter = timestamppb.New(time.Now())
				return cloned
			}(),
		},
	}

	for vName, tc := range validationTestCases {
		t.Run(vName, func(t *testing.T) {
			tokbytes, err := proto.Marshal(tc.token)
			require.NoError(t, err, "token should marshal")
			require.NotNil(t, tokbytes, "tokbytes should not be nil")
			lst, ok := proto.Clone(st).(*tokenpb.SignedToken)
			require.True(t, ok, "should cast")
			lst.Prototoken = tokbytes
			ctx := context.Background()
			verr := m.Validate(ctx, lst)
			require.ErrorIs(t, verr, tc.expectedError, "expected error should match err")
		})
	}

	t.Run("invalid-token", func(t *testing.T) {
		d, err := setupTest(t.Name(), nil)
		require.NoError(t, err)

		// we need the signed token and the manager for this test
		st, m := d.st, d.m
		st.Prototoken = []byte("[]")
		err = m.Validate(context.Background(), st)
		require.ErrorIs(t, err, prototokens.ErrUnmarshal)
	})
}

func TestInvalid(t *testing.T) {
	data, err := setupTest(t.Name(), nil)
	require.NoError(t, err)

	st, _, m := data.st, data.pt, data.m
	err = m.ValidFor(context.Background(), st, tokenpb.TokenUsages_TOKEN_USAGES_MACHINE)
	require.ErrorIs(t, err, prototokens.ErrNotValidForUsage)
}

type setupData struct {
	pt *tokenpb.ProtoToken
	st *tokenpb.SignedToken
	m  prototokens.TokenManager
}

// inlining workaround
var benchVerifyErr error
var benchSignTok *tokenpb.SignedToken
var benchSignErr error

func BenchmarkVerify(b *testing.B) {
	b.StopTimer()

	d, _ := setupTest(b.Name(), nil)

	for i := 0; i < b.N; i++ {
		m := d.m
		st := d.st
		b.StartTimer()
		err := m.Validate(context.Background(), st)
		b.StopTimer()
		benchVerifyErr = err
	}
}

func BenchmarkSign(b *testing.B) {
	b.StopTimer()

	d, _ := setupTest(b.Name(), nil)
	newtok, _ := prototokens.New(5 * time.Minute)

	for i := 0; i < b.N; i++ {
		m := d.m
		pt := newtok
		b.StartTimer()
		st, err := m.Sign(context.Background(), pt)
		b.StopTimer()
		benchSignErr = err
		benchSignTok = st
	}
}

func setupTest(testName string, keyDataFunc KeyDataFunc) (*setupData, error) {
	if keyDataFunc == nil {
		keydata := make([]byte, 32)
		_, err := io.ReadFull(rand.Reader, keydata)
		if err != nil {
			return nil, err
		}
		keyDataFunc = func(_ context.Context) []byte {
			return keydata
		}
	}

	m, err := New(keyDataFunc)
	if err != nil {
		return nil, err
	}

	pt, err := prototokens.New(
		1*time.Hour,
		prototokens.WithID(testName+"_id"),
		prototokens.WithSID(testName+"_sid"),
		prototokens.WithVendor([]byte(testName+"_vendor")),
		prototokens.WithUsages(tokenpb.TokenUsages_TOKEN_USAGES_HUMAN),
	)
	if err != nil {
		return nil, err
	}

	st, err := m.Sign(context.TODO(), pt)
	if err != nil {
		return nil, err
	}
	return &setupData{
		pt: pt,
		st: st,
		m:  m,
	}, nil
}
