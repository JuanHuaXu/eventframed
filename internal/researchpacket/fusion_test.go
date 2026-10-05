package researchpacket

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestPromoteWithinBaseline(t *testing.T) {
	tests := []struct {
		baseline, priority, want []string
	}{
		{[]string{"a", "b", "c"}, []string{"c", "z"}, []string{"c", "a", "b"}},
		{[]string{"a", "b"}, []string{"z", "a"}, []string{"a", "b"}},
		{[]string{"a", "b"}, nil, []string{"a", "b"}},
		{nil, []string{"z"}, nil},
	}
	for _, tt := range tests {
		got, err := PromoteWithinBaseline(tt.baseline, tt.priority)
		if err != nil || !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("got %v, %v; want %v", got, err, tt.want)
		}
	}
}

func TestSupportAndOrderProperties(t *testing.T) {
	rng := rand.New(rand.NewSource(451))
	for trial := 0; trial < 2000; trial++ {
		perm := rng.Perm(20)
		baseline := make([]string, rng.Intn(11))
		for i := range baseline {
			baseline[i] = fmt.Sprint(perm[i])
		}
		priority := []string{fmt.Sprint(rng.Intn(20))}
		got, err := PromoteWithinBaseline(baseline, priority)
		if err != nil || len(got) != len(baseline) {
			t.Fatalf("trial %d: size or error: %v, %v", trial, got, err)
		}
		members := make(map[string]bool, len(baseline))
		for _, id := range baseline {
			members[id] = true
		}
		for _, id := range got {
			if !members[id] {
				t.Fatalf("trial %d: gained support %s", trial, id)
			}
			delete(members, id)
		}
		if len(members) != 0 {
			t.Fatalf("trial %d: lost support %v", trial, members)
		}
		if len(baseline) > 0 {
			want := baseline[0]
			for _, id := range baseline {
				if id == priority[0] {
					want = id
				}
			}
			if got[0] != want {
				t.Fatalf("trial %d: first %s, want %s", trial, got[0], want)
			}
			withoutWinner := func(ids []string) []string {
				out := make([]string, 0, len(ids))
				for _, id := range ids {
					if id != got[0] {
						out = append(out, id)
					}
				}
				return out
			}
			if !reflect.DeepEqual(withoutWinner(got), withoutWinner(baseline)) {
				t.Fatalf("trial %d: non-winner order changed", trial)
			}
		}
	}
}

func TestRejectsInvalidPackets(t *testing.T) {
	for _, tt := range []struct{ b, p []string }{
		{[]string{"a", "a"}, nil},
		{[]string{"a"}, []string{"b", "b"}},
		{[]string{""}, nil},
		{nil, []string{""}},
	} {
		if _, err := PromoteWithinBaseline(tt.b, tt.p); err == nil {
			t.Fatalf("accepted invalid packet: %v %v", tt.b, tt.p)
		}
	}
}

func BenchmarkPromoteWithinBaseline(b *testing.B) {
	base := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	priority := []string{"i", "a", "z"}
	for i := 0; i < b.N; i++ {
		if _, err := PromoteWithinBaseline(base, priority); err != nil {
			b.Fatal(err)
		}
	}
}
