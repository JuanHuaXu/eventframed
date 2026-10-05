package researchcalendar

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
)

// BindingDigest binds the unshortened query, not merely its retrieval focus.
// It is an audit join key, not authentication or a privacy guarantee.
func BindingDigest(ctx context.Context) (string, error) {
	b, ok := ctx.Value(bindingKey{}).(binding)
	if !ok || b.tenant == "" || len(b.original) > 4096 {
		return "", errors.New("missing calendar binding")
	}
	data, err := json.Marshal([2]string{b.tenant, b.original})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
