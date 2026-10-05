package libravdbstore

import (
	"fmt"
	"math"
	"testing"
	"time"
)

type frontierRowV21 struct {
	id        string
	angle     float64
	available time.Time
}

func frontierAsOfV21(rows []frontierRowV21, asOf time.Time, k int) ([]denseOracleRowV7, error) {
	visible := make([]denseOracleRowV7, 0, len(rows))
	for _, row := range rows {
		if !row.available.After(asOf) {
			visible = append(visible, denseOracleRowV7{id: row.id, angle: row.angle})
		}
	}
	if len(visible) < k {
		return nil, fmt.Errorf("visible rows=%d below k=%d", len(visible), k)
	}
	return sortedDenseOracleV7(visible)[:k], nil
}

func frontierSameV21(left, right []denseOracleRowV7) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i].id != right[i].id {
			return false
		}
	}
	return true
}

func TestResearchFrontierChurnV21(t *testing.T) {
	query := denseQueryV6()
	asOf := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	rows := []frontierRowV21{
		{id: "past100", available: asOf},
		{id: "at120", available: asOf},
	}
	for i := 0; i < 198; i++ {
		rows = append(rows, frontierRowV21{id: fmt.Sprintf("eligible-%03d", i),
			angle: .005 * float64(i+1), available: asOf})
	}
	for i := 0; i < 16; i++ {
		rows = append(rows, frontierRowV21{id: fmt.Sprintf("future-%03d", i),
			angle: .001 * float64(i+1), available: asOf.Add(time.Hour)})
	}
	ks := []int{10, 50, 150}
	changes := make([]int, len(ks))
	firstUnchanged := make([]int, len(ks))
	for i := range firstUnchanged {
		firstUnchanged[i] = -1
	}
	for i := 0; i < 128; i++ {
		angle := .00213 * float64(i+1)
		vector := denseRowV6(query, i, angle)
		var dot, norm float64
		for j := range query {
			dot += float64(query[j]) * float64(vector[j])
			norm += float64(vector[j]) * float64(vector[j])
		}
		if got := dot / math.Sqrt(norm); math.Abs(got-math.Cos(angle)) > 1e-5 {
			t.Fatalf("write %d cosine=%g want=%g", i, got, math.Cos(angle))
		}
		before := make([][]denseOracleRowV7, len(ks))
		for j, k := range ks {
			var err error
			before[j], err = frontierAsOfV21(rows, asOf, k)
			if err != nil {
				t.Fatal(err)
			}
		}
		rows = append(rows, frontierRowV21{id: fmt.Sprintf("visible-%03d", i),
			angle: angle, available: asOf})
		for j, k := range ks {
			after, err := frontierAsOfV21(rows, asOf, k)
			if err != nil {
				t.Fatal(err)
			}
			if frontierSameV21(before[j], after) {
				if firstUnchanged[j] == -1 {
					firstUnchanged[j] = i
				}
			} else {
				changes[j]++
				if firstUnchanged[j] >= 0 {
					t.Fatalf("k=%d changed again after first nonimpacting append %d", k, firstUnchanged[j])
				}
			}
		}
	}
	want := []int{6, 34, 104}
	for j, k := range ks {
		if changes[j] != want[j] || firstUnchanged[j] != want[j] {
			t.Fatalf("k=%d changes=%d first unchanged=%d want=%d", k, changes[j], firstUnchanged[j], want[j])
		}
		final, err := frontierAsOfV21(rows, asOf, k)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("k=%d changes=%d/128 first_nonimpact=%d final_cutoff_angle=%.5f", k,
			changes[j], firstUnchanged[j], final[k-1].angle)
	}
}
