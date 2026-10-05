package researchindex

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math"
	"sort"
)

// ResearchPartition is an ID ownership rule, not semantic query routing. All
// partitions must be queried. Changing the count requires a new whole layout.
func ResearchPartition(id string, count int) (int, error) {
	if id == "" || count < 1 || count > 16 {
		return 0, errors.New("invalid partition key or count")
	}
	h := sha256.Sum256([]byte(id))
	return int(binary.LittleEndian.Uint64(h[:8]) % uint64(count)), nil
}

// MergeResearchPartitions merges disjoint immutable partitions. Exact top-k
// requires exact local top-k inputs; ANN inputs retain their recall limitation.
// This is NOT a merge for overlapping versions, tombstones or stale index runs.
func MergeResearchPartitions(parts [][]Candidate, k int) ([]Candidate, error) {
	if len(parts) < 1 || len(parts) > 16 || k < 1 || k > 200 {
		return nil, errors.New("invalid partition merge bounds")
	}
	all := make([]Candidate, 0, len(parts)*k)
	seen := make(map[string]bool, len(parts)*k)
	for p, candidates := range parts {
		if len(candidates) > k {
			return nil, errors.New("too many local candidates")
		}
		for _, c := range candidates {
			owner, err := ResearchPartition(c.ID, len(parts))
			if err != nil || owner != p || seen[c.ID] || math.IsNaN(c.Score) || math.IsInf(c.Score, 0) {
				return nil, errors.New("invalid partition candidate")
			}
			seen[c.ID] = true
			all = append(all, c)
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
