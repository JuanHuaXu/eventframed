package retrieval

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	oldproto "github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	rankCandidatesMethod        = "/libravdb.ipc.v1.LibravDB/RankCandidates"
	searchTextCollectionsMethod = "/libravdb.ipc.v1.LibravDB/SearchTextCollections"
	insertTextMethod            = "/libravdb.ipc.v1.LibravDB/InsertText"
	listByMetaMethod            = "/libravdb.ipc.v1.LibravDB/ListByMeta"
	deleteTextMethod            = "/libravdb.ipc.v1.LibravDB/Delete"
	deleteTextBatchMethod       = "/libravdb.ipc.v1.LibravDB/DeleteBatch"
)

type LibraVDBRanker struct {
	connection *grpc.ClientConn
	guard      *contractGuard
	writeMu    sync.Mutex
}

type ContractClientConfig struct {
	Endpoint         string
	TLSMode          string
	CAFile           string
	ClientCertFile   string
	ClientKeyFile    string
	MaxConcurrent    int
	RequestTimeout   time.Duration
	MaxAttempts      int
	FailureThreshold int
	OpenDuration     time.Duration
}

func OpenLibraVDBContracts(endpoint string) (*LibraVDBRanker, error) {
	return OpenLibraVDBContractsWithConfig(ContractClientConfig{Endpoint: endpoint, TLSMode: "insecure"})
}

func OpenLibraVDBContractsWithConfig(config ContractClientConfig) (*LibraVDBRanker, error) {
	endpoint := strings.TrimSpace(config.Endpoint)
	if endpoint == "" {
		return nil, errors.New("LibraVDB contract endpoint is required")
	}
	target := endpoint
	transport, err := contractTransportCredentials(config, endpoint)
	if err != nil {
		return nil, err
	}
	options := []grpc.DialOption{grpc.WithTransportCredentials(transport)}
	if strings.HasPrefix(endpoint, "unix:") {
		socketPath := strings.TrimPrefix(endpoint, "unix:")
		if socketPath == "" {
			return nil, errors.New("LibraVDB unix endpoint has no socket path")
		}
		target = "passthrough:///libravdb-contracts"
		options = append(options, grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
		}))
	} else {
		target = strings.TrimPrefix(endpoint, "tcp:")
	}
	connection, err := grpc.NewClient(target, options...)
	if err != nil {
		return nil, fmt.Errorf("connect LibraVDB contracts: %w", err)
	}
	return &LibraVDBRanker{connection: connection, guard: newContractGuard(config)}, nil
}

var ErrContractCircuitOpen = errors.New("LibraVDB contract circuit is open")

type contractGuard struct {
	permits             chan struct{}
	timeout             time.Duration
	attempts            int
	failureThreshold    int
	openDuration        time.Duration
	mu                  sync.Mutex
	consecutiveFailures int
	openUntil           time.Time
}

func newContractGuard(config ContractClientConfig) *contractGuard {
	concurrency := config.MaxConcurrent
	if concurrency <= 0 {
		concurrency = 16
	}
	timeout := config.RequestTimeout
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	attempts := config.MaxAttempts
	if attempts <= 0 {
		attempts = 2
	}
	threshold := config.FailureThreshold
	if threshold <= 0 {
		threshold = 5
	}
	openDuration := config.OpenDuration
	if openDuration <= 0 {
		openDuration = 5 * time.Second
	}
	return &contractGuard{permits: make(chan struct{}, concurrency), timeout: timeout, attempts: attempts, failureThreshold: threshold, openDuration: openDuration}
}

