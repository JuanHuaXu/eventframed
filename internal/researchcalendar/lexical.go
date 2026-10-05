package researchcalendar

import (
	"context"
	"errors"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

type LexicalResult struct {
	Order  []int     `json:"order"`
	Scores []float64 `json:"scores"`
	Method string    `json:"method"`
}

var lexicalDates = regexp.MustCompile(`(?i)\b\d{4}-\d{2}-\d{2}\b|\b\d{1,2}\s+(?:January|February|March|April|May|June|July|August|September|October|November|December)\s+\d{4}\b`)
var lexicalTokens = regexp.MustCompile(`[a-z]+|[0-9]+(?:\.[0-9]+)*`)
var lexicalStops = func() map[string]bool {
	m := map[string]bool{}
	for _, s := range strings.Fields("a an the on in at to of for from with by and is was were did do does it that this which what when why how before after not request outcome recorded") {
		m[s] = true
	}
	return m
}()

func lexicalTerms(s string) []string {
	s = strings.ToLower(lexicalDates.ReplaceAllString(s, " "))
	seen := map[string]bool{}
	out := []string{}
	for _, w := range lexicalTokens.FindAllString(s, -1) {
		if lexicalStops[w] || (!(w[0] >= '0' && w[0] <= '9') && len(w) < 3) {
			continue
		}
		switch {
		case len(w) > 5 && strings.HasSuffix(w, "ing"):
			w = w[:len(w)-3]
		case len(w) > 4 && strings.HasSuffix(w, "ed"):
			w = w[:len(w)-2]
		case len(w) > 4 && strings.HasSuffix(w, "s"):
			w = w[:len(w)-1]
		}
		if !seen[w] {
			out = append(out, w)
			seen[w] = true
		}
	}
	return out
}

// LexicalOrder ports the frozen what-lexical-v2 comparator. Ordered term slices
// keep reductions deterministic; maps are used only for membership/counts.
// Scores are search features, never properly scored probabilities. No inputs
// are mutated and this helper does not implement packet selection.
func LexicalOrder(ctx context.Context, query string, candidates []retrieval.Candidate, plan *PriorityPlan) (LexicalResult, error) {
	if err := ctx.Err(); err != nil {
		return LexicalResult{}, err
	}
	n := len(candidates)
	if len(query) > 4096 || n == 0 || n > 200 {
		return LexicalResult{}, errors.New("bounded lexical input required")
	}
	seen := map[string]bool{}
	df := map[string]int{}
	docs := make([][]string, n)
	for i, c := range candidates {
		if err := ctx.Err(); err != nil {
			return LexicalResult{}, err
		}
		if c.ID == "" || seen[c.ID] || len(c.Text) > 8192 || math.IsNaN(c.Score) || math.IsInf(c.Score, 0) {
			return LexicalResult{}, errors.New("invalid lexical candidate")
		}
		seen[c.ID] = true
		text, ok := what(c.Text)
		if !ok {
			return LexicalResult{}, errors.New("one canonical what field required")
		}
		docs[i] = lexicalTerms(text)
		for _, w := range docs[i] {
			df[w]++
		}
	}
	q := lexicalTerms(query)
	qm := map[string]bool{}
	weight := func(w string) float64 { return 1 + math.Log(float64(n+1)/float64(df[w]+1)) }
	qnorm := 0.0
	for _, w := range q {
		qm[w] = true
		v := weight(w)
		qnorm += v * v
	}
	qnorm = math.Sqrt(qnorm)
	out := LexicalResult{Order: make([]int, n), Scores: make([]float64, n), Method: "what-lexical-v2"}
	for i, d := range docs {
		norm, dot := 0.0, 0.0
		for _, w := range d {
			v := weight(w)
			norm += v * v
			if qm[w] {
				dot += v * v
			}
		}
		norm = math.Sqrt(norm)
		if norm > 0 && qnorm > 0 {
			out.Scores[i] = dot / (norm * qnorm)
		}
		out.Order[i] = i
	}
	states := map[string]string{}
	if plan != nil {
		for _, d := range plan.Decisions {
			states[d.EventID] = d.State
		}
		if len(states) != n {
			return LexicalResult{}, errors.New("plan identity mismatch")
		}
		for _, c := range candidates {
			if _, ok := states[c.ID]; !ok {
				return LexicalResult{}, errors.New("plan identity mismatch")
			}
		}
	}
	group := func(i int) int {
		if plan != nil && !plan.AllContradicted && states[candidates[i].ID] == "contradicted" {
			return 1
		}
		return 0
	}
	sort.SliceStable(out.Order, func(i, j int) bool {
		a, b := out.Order[i], out.Order[j]
		if group(a) != group(b) {
			return group(a) < group(b)
		}
		if out.Scores[a] != out.Scores[b] {
			return out.Scores[a] > out.Scores[b]
		}
		if candidates[a].Score != candidates[b].Score {
			return candidates[a].Score > candidates[b].Score
		}
		return a < b
	})
	if err := ctx.Err(); err != nil {
		return LexicalResult{}, err
	}
	return out, nil
}
