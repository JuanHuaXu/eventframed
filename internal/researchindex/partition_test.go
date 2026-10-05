package researchindex

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func TestPartitionMergeMatchesExactOracle(t *testing.T) {
	r := rand.New(rand.NewSource(97169))
	for round := 0; round < 200; round++ {
		count := 1 + r.Intn(16)
		k := 1 + r.Intn(50)
		parts := make([][]Candidate, count)
		var all []Candidate
		for i := 0; i < 500; i++ {
			c := Candidate{fmt.Sprintf("id-%d", i), float64(r.Intn(100))}
			p, _ := ResearchPartition(c.ID, count)
			parts[p] = append(parts[p], c)
			all = append(all, c)
		}
		less := func(a, b Candidate) bool {
			if a.Score == b.Score {
				return a.ID < b.ID
			}
			return a.Score > b.Score
		}
		for i := range parts {
			sort.Slice(parts[i], func(a, b int) bool { return less(parts[i][a], parts[i][b]) })
			parts[i] = parts[i][:min(k, len(parts[i]))]
		}
		sort.Slice(all, func(i, j int) bool { return less(all[i], all[j]) })
		got, err := MergeResearchPartitions(parts, k)
		if err != nil || !reflect.DeepEqual(got, all[:k]) {
			t.Fatal(round, got, err)
		}
	}
}

func TestPartitionMergeRejectsForeignAndDuplicate(t *testing.T) {
	p, _ := ResearchPartition("a", 2)
	parts := make([][]Candidate, 2)
	parts[1-p] = []Candidate{{"a", 1}}
	if _, err := MergeResearchPartitions(parts, 1); err == nil {
		t.Fatal("foreign owner accepted")
	}
	if _, err := MergeResearchPartitions([][]Candidate{{{"a", 1}, {"a", 1}}}, 2); err == nil {
		t.Fatal("duplicate accepted")
	}
	for _, count := range []int{0, 17} {
		if _, err := ResearchPartition("a", count); err == nil {
			t.Fatal("unbounded layout")
		}
	}
}
