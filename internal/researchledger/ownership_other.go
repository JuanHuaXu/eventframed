//go:build !darwin && !linux

package researchledger

import "errors"

func acquireLedgerOwnership(string) (func() error, error) {
	return nil, errors.New("research ledger ownership unsupported on this platform")
}
