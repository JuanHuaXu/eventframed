package researchfusion

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
)

type fakeEmbed struct {
	key  string
	fail bool
	dim  int
}

func (f *fakeEmbed) Dimension() int   { return f.dim }
func (f *fakeEmbed) Name() string     { return "fake" }
func (f *fakeEmbed) ModelKey() string { return f.key }
func (f *fakeEmbed) Embed(s string) ([]float32, error) {
	return f.EmbedDocumentContext(context.Background(), s)
}
func (f *fakeEmbed) EmbedDocumentContext(_ context.Context, s string) ([]float32, error) {
	if f.fail {
		return nil, errors.New("failure")
	}
	return []float32{1, 2}, nil
}
func (f *fakeEmbed) EmbedQueryContext(_ context.Context, s string) ([]float32, error) {
	return []float32{3, 4}, nil
}
func TestMemoOwnershipRolesCaps(t *testing.T) {
	f := &fakeEmbed{key: "a", dim: 2}
	m, err := NewMemo(f, 2)
	if err != nil {
		t.Fatal(err)
	}
	d, _ := m.Embed("same")
	d[0] = 100
	q, _ := m.EmbedQuery("same")
	q[0] = 100
	d, _ = m.EmbedDocument("same")
	q, _ = m.EmbedQuery("same")
	if d[0] != 1 || q[0] != 3 {
		t.Fatal("ownership or roles")
	}
	if _, err = m.Embed("other"); err == nil {
		t.Fatal("cap")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = m.EmbedQueryContext(ctx, "same"); err == nil {
		t.Fatal("canceled hit")
	}
	f.key = "b"
	if _, err = m.Embed("same"); err == nil {
		t.Fatal("model drift")
	}
	if m.Stats().Misses != 2 || m.Stats().Hits != 2 {
		t.Fatal(m.Stats())
	}
}
func TestMemoFailureAndConcurrent(t *testing.T) {
	f := &fakeEmbed{key: "a", dim: 2, fail: true}
	m, _ := NewMemo(f, 2)
	if _, err := m.Embed("x"); err == nil || m.Stats().Entries != 0 {
		t.Fatal("failed insert")
	}
	f.fail = false
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := m.Embed("x")
			if err != nil || v[0] != 1 {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if m.Stats().Misses != 1 || m.Stats().Hits != 15 {
		t.Fatal(m.Stats())
	}
}

func TestMemoIdentityAndTextBounds(t *testing.T) {
	f := &fakeEmbed{key: "a", dim: 2}
	m, _ := NewMemo(f, 2)
	if _, err := m.Embed(strings.Repeat("x", 16385)); err == nil {
		t.Fatal("text cap")
	}
	f.dim = 3
	if m.Dimension() != 2 {
		t.Fatal("bound dimension changed")
	}
	if _, err := m.Embed("x"); err == nil {
		t.Fatal("dimension drift")
	}
	f.dim = 0
	if _, err := NewMemo(f, 2); err == nil {
		t.Fatal("invalid dimension")
	}
}
