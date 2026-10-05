package researchcalendar

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

// MarginRanker reserves a cross-partition gap against a declared later additive
// rank-delta bound. It is research-only: caller configuration must honor MaxDelta.
// It does not protect against arbitrary later rerankers, diversity or resolution.
type MarginRanker struct{ MaxDelta float64 }

func (m MarginRanker) ContractName() string {
	return fmt.Sprintf("research/calendar-margin-v1/%.17g", m.MaxDelta)
}

func (m MarginRanker) RankCandidates(ctx context.Context, req retrieval.RankRequest) ([]retrieval.Candidate, error) {
	if math.IsNaN(m.MaxDelta) || math.IsInf(m.MaxDelta, 0) || m.MaxDelta < 0 || m.MaxDelta > .49 || req.K1 != req.K2 {
		return nil, errors.New("calendar margin requires finite bound <=.49 and complete frontier")
	}
	out, err := (Ranker{}).RankCandidates(ctx, req)
	if err != nil {
		return nil, err
	}
	b := ctx.Value(bindingKey{}).(binding)
	rs := rules(b.original)
	bad := make([]bool, len(out))
	count := 0
	for i, c := range out {
		bad[i] = contradict(rs, c.Text)
		if bad[i] {
			count++
		}
	}
	if count == 0 || count == len(out) {
		return out, nil
	}
	a := (1 - 2*m.MaxDelta) / 2
	for i := range out {
		// Scores are strictly inside each band. Cross-band gap >2*MaxDelta;
		// adding bounded opposite corrections cannot reverse the partition.
		ordinal := float64(len(out)-i) / float64(len(out)+1)
		out[i].Score = a * ordinal
		if !bad[i] {
			out[i].Score += 1 - a
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
