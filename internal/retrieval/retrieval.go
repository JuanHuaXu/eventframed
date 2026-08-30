package retrieval

import "context"

// Candidate mirrors the fields consumed and returned by LibraVDB's
// RankCandidates contract. Score is LibraVDB's upstream/base ranking signal;
// EventFrame adjustments are applied after this contract returns. Metadata
// remains owned by LibraVDB.
type Candidate struct {
	ID       string
	Text     string
	Score    float64
	Metadata []byte
}

type RankRequest struct {
	Candidates []Candidate
	QueryText  string
	SessionID  string
	UserID     string
	K1         int
	K2         int
}

type SearchRequest struct {
	Collections         []string
	QueryText           string
	K                   int
	ExcludeByCollection map[string][]string
}

type CandidateRetriever interface {
	SearchTextCollections(context.Context, SearchRequest) ([]Candidate, error)
	RetrievalContractName() string
}

// TextIndexer exists for isolated contract tests and offline evaluation data
// preparation. Production uses CandidateIndex through eventframed.
type TextIndexer interface {
	InsertText(context.Context, string, Candidate) error
}

// CandidateIndex owns the production lifecycle of contract-visible candidate
// records. EnsureText must be idempotent for the supplied identity pair.
type CandidateIndex interface {
	EnsureText(context.Context, string, Candidate, string, string) error
	DeleteText(context.Context, string, string) error
	DeleteTextBatch(context.Context, string, []string) error
}

type CandidateRanker interface {
	RankCandidates(context.Context, RankRequest) ([]Candidate, error)
	ContractName() string
}

type PassthroughRanker struct{}

func (PassthroughRanker) RankCandidates(_ context.Context, request RankRequest) ([]Candidate, error) {
	limit := min(request.K2, len(request.Candidates))
	return append([]Candidate(nil), request.Candidates[:limit]...), nil
}

func (PassthroughRanker) ContractName() string { return "embedded-search-order" }
