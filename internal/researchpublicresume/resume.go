// Package researchpublicresume verifies a public corpus prefix after reopening
// an owned research store. Receipt success alone never authorizes skipping data.
package researchpublicresume

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchpublicpool"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	oldproto "github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const Method = "/libravdb.ipc.v1.LibravDB/ListByMeta"

type Reader interface {
	Read(context.Context, string, string) ([]retrieval.Candidate, error)
}

type Native struct{ connection *grpc.ClientConn }

// Open rejects symlink escapes and accepts only this task's explicit research
// socket. There is no default endpoint, secret discovery, remote transport or retry.
func Open(root, cwd string) (*Native, error) {
	a, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	resolved, err := filepath.EvalSymlinks(a)
	if err != nil {
		return nil, err
	}
	base, err := filepath.EvalSymlinks(filepath.Join(cwd, "research"))
	if err != nil {
		return nil, err
	}
	if a != resolved || !strings.HasPrefix(resolved, base+string(os.PathSeparator)) {
		return nil, errors.New("unowned resume root")
	}
	socket := filepath.Join(a, "native.sock")
	info, err := os.Lstat(socket)
	if err != nil || info.Mode()&os.ModeSocket == 0 {
		return nil, errors.New("owned resume socket missing")
	}
	c, err := grpc.NewClient("passthrough:///research-public-resume", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", socket)
		}))
	if err != nil {
		return nil, err
	}
	return &Native{connection: c}, nil
}

func (n *Native) Close() error { return n.connection.Close() }

func (n *Native) Read(ctx context.Context, collection, digest string) ([]retrieval.Candidate, error) {
	if ctx == nil || collection == "" || len(digest) != 64 {
		return nil, errors.New("invalid prefix read")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	response := &wireResponse{}
	if err := n.connection.Invoke(ctx, Method, &wireRequest{Collection: collection, Key: "source_sha256", Value: digest}, response); err != nil {
		return nil, err
	}
	if len(response.Results) != 1 || response.Results[0] == nil {
		return nil, errors.New("prefix source missing or duplicate")
	}
	v := response.Results[0]
	return []retrieval.Candidate{{ID: v.ID, Text: v.Text, Score: v.Score, Metadata: append([]byte(nil), v.MetadataJSON...)}}, nil
}

// ValidatePlan requires the original leading source order, not a convenient
// subset of records found after reopening. It returns no writable registry data.
func ValidatePlan(r *researchpublicpool.Registry, ids []string) error {
	if r == nil || len(ids) == 0 {
		return errors.New("empty resume plan")
	}
	e := r.Entries()
	if len(ids) > len(e) {
		return errors.New("resume prefix too long")
	}
	for i, id := range ids {
		if id != e[i].Candidate.ID {
			return errors.New("not the canonical corpus prefix")
		}
	}
	return nil
}

// Verify must finish completely before the caller inserts new rows. Emitted
// raw reads are audit evidence; a partial trace is not a successful verification.
func Verify(ctx context.Context, r *researchpublicpool.Registry, ids []string, asOf time.Time, reader Reader,
	emit func(int, researchpublicpool.Entry, []retrieval.Candidate, int64) error) error {
	if ctx == nil || reader == nil || emit == nil {
		return errors.New("invalid prefix verification")
	}
	if err := ValidatePlan(r, ids); err != nil {
		return err
	}
	for i, e := range r.Entries()[:len(ids)] {
		if err := ctx.Err(); err != nil {
			return err
		}
		start := time.Now()
		rows, err := reader.Read(ctx, r.Collection(), e.SourceSHA256)
		ns := time.Since(start).Nanoseconds()
		if err != nil {
			return err
		}
		if len(rows) != 1 || rows[0].ID != e.Candidate.ID {
			return errors.New("stored prefix identity changed")
		}
		if _, err = r.Bind(ctx, rows, asOf, 1); err != nil {
			return err
		}
		if err = emit(i, e, rows, ns); err != nil {
			return err
		}
	}
	return ctx.Err()
}

// Field numbers come from the public original plugin's ListByMeta contract,
// matching the existing daemon client. Only this read-only RPC is implemented.
type wireRequest struct {
	Collection string `protobuf:"bytes,1,opt,name=collection,proto3"`
	Key        string `protobuf:"bytes,2,opt,name=key,proto3"`
	Value      string `protobuf:"bytes,3,opt,name=value,proto3"`
}

func (m *wireRequest) Reset()         { *m = wireRequest{} }
func (m *wireRequest) String() string { return oldproto.CompactTextString(m) }
func (*wireRequest) ProtoMessage()    {}

type wireResult struct {
	ID           string  `protobuf:"bytes,1,opt,name=id,proto3"`
	Score        float64 `protobuf:"fixed64,2,opt,name=score,proto3"`
	Text         string  `protobuf:"bytes,3,opt,name=text,proto3"`
	MetadataJSON []byte  `protobuf:"bytes,4,opt,name=metadata_json,json=metadataJson,proto3"`
}

func (m *wireResult) Reset()         { *m = wireResult{} }
func (m *wireResult) String() string { return oldproto.CompactTextString(m) }
func (*wireResult) ProtoMessage()    {}

type wireResponse struct {
	Results []*wireResult `protobuf:"bytes,1,rep,name=results,proto3"`
}

func (m *wireResponse) Reset()         { *m = wireResponse{} }
func (m *wireResponse) String() string { return oldproto.CompactTextString(m) }
func (*wireResponse) ProtoMessage()    {}
