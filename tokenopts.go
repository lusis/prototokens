package prototokens

import (
	"fmt"

	tokenpb "github.com/lusis/prototokens/proto/gen/go/prototokens/v1"
)

// TokenOpt is an option for creating a token
type TokenOpt func(*tokenpb.ProtoToken) error

// WithID provides a custom id for the new token
func WithID(id string) TokenOpt {
	return func(pt *tokenpb.ProtoToken) error {
		if id == "" {
			return fmt.Errorf("id cannot be empty")
		}
		// this shouldn't happen since these are only used by New
		// but might as well prevent a footgun
		if pt.GetId() != "" {
			return fmt.Errorf("%w: id", ErrOverwrite)
		}
		pt.Id = id
		return nil
	}
}

// WithUsages sets the valid usages for a token
func WithUsages(usages ...tokenpb.TokenUsages) TokenOpt {
	return func(pt *tokenpb.ProtoToken) error {
		if len(usages) == 0 {
			return fmt.Errorf("at least one usage must be provided")
		}
		if pt.GetUsages() != nil {
			return fmt.Errorf("%w: usages", ErrOverwrite)
		}
		pt.Usages = usages
		return nil
	}
}

// WithSID provides a custom sid for the new token
func WithSID(sid string) TokenOpt {
	return func(pt *tokenpb.ProtoToken) error {
		if sid == "" {
			return fmt.Errorf("%w: sid", ErrOverwrite)
		}
		// this shouldn't happen since these are only used by New
		// but might as well prevent a footgun
		if pt.GetSid() != "" {
			return fmt.Errorf("%w: sid", ErrOverwrite)
		}
		pt.Sid = sid
		return nil
	}
}

// WithVendor populates the vendor field for the new token
func WithVendor(data []byte) TokenOpt {
	return func(pt *tokenpb.ProtoToken) error {
		if data == nil {
			return fmt.Errorf("data cannot be empty")
		}
		if pt.GetVendor() != nil {
			return fmt.Errorf("%w: vendor", ErrOverwrite)
		}
		pt.Vendor = data
		return nil
	}
}
