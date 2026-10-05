package retrieval

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"unicode"

	oldproto "github.com/golang/protobuf/proto"
	"google.golang.org/grpc/metadata"
)

// Field numbers and RPC names mirror libravdb-client.ts's public IPC contracts,
// not the native embedded database API. The tenant namespace is RPC metadata.
type UpsertUserCardRequest struct {
	UserID   string `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	CardJSON string `protobuf:"bytes,2,opt,name=card_json,json=cardJson,proto3" json:"card_json,omitempty"`
}
type UpsertUserCardResponse struct {
	OK           bool   `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
	CardID       string `protobuf:"bytes,2,opt,name=card_id,json=cardId,proto3" json:"card_id,omitempty"`
	PreviousHash string `protobuf:"bytes,3,opt,name=previous_hash,json=previousHash,proto3" json:"previous_hash,omitempty"`
}
type GetUserCardRequest struct {
	UserID string `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
}
type GetUserCardResponse struct {
	CardJSON  string `protobuf:"bytes,1,opt,name=card_json,json=cardJson,proto3" json:"card_json,omitempty"`
	UpdatedAt int64  `protobuf:"varint,2,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	Version   int32  `protobuf:"varint,3,opt,name=version,proto3" json:"version,omitempty"`
}

func (m *UpsertUserCardRequest) Reset()          { *m = UpsertUserCardRequest{} }
func (m *UpsertUserCardRequest) String() string  { return oldproto.CompactTextString(m) }
func (*UpsertUserCardRequest) ProtoMessage()     {}
func (m *UpsertUserCardResponse) Reset()         { *m = UpsertUserCardResponse{} }
func (m *UpsertUserCardResponse) String() string { return oldproto.CompactTextString(m) }
func (*UpsertUserCardResponse) ProtoMessage()    {}
func (m *GetUserCardRequest) Reset()             { *m = GetUserCardRequest{} }
func (m *GetUserCardRequest) String() string     { return oldproto.CompactTextString(m) }
func (*GetUserCardRequest) ProtoMessage()        {}
func (m *GetUserCardResponse) Reset()            { *m = GetUserCardResponse{} }
func (m *GetUserCardResponse) String() string    { return oldproto.CompactTextString(m) }
func (*GetUserCardResponse) ProtoMessage()       {}

type UserCardReader interface {
	GetUserCard(context.Context, string, GetUserCardRequest) (GetUserCardResponse, error)
}

func userCardContext(ctx context.Context, tenantKey, userID string) (context.Context, error) {
	if strings.TrimSpace(tenantKey) == "" || strings.TrimSpace(userID) == "" || len(tenantKey) > 1024 || len(userID) > 1024 || strings.ContainsAny(tenantKey, "\r\n") {
		return nil, errors.New("user-card RPC requires bounded tenant key and user ID")
	}
	if strings.IndexFunc(userID, unicode.IsControl) >= 0 || strings.IndexFunc(tenantKey, func(r rune) bool { return r < 32 || r > 126 }) >= 0 {
		return nil, errors.New("invalid user-card identity or non-ASCII tenant metadata")
	}
	md, _ := metadata.FromOutgoingContext(ctx)
	md = md.Copy()
	md.Set("x-libravdb-tenant-key", tenantKey)
	return metadata.NewOutgoingContext(ctx, md), nil
}

func (r *LibraVDBRanker) GetUserCard(ctx context.Context, tenantKey string, request GetUserCardRequest) (GetUserCardResponse, error) {
	ctx, err := userCardContext(ctx, tenantKey, request.UserID)
	if err != nil {
		return GetUserCardResponse{}, err
	}
	var response GetUserCardResponse
	err = r.invoke(ctx, "/libravdb.ipc.v1.LibravDB/GetUserCard", &request, &response, false)
	return response, err
}

// Upsert replaces the card; it is deliberately not called by text ingestion.
// There is no CAS or idempotency token in this contract, so a timed-out write
// must not be retried automatically (the server increments the card version).
func (r *LibraVDBRanker) UpsertUserCard(ctx context.Context, tenantKey string, request UpsertUserCardRequest) (UpsertUserCardResponse, error) {
	ctx, err := userCardContext(ctx, tenantKey, request.UserID)
	if err != nil {
		return UpsertUserCardResponse{}, err
	}
	var object map[string]json.RawMessage
	if len(request.CardJSON) > 256<<10 || json.Unmarshal([]byte(request.CardJSON), &object) != nil || object == nil {
		return UpsertUserCardResponse{}, errors.New("card_json must be a bounded JSON object")
	}
	var response UpsertUserCardResponse
	err = r.invokeWithAttempts(ctx, "/libravdb.ipc.v1.LibravDB/UpsertUserCard", &request, &response, true, 1)
	if err == nil && !response.OK {
		err = errors.New("UpsertUserCard returned ok=false")
	}
	return response, err
}
