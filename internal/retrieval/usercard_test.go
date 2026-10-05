package retrieval

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	oldproto "github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestUserCardWireContractAndTenantScope(t *testing.T) {
	listener := bufconn.Listen(1 << 20)
	var upserts atomic.Int32
	server := grpc.NewServer(grpc.UnknownServiceHandler(func(_ any, stream grpc.ServerStream) error {
		md, _ := metadata.FromIncomingContext(stream.Context())
		if keys := md.Get("x-libravdb-tenant-key"); len(keys) != 1 || keys[0] != "tenant-a" {
			return status.Error(codes.PermissionDenied, "wrong tenant scope")
		}
		method, _ := grpc.MethodFromServerStream(stream)
		switch method {
		case "/libravdb.ipc.v1.LibravDB/GetUserCard":
			var request GetUserCardRequest
			if err := stream.RecvMsg(&request); err != nil {
				return err
			}
			if request.UserID != "account:a" {
				return status.Error(codes.InvalidArgument, "wrong key")
			}
			return stream.SendMsg(&GetUserCardResponse{CardJSON: `{"name":"Example Operator"}`, UpdatedAt: 1720000000000, Version: 3})
		case "/libravdb.ipc.v1.LibravDB/UpsertUserCard":
			var request UpsertUserCardRequest
			if err := stream.RecvMsg(&request); err != nil {
				return err
			}
			upserts.Add(1)
			if request.CardJSON == `{"retry":true}` {
				return status.Error(codes.Unavailable, "ambiguous committed write")
			}
			return stream.SendMsg(&UpsertUserCardResponse{OK: true, CardID: "__user_card__", PreviousHash: "previous"})
		}
		return status.Error(codes.Unimplemented, "unknown method")
	}))
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	connection, err := grpc.NewClient("passthrough:///test", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	client := &LibraVDBRanker{connection: connection, guard: newContractGuard(ContractClientConfig{RequestTimeout: time.Second, MaxAttempts: 3})}
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("x-libravdb-tenant-key", "tenant-other"))
	card, err := client.GetUserCard(ctx, "tenant-a", GetUserCardRequest{UserID: "account:a"})
	if err != nil || card.Version != 3 || card.UpdatedAt != 1720000000000 {
		t.Fatalf("get = %+v, %v", card, err)
	}
	result, err := client.UpsertUserCard(ctx, "tenant-a", UpsertUserCardRequest{UserID: "account:a", CardJSON: `{}`})
	if err != nil || !result.OK || result.CardID != "__user_card__" || result.PreviousHash != "previous" {
		t.Fatalf("upsert = %+v, %v", result, err)
	}
	_, err = client.UpsertUserCard(ctx, "tenant-a", UpsertUserCardRequest{UserID: "account:a", CardJSON: `{"retry":true}`})
	if status.Code(err) != codes.Unavailable || upserts.Load() != 2 {
		t.Fatalf("non-idempotent write retried: %d, %v", upserts.Load(), err)
	}
	if _, err := client.GetUserCard(ctx, "", GetUserCardRequest{UserID: "account:a"}); err == nil {
		t.Fatal("missing tenant accepted")
	}
	if _, err := client.UpsertUserCard(ctx, "tenant-a", UpsertUserCardRequest{UserID: "account:a", CardJSON: `null`}); err == nil {
		t.Fatal("non-object card accepted")
	}
}

func TestUserCardFieldNumbersAgainstPublicContract(t *testing.T) {
	// Independent wire literals: field 1 user_id, field 2 card_json. This catches
	// a client/server fake agreeing with each other on the wrong field numbers.
	request := &UpsertUserCardRequest{UserID: "u", CardJSON: "{}"}
	encoded, err := oldproto.Marshal(request)
	if err != nil || string(encoded) != string([]byte{0x0a, 1, 'u', 0x12, 2, '{', '}'}) {
		t.Fatalf("request bytes %x, %v", encoded, err)
	}
	var response GetUserCardResponse
	if err := oldproto.Unmarshal([]byte{0x0a, 2, '{', '}', 0x10, 7, 0x18, 3}, &response); err != nil || response.CardJSON != "{}" || response.UpdatedAt != 7 || response.Version != 3 {
		t.Fatalf("response %+v, %v", response, err)
	}
}
