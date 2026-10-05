package researchindex

import (
	"errors"
	"math"
	"sort"
)

// RunEntry describes the complete manifest, including tombstones, of one
// immutable run. Runs are supplied newest first; each run has at most one
// version per ID. Graphs contain only nondeleted entries from their own run.
type RunEntry struct {
	ID      string
	Deleted bool
}

// RunMergePlan is prepared off the query path from a coherent run snapshot.
// It owns its maps. Publishing it with mismatched graph handles is unsupported.
type RunMergePlan struct {
	members []map[string]bool
	owner   map[string]int
	limits  []int
	k       int
}

func NewRunMergePlan(runs [][]RunEntry, k, candidateCap int) (*RunMergePlan, error) {
	if len(runs) < 1 || len(runs) > 16 || k < 1 || k > 200 || candidateCap < k || candidateCap > 328 {
		return nil, errors.New("invalid run merge bounds")
	}
	p := &RunMergePlan{members: make([]map[string]bool, len(runs)), owner: make(map[string]int), limits: make([]int, len(runs)), k: k}
	for i, run := range runs {
		members := make(map[string]bool, len(run))
		live, shadow := 0, 0
		for _, r := range run {
			if _, exists := members[r.ID]; exists || r.ID == "" {
				return nil, errors.New("invalid run manifest")
			}
			members[r.ID] = r.Deleted
			_, newer := p.owner[r.ID]
			if !r.Deleted {
				live++
				if newer {
					shadow++
				}
			}
			if !newer {
				p.owner[r.ID] = i
			}
		}
		// At most shadow prefix entries can be discarded. With an exact
		// sorted prefix, k+shadow therefore suffices, or the entire run.
		need := live
		if shadow < live && k < live-shadow {
			need = k + shadow
		}
		if need > candidateCap {
			return nil, ErrCapacity
		}
		p.members[i] = members
		p.limits[i] = need
	}
	return p, nil
}

func (p *RunMergePlan) Limits() []int { return append([]int(nil), p.limits...) }

// Merge requires the full requested prefix from every run. It cannot prove
// that an ANN prefix is exact: ANN approximation remains explicit. Membership,
// ordering and prefix-length validation never certify unseen ANN candidates.
func (p *RunMergePlan) Merge(prefixes [][]Candidate) ([]Candidate, error) {
	if len(prefixes) != len(p.members) {
		return nil, errors.New("run count mismatch")
	}
	all := make([]Candidate, 0, len(prefixes)*p.k)
	for i, prefix := range prefixes {
		if len(prefix) != p.limits[i] {
			return nil, errors.New("incomplete or oversized run prefix")
		}
		seen := make(map[string]bool, len(prefix))
		for j, c := range prefix {
			deleted, exists := p.members[i][c.ID]
			if !exists || deleted || seen[c.ID] || math.IsNaN(c.Score) || math.IsInf(c.Score, 0) {
				return nil, errors.New("invalid run candidate")
			}
			if j > 0 && runCandidateLess(c, prefix[j-1]) {
				return nil, errors.New("unsorted run prefix")
			}
			seen[c.ID] = true
			if p.owner[c.ID] == i {
				all = append(all, c)
			}
		}
	}
	sort.Slice(all, func(i, j int) bool { return runCandidateLess(all[i], all[j]) })
	return all[:min(p.k, len(all))], nil
}

func runCandidateLess(a, b Candidate) bool {
	if a.Score == b.Score {
		return a.ID < b.ID
	}
	return a.Score > b.Score
}
