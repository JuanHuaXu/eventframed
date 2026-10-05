package researchpublicresume

import (
	"context"
	"errors"
	"math"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchpublicframe"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicpool"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"google.golang.org/grpc"
)

type fakeReader struct {
	r      *researchpublicpool.Registry
	mutate func([]retrieval.Candidate) []retrieval.Candidate
	calls  int
	err    error
}

func (f *fakeReader) Read(_ context.Context, _ string, digest string) ([]retrieval.Candidate, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	for _, e := range f.r.Entries() {
		if e.SourceSHA256 == digest {
			rows := []retrieval.Candidate{e.Candidate}
			if f.mutate != nil {
				rows = f.mutate(rows)
			}
			return rows, nil
		}
	}
	return nil, nil
}
func fixture(t *testing.T) (*researchpublicpool.Registry, []string, time.Time) {
	t.Helper()
	clock := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)
	r, err := researchpublicpool.New(context.Background(), []researchpublicframe.Record{{ID: "a", Title: "Public title", Text: "Public abstract"}, {ID: "b", Title: "Another title", Text: "Another abstract"}},
		researchpublicframe.Config{TenantID: "t", SessionID: "s", ImportedAt: clock}, "public-scientific")
	if err != nil {
		t.Fatal(err)
	}
	e := r.Entries()
	return r, []string{e[0].Candidate.ID, e[1].Candidate.ID}, clock
}
func TestResumePrefixRequiresActualWholeRead(t *testing.T) {
	r, ids, clock := fixture(t)
	f := &fakeReader{r: r}
	emitted := 0
	err := Verify(context.Background(), r, ids, clock, f, func(i int, e researchpublicpool.Entry, rows []retrieval.Candidate, ns int64) error {
		if i != emitted || e.Candidate.ID != ids[i] || len(rows) != 1 || ns < 0 {
			t.Fatal("bad actual read")
		}
		emitted++
		return nil
	})
	if err != nil || emitted != 2 || f.calls != 2 {
		t.Fatalf("verification incomplete: %v", err)
	}
	for _, bad := range [][]string{nil, {ids[1]}, {ids[0], ids[0]}, {ids[0], ids[1], "new"}} {
		if ValidatePlan(r, bad) == nil {
			t.Fatal("accepted wrong prefix")
		}
	}
}
func TestResumeRejectsMissingCorruptAndFuture(t *testing.T) {
	r, ids, clock := fixture(t)
	mutations := []func([]retrieval.Candidate) []retrieval.Candidate{
		func(v []retrieval.Candidate) []retrieval.Candidate { return nil },
		func(v []retrieval.Candidate) []retrieval.Candidate { return append(v, v[0]) },
		func(v []retrieval.Candidate) []retrieval.Candidate { v[0].ID = "unknown"; return v },
		func(v []retrieval.Candidate) []retrieval.Candidate { v[0].Text += "altered"; return v },
		func(v []retrieval.Candidate) []retrieval.Candidate { v[0].Metadata = []byte("{}"); return v },
		func(v []retrieval.Candidate) []retrieval.Candidate { v[0].Score = math.Inf(1); return v },
	}
	for _, m := range mutations {
		f := &fakeReader{r: r, mutate: m}
		emitted := 0
		if Verify(context.Background(), r, ids, clock, f, func(int, researchpublicpool.Entry, []retrieval.Candidate, int64) error { emitted++; return nil }) == nil || emitted != 0 {
			t.Fatal("accepted corrupt prefix")
		}
	}
	f := &fakeReader{r: r}
	if Verify(context.Background(), r, ids, clock.Add(-time.Second), f, func(int, researchpublicpool.Entry, []retrieval.Candidate, int64) error { return nil }) == nil {
		t.Fatal("accepted unavailable prefix")
	}
}
func TestResumeCancellationAndSinkFailure(t *testing.T) {
	r, ids, clock := fixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	f := &fakeReader{r: r}
	if Verify(ctx, r, ids, clock, f, func(int, researchpublicpool.Entry, []retrieval.Candidate, int64) error { return nil }) == nil || f.calls != 0 {
		t.Fatal("read after cancel")
	}
	f = &fakeReader{r: r}
	if Verify(context.Background(), r, ids, clock, f, func(int, researchpublicpool.Entry, []retrieval.Candidate, int64) error {
		return errors.New("sink failed")
	}) == nil || f.calls != 1 {
		t.Fatal("continued after sink failure")
	}
}

type rpcStub struct{ r *researchpublicpool.Registry }

func TestResumeActualTypedRPCAndOwnedSocket(t *testing.T) {
	r, ids, clock := fixture(t)
	// Darwin's Unix socket pathname limit is smaller than t.TempDir's long
	// test-name path; use a short, physically resolved owned fixture directory.
	tmp, err := os.MkdirTemp("/tmp", "efpr-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmp) })
	cwd, err := filepath.EvalSymlinks(tmp)
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(cwd, "research", "owned")
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	l, err := net.Listen("unix", filepath.Join(root, "native.sock"))
	if err != nil {
		t.Fatal(err)
	}
	srv := grpc.NewServer()
	srv.RegisterService(&grpc.ServiceDesc{ServiceName: "libravdb.ipc.v1.LibravDB", HandlerType: (*any)(nil), Methods: []grpc.MethodDesc{{MethodName: "ListByMeta", Handler: func(_ any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
		q := &wireRequest{}
		if err := dec(q); err != nil {
			return nil, err
		}
		if q.Collection != r.Collection() || q.Key != "source_sha256" {
			return nil, errors.New("wrong request contract")
		}
		for _, e := range r.Entries() {
			if e.SourceSHA256 == q.Value {
				return &wireResponse{Results: []*wireResult{{ID: e.Candidate.ID, Text: e.Candidate.Text, Score: .4, MetadataJSON: e.Candidate.Metadata}}}, nil
			}
		}
		return &wireResponse{}, nil
	}}}}, &rpcStub{r: r})
	done := make(chan error, 1)
	go func() { done <- srv.Serve(l) }()
	defer func() { srv.Stop(); <-done }()
	n, err := Open(root, cwd)
	if err != nil {
		t.Fatal(err)
	}
	defer n.Close()
	if err = Verify(context.Background(), r, ids, clock, n, func(int, researchpublicpool.Entry, []retrieval.Candidate, int64) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if _, err = n.Read(context.Background(), r.Collection(), "short"); err == nil {
		t.Fatal("accepted short digest")
	}
	alias := filepath.Join(cwd, "research", "alias")
	if err = os.Symlink(root, alias); err != nil {
		t.Fatal(err)
	}
	if _, err = Open(alias, cwd); err == nil {
		t.Fatal("accepted symlink root")
	}
	if _, err = Open(cwd, cwd); err == nil {
		t.Fatal("accepted root outside research")
	}
}
