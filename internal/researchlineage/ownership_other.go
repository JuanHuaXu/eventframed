//go:build !darwin && !linux

package researchlineage

import "errors"

func acquireOwnership(string) (func() error, error) {
	return nil, errors.New("research lineage requires process ownership support")
}
