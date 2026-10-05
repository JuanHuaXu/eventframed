package researchledger

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestMotionBootstrapAtomicBindingAndLegacyRejection(t *testing.T) {
	ctx := context.Background()
	origin, err := json.Marshal(model.Snapshot{RuntimeVersion: 2, ContractVersion: 1})
	if err != nil {
		t.Fatal(err)
	}
	seal := make([]byte, 32)
	seal[0] = 7
	path := t.TempDir() + "/motion.sqlite"
	log, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if foundSeal, foundOrigin, err := log.MotionBootstrap(ctx); err != nil || foundSeal != nil || foundOrigin != nil {
		t.Fatalf("fresh motion metadata: %x %q %v", foundSeal, foundOrigin, err)
	}
	if err = log.BindMotionBootstrap(ctx, seal, origin); err != nil {
		t.Fatal(err)
	}
	if err = log.BindMotionBootstrap(ctx, seal, origin); err != nil {
		t.Fatalf("exact motion retry: %v", err)
	}
	wrongSeal := append([]byte(nil), seal...)
	wrongSeal[0]++
	if err = log.BindMotionBootstrap(ctx, wrongSeal, origin); err == nil {
		t.Fatal("changed motion seal accepted")
	}
	wrongOrigin, _ := json.Marshal(model.Snapshot{RuntimeVersion: 3, ContractVersion: 1})
	if err = log.BindMotionBootstrap(ctx, seal, wrongOrigin); err == nil {
		t.Fatal("changed motion origin accepted")
	}
	if err = log.Close(); err != nil {
		t.Fatal(err)
	}
	log, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	gotSeal, gotOrigin, err := log.MotionBootstrap(ctx)
	if err != nil || !bytes.Equal(gotSeal, seal) || !bytes.Equal(gotOrigin, origin) {
		t.Fatalf("reopened motion metadata: %x %q %v", gotSeal, gotOrigin, err)
	}
	if err = log.Close(); err != nil {
		t.Fatal(err)
	}
	legacy, err := Open(t.TempDir() + "/legacy.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer legacy.Close()
	if err = legacy.BindBootstrap(ctx, seal); err != nil {
		t.Fatal(err)
	}
	if _, _, err = legacy.MotionBootstrap(ctx); err == nil {
		t.Fatal("legacy log exposed a guessed origin")
	}
	if err = legacy.BindMotionBootstrap(ctx, seal, origin); err == nil {
		t.Fatal("legacy log silently upgraded")
	}
}