func (r *LibraVDBRanker) invoke(ctx context.Context, method string, request, response oldproto.Message, write bool) error {
	if write {
		r.writeMu.Lock()
		defer r.writeMu.Unlock()
	}
	if err := r.guard.acquire(ctx); err != nil {
		return err
	}
	defer r.guard.release()
	var err error
	for attempt := 0; attempt < r.guard.attempts; attempt++ {
		if attempt > 0 {
			delay := time.Duration(25*(1<<(attempt-1))) * time.Millisecond
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
		attemptCtx, cancel := context.WithTimeout(ctx, r.guard.timeout)
		err = r.connection.Invoke(attemptCtx, method, request, response)
		cancel()
		if err == nil {
			r.guard.success()
			return nil
		}
		if !retryableContractError(err) {
			return err
		}
		r.guard.failure()
		if r.guard.isOpen() {
			return fmt.Errorf("%w: %v", ErrContractCircuitOpen, err)
		}
	}
	return err
}

func (g *contractGuard) acquire(ctx context.Context) error {
	if g.isOpen() {
		return ErrContractCircuitOpen
	}
	select {
	case g.permits <- struct{}{}:
		if g.isOpen() {
			<-g.permits
			return ErrContractCircuitOpen
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (g *contractGuard) release() { <-g.permits }
func (g *contractGuard) isOpen() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.openUntil.IsZero() {
		return false
	}
	if time.Now().Before(g.openUntil) {
		return true
	}
	g.openUntil = time.Time{}
	g.consecutiveFailures = 0
	return false
}
func (g *contractGuard) failure() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.consecutiveFailures++
	if g.consecutiveFailures >= g.failureThreshold {
		g.openUntil = time.Now().Add(g.openDuration)
	}
}
func (g *contractGuard) success() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.consecutiveFailures = 0
	g.openUntil = time.Time{}
}

func retryableContractError(err error) bool {
	switch status.Code(err) {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted:
		return true
	default:
		return false
	}
}

func (r *LibraVDBRanker) Ready(ctx context.Context) error {
	if r.guard.isOpen() {
		return ErrContractCircuitOpen
	}
	state := r.connection.GetState()
	if state == connectivity.Ready {
		return nil
	}
	if state == connectivity.Shutdown {
		return errors.New("LibraVDB contract connection is shut down")
	}
	r.connection.Connect()
	for {
		if !r.connection.WaitForStateChange(ctx, state) {
			return ctx.Err()
		}
		state = r.connection.GetState()
		if state == connectivity.Ready {
			return nil
		}
		if state == connectivity.TransientFailure || state == connectivity.Shutdown {
			return fmt.Errorf("LibraVDB contract connection is %s", state)
		}
	}
}

func contractTransportCredentials(config ContractClientConfig, endpoint string) (credentials.TransportCredentials, error) {
	mode := strings.TrimSpace(config.TLSMode)
	if mode == "" {
		mode = "auto"
	}
	useTLS := mode == "tls" || (mode == "auto" && !localContractEndpoint(endpoint))
	if mode != "auto" && mode != "tls" && mode != "insecure" {
		return nil, fmt.Errorf("unsupported LibraVDB TLS mode %q", mode)
	}
	if !useTLS {
		if config.CAFile != "" || config.ClientCertFile != "" || config.ClientKeyFile != "" {
			return nil, errors.New("LibraVDB TLS files require a TLS transport")
		}
		return insecure.NewCredentials(), nil
	}
	if (config.ClientCertFile == "") != (config.ClientKeyFile == "") {
		return nil, errors.New("LibraVDB TLS client certificate and key must be configured together")
	}
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if config.CAFile != "" {
		pem, err := os.ReadFile(config.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read LibraVDB TLS CA: %w", err)
		}
		roots := x509.NewCertPool()
		if !roots.AppendCertsFromPEM(pem) {
			return nil, errors.New("LibraVDB TLS CA contains no certificates")
		}
		tlsConfig.RootCAs = roots
	}
	if config.ClientCertFile != "" {
		certificate, err := tls.LoadX509KeyPair(config.ClientCertFile, config.ClientKeyFile)
		if err != nil {
			return nil, fmt.Errorf("load LibraVDB mTLS client identity: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{certificate}
	}
	return credentials.NewTLS(tlsConfig), nil
}

func localContractEndpoint(endpoint string) bool {
	if strings.HasPrefix(endpoint, "unix:") {
		return true
	}
	host, _, err := net.SplitHostPort(strings.TrimPrefix(endpoint, "tcp:"))
	if err != nil {
		return false
	}
	return host == "localhost" || net.ParseIP(host).IsLoopback()
}

// OpenLibraVDBRanker is retained for evaluation callers that only use the
// ranking subset of the contract client.
func OpenLibraVDBRanker(endpoint string) (*LibraVDBRanker, error) {
	return OpenLibraVDBContracts(endpoint)
}

func (r *LibraVDBRanker) Close() error { return r.connection.Close() }

func (r *LibraVDBRanker) ContractName() string { return "libravdb.ipc.v1.LibravDB/RankCandidates" }

func (r *LibraVDBRanker) RetrievalContractName() string {
	return "libravdb.ipc.v1.LibravDB/SearchTextCollections"
}

func (r *LibraVDBRanker) SearchTextCollections(ctx context.Context, request SearchRequest) ([]Candidate, error) {
	if len(request.Collections) == 0 || strings.TrimSpace(request.QueryText) == "" || request.K <= 0 {
		return nil, errors.New("SearchTextCollections requires collections, query text, and k > 0")
	}
	wireRequest := &searchTextCollectionsRequest{
		Collections:         request.Collections,
		Text:                request.QueryText,
		K:                   int32(request.K),
		ExcludeByCollection: make(map[string]*stringList, len(request.ExcludeByCollection)),
	}
	for collection, ids := range request.ExcludeByCollection {
		wireRequest.ExcludeByCollection[collection] = &stringList{Values: ids}
	}
	wireResponse := &searchTextResponse{}
	if err := r.invoke(ctx, searchTextCollectionsMethod, wireRequest, wireResponse, false); err != nil {
		return nil, fmt.Errorf("LibraVDB SearchTextCollections: %w", err)
	}
	results := make([]Candidate, 0, len(wireResponse.Results))
	for _, result := range wireResponse.Results {
		if result == nil || result.ID == "" {
			return nil, errors.New("LibraVDB SearchTextCollections returned an invalid candidate")
		}
		results = append(results, Candidate{ID: result.ID, Text: result.Text, Score: result.Score, Metadata: result.MetadataJSON})
	}
	return results, nil
}

func (r *LibraVDBRanker) InsertText(ctx context.Context, collection string, candidate Candidate) error {
	if strings.TrimSpace(collection) == "" || candidate.ID == "" || strings.TrimSpace(candidate.Text) == "" {
		return errors.New("InsertText requires collection, candidate ID, and text")
	}
	wireResponse := &insertTextResponse{}
	if err := r.invoke(ctx, insertTextMethod, &insertTextRequest{
		Collection:   collection,
		ID:           candidate.ID,
		Text:         candidate.Text,
		MetadataJSON: candidate.Metadata,
	}, wireResponse, true); err != nil {
		return fmt.Errorf("LibraVDB InsertText: %w", err)
	}
	if !wireResponse.OK {
		return errors.New("LibraVDB InsertText returned ok=false")
	}
	return nil
}

func (r *LibraVDBRanker) EnsureText(ctx context.Context, collection string, candidate Candidate, identityKey, identityValue string) error {
	err := r.InsertText(ctx, collection, candidate)
	if err == nil {
		return nil
	}
	if identityKey == "" || identityValue == "" {
		return err
	}
	wireResponse := &listByMetaResponse{}
	if lookupErr := r.invoke(ctx, listByMetaMethod, &listByMetaRequest{
		Collection: collection, Key: identityKey, Value: identityValue,
	}, wireResponse, false); lookupErr != nil {
		return fmt.Errorf("%v; LibraVDB ListByMeta reconciliation: %w", err, lookupErr)
	}
	for _, result := range wireResponse.Results {
		if result != nil && result.ID == candidate.ID && result.Text == candidate.Text {
			return nil
		}
	}
	return err
}

func (r *LibraVDBRanker) DeleteText(ctx context.Context, collection, id string) error {
	if strings.TrimSpace(collection) == "" || strings.TrimSpace(id) == "" {
		return errors.New("Delete requires collection and ID")
	}
	wireResponse := &deleteTextResponse{}
	if err := r.invoke(ctx, deleteTextMethod, &deleteTextRequest{Collection: collection, ID: id}, wireResponse, true); err != nil {
		return fmt.Errorf("LibraVDB Delete: %w", err)
	}
	if !wireResponse.OK {
		return errors.New("LibraVDB Delete returned ok=false")
	}
	return nil
}

func (r *LibraVDBRanker) DeleteTextBatch(ctx context.Context, collection string, ids []string) error {
	if strings.TrimSpace(collection) == "" || len(ids) == 0 {
		return errors.New("DeleteBatch requires collection and IDs")
	}
	wireResponse := &deleteTextBatchResponse{}
	if err := r.invoke(ctx, deleteTextBatchMethod, &deleteTextBatchRequest{Collection: collection, IDs: ids}, wireResponse, true); err != nil {
		return fmt.Errorf("LibraVDB DeleteBatch: %w", err)
	}
	if !wireResponse.OK {
		return errors.New("LibraVDB DeleteBatch returned ok=false")
	}
	return nil
}

func (r *LibraVDBRanker) RankCandidates(ctx context.Context, request RankRequest) ([]Candidate, error) {
	if request.K1 <= 0 || request.K2 <= 0 || request.K2 > request.K1 {
		return nil, errors.New("RankCandidates requires 0 < k2 <= k1")
	}
	wireRequest := &rankCandidatesRequest{
		Candidates: make([]*rankCandidate, 0, len(request.Candidates)), QueryText: request.QueryText,
		SessionID: request.SessionID, UserID: request.UserID, K1: int32(request.K1), K2: int32(request.K2),
	}
	for _, candidate := range request.Candidates {
		wireRequest.Candidates = append(wireRequest.Candidates, &rankCandidate{
			ID: candidate.ID, Text: candidate.Text, Score: candidate.Score, MetadataJSON: candidate.Metadata,
		})
	}
	wireResponse := &rankCandidatesResponse{}
	if err := r.invoke(ctx, rankCandidatesMethod, wireRequest, wireResponse, false); err != nil {
		return nil, fmt.Errorf("LibraVDB RankCandidates: %w", err)
	}
	ranked := make([]Candidate, 0, len(wireResponse.Ranked))
	for _, candidate := range wireResponse.Ranked {
		if candidate == nil || candidate.ID == "" {
			return nil, errors.New("LibraVDB RankCandidates returned an invalid candidate")
		}
		ranked = append(ranked, Candidate{ID: candidate.ID, Text: candidate.Text, Score: candidate.Score, Metadata: candidate.MetadataJSON})
	}
	return ranked, nil
}

// These compact wire types intentionally cover only the frozen public
// RankCandidates messages from @xdarkicex/libravdb-contracts. The RPC name and
// field numbers are the compatibility boundary; no daemon ranking logic is
// reproduced in EventFrame.
type rankCandidate struct {
	ID           string  `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Text         string  `protobuf:"bytes,2,opt,name=text,proto3" json:"text,omitempty"`
	Score        float64 `protobuf:"fixed64,3,opt,name=score,proto3" json:"score,omitempty"`
	MetadataJSON []byte  `protobuf:"bytes,4,opt,name=metadata_json,json=metadataJson,proto3" json:"metadata_json,omitempty"`
}

func (m *rankCandidate) Reset()         { *m = rankCandidate{} }
func (m *rankCandidate) String() string { return oldproto.CompactTextString(m) }
func (*rankCandidate) ProtoMessage()    {}

type rankCandidatesRequest struct {
	Candidates []*rankCandidate `protobuf:"bytes,1,rep,name=candidates,proto3" json:"candidates,omitempty"`
	QueryText  string           `protobuf:"bytes,2,opt,name=query_text,json=queryText,proto3" json:"query_text,omitempty"`
	SessionID  string           `protobuf:"bytes,3,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	UserID     string           `protobuf:"bytes,4,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	K1         int32            `protobuf:"varint,5,opt,name=k1,proto3" json:"k1,omitempty"`
	K2         int32            `protobuf:"varint,6,opt,name=k2,proto3" json:"k2,omitempty"`
}

func (m *rankCandidatesRequest) Reset()         { *m = rankCandidatesRequest{} }
func (m *rankCandidatesRequest) String() string { return oldproto.CompactTextString(m) }
func (*rankCandidatesRequest) ProtoMessage()    {}

type rankCandidatesResponse struct {
	Ranked []*rankCandidate `protobuf:"bytes,1,rep,name=ranked,proto3" json:"ranked,omitempty"`
}

func (m *rankCandidatesResponse) Reset()         { *m = rankCandidatesResponse{} }
func (m *rankCandidatesResponse) String() string { return oldproto.CompactTextString(m) }
func (*rankCandidatesResponse) ProtoMessage()    {}

type stringList struct {
	Values []string `protobuf:"bytes,1,rep,name=values,proto3" json:"values,omitempty"`
}

func (m *stringList) Reset()         { *m = stringList{} }
func (m *stringList) String() string { return oldproto.CompactTextString(m) }
func (*stringList) ProtoMessage()    {}

type searchTextCollectionsRequest struct {
	Collections         []string               `protobuf:"bytes,1,rep,name=collections,proto3" json:"collections,omitempty"`
	Text                string                 `protobuf:"bytes,2,opt,name=text,proto3" json:"text,omitempty"`
	K                   int32                  `protobuf:"varint,3,opt,name=k,proto3" json:"k,omitempty"`
	ExcludeByCollection map[string]*stringList `protobuf:"bytes,4,rep,name=exclude_by_collection,json=excludeByCollection,proto3" json:"exclude_by_collection,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (m *searchTextCollectionsRequest) Reset()         { *m = searchTextCollectionsRequest{} }
func (m *searchTextCollectionsRequest) String() string { return oldproto.CompactTextString(m) }
func (*searchTextCollectionsRequest) ProtoMessage()    {}

type searchResult struct {
	ID           string  `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Score        float64 `protobuf:"fixed64,2,opt,name=score,proto3" json:"score,omitempty"`
	Text         string  `protobuf:"bytes,3,opt,name=text,proto3" json:"text,omitempty"`
	MetadataJSON []byte  `protobuf:"bytes,4,opt,name=metadata_json,json=metadataJson,proto3" json:"metadata_json,omitempty"`
	Version      uint64  `protobuf:"varint,5,opt,name=version,proto3" json:"version,omitempty"`
}

func (m *searchResult) Reset()         { *m = searchResult{} }
func (m *searchResult) String() string { return oldproto.CompactTextString(m) }
func (*searchResult) ProtoMessage()    {}

type searchTextResponse struct {
	Results []*searchResult `protobuf:"bytes,1,rep,name=results,proto3" json:"results,omitempty"`
}

func (m *searchTextResponse) Reset()         { *m = searchTextResponse{} }
func (m *searchTextResponse) String() string { return oldproto.CompactTextString(m) }
func (*searchTextResponse) ProtoMessage()    {}

type insertTextRequest struct {
	Collection   string `protobuf:"bytes,1,opt,name=collection,proto3" json:"collection,omitempty"`
	ID           string `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Text         string `protobuf:"bytes,3,opt,name=text,proto3" json:"text,omitempty"`
	MetadataJSON []byte `protobuf:"bytes,4,opt,name=metadata_json,json=metadataJson,proto3" json:"metadata_json,omitempty"`
}

func (m *insertTextRequest) Reset()         { *m = insertTextRequest{} }
func (m *insertTextRequest) String() string { return oldproto.CompactTextString(m) }
func (*insertTextRequest) ProtoMessage()    {}

type insertTextResponse struct {
	OK bool `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
}

func (m *insertTextResponse) Reset()         { *m = insertTextResponse{} }
func (m *insertTextResponse) String() string { return oldproto.CompactTextString(m) }
func (*insertTextResponse) ProtoMessage()    {}

type listByMetaRequest struct {
	Collection string `protobuf:"bytes,1,opt,name=collection,proto3" json:"collection,omitempty"`
	Key        string `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
	Value      string `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *listByMetaRequest) Reset()         { *m = listByMetaRequest{} }
func (m *listByMetaRequest) String() string { return oldproto.CompactTextString(m) }
func (*listByMetaRequest) ProtoMessage()    {}

type listByMetaResponse struct {
	Results []*searchResult `protobuf:"bytes,1,rep,name=results,proto3" json:"results,omitempty"`
}

func (m *listByMetaResponse) Reset()         { *m = listByMetaResponse{} }
func (m *listByMetaResponse) String() string { return oldproto.CompactTextString(m) }
func (*listByMetaResponse) ProtoMessage()    {}

type deleteTextRequest struct {
	Collection string `protobuf:"bytes,1,opt,name=collection,proto3" json:"collection,omitempty"`
	ID         string `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
}

func (m *deleteTextRequest) Reset()         { *m = deleteTextRequest{} }
func (m *deleteTextRequest) String() string { return oldproto.CompactTextString(m) }
func (*deleteTextRequest) ProtoMessage()    {}

type deleteTextResponse struct {
	OK bool `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
}

func (m *deleteTextResponse) Reset()         { *m = deleteTextResponse{} }
func (m *deleteTextResponse) String() string { return oldproto.CompactTextString(m) }
func (*deleteTextResponse) ProtoMessage()    {}

type deleteTextBatchRequest struct {
	Collection string   `protobuf:"bytes,1,opt,name=collection,proto3" json:"collection,omitempty"`
	IDs        []string `protobuf:"bytes,2,rep,name=ids,proto3" json:"ids,omitempty"`
}

func (m *deleteTextBatchRequest) Reset()         { *m = deleteTextBatchRequest{} }
func (m *deleteTextBatchRequest) String() string { return oldproto.CompactTextString(m) }
func (*deleteTextBatchRequest) ProtoMessage()    {}

type deleteTextBatchResponse struct {
	OK bool `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
}

func (m *deleteTextBatchResponse) Reset()         { *m = deleteTextBatchResponse{} }
func (m *deleteTextBatchResponse) String() string { return oldproto.CompactTextString(m) }
func (*deleteTextBatchResponse) ProtoMessage()    {}
