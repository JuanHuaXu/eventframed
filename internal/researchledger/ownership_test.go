package researchledger

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRejectSecondLedgerOwner(t *testing.T) {
	path := filepath.Join(t.TempDir(), "owned.sqlite")
	first, e := Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer first.Close()
	second, e := Open(path)
	if e == nil {
		second.Close()
		t.Fatal("second independent ledger owner accepted")
	}
	alias := filepath.Join(t.TempDir(), "alias.sqlite")
	if e = os.Symlink(path, alias); e != nil {
		t.Fatal(e)
	}
	if duplicate, e := Open(alias); e == nil {
		duplicate.Close()
		t.Fatal("symlink bypassed ownership")
	}
	if e = first.Close(); e != nil {
		t.Fatal(e)
	}
	third, e := Open(path)
	if e != nil {
		t.Fatal("ownership not released", e)
	}
	third.Close()
}
