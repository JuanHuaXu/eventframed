// Package researchindex contains unwired index-publication research components.
package researchindex

import (
	"errors"
	"math"
	"sort"
)

type Candidate struct {
	ID    string
	Score float64
}
type Change struct {
	Candidate
	Deleted bool
}

// Merge combines a base generation's k+len(delta) best candidates with the
// complete bounded delta. Changes shadow base IDs even when their score falls
// or they are tombstones. Scores must use the same query/metric/normalization.
// Exact global top-k follows only if the base prefix is exact; an ANN prefix
// retains ANN approximation. This function cannot certify an unseen prefix.
func Merge(base []Candidate, delta []Change, k int) ([]Candidate, error) {
	if k < 1 || k > 200 || len(delta) > 128 || len(base) > k+len(delta) {
		return nil, errors.New("research merge bounds exceeded")
	}
	valid := func(c Candidate) bool { return c.ID != "" && !math.IsNaN(c.Score) && !math.IsInf(c.Score, 0) }
	shadow := make(map[string]bool, len(delta))
	all := make([]Candidate, 0, len(base)+len(delta))
	for _, d := range delta {
		if !valid(d.Candidate) || shadow[d.ID] {
			return nil, errors.New("invalid or duplicate delta ID")
		}
		shadow[d.ID] = true
		if !d.Deleted {
			all = append(all, d.Candidate)
		}
	}
	seen := make(map[string]bool, len(base))
	for _, b := range base {
		if !valid(b) || seen[b.ID] {
			return nil, errors.New("invalid or duplicate base ID")
		}
		seen[b.ID] = true
		if !shadow[b.ID] {
			all = append(all, b)
		}
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Score == all[j].Score {
			return all[i].ID < all[j].ID
		}
		return all[i].Score > all[j].Score
	})
	return all[:min(k, len(all))], nil
}
