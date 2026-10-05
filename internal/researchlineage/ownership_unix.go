//go:build darwin || linux

package researchlineage

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

func acquireOwnership(path string) (func() error, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = f.Close()
		return nil, errors.New("research lineage already owned or cannot be locked")
	}
	return f.Close, nil
}
