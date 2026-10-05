//go:build darwin || linux

package researchledger

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
)

func acquireLedgerOwnership(path string) (func() error, error) {
	f, e := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if e != nil {
		return nil, e
	}
	if e = unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); e != nil {
		f.Close()
		return nil, errors.New("research ledger already owned or cannot be locked")
	}
	// Never unlink the sidecar: another owner must lock the same inode. Closing
	// releases ownership even after abrupt process exit, without a stale PID file.
	return f.Close, nil
}
