package researchvalidity

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func posteriorFixture() (Record, Request) {
	base := model.Snapshot{
		RuntimeVersion: 10, EvidenceEpoch: 4, PolicyVersion: 2,
		ContractVersion: 16, GraphVersion: 3, AbstractionVersion: 3,
		AgencyVersion: 1,
	}
	record := Record{
		TenantID: "tenant", PosteriorKey: "event:a", QueryDigest: "query",
		ModelID: "model", HorizonKey: "horizon", SourceID: "source",
		SourceAt:       time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		FrontierDigest: "frontier", FrontierK: 50, CutoffLower: 0.4, Base: base,
	}
	request := Request{
		TenantID: record.TenantID, PosteriorKey: record.PosteriorKey,
		QueryDigest: record.QueryDigest, FrontierDigest: record.FrontierDigest,
		ModelID:    record.ModelID,
		HorizonKey: record.HorizonKey, SourceID: record.SourceID,
		AsOf: record.SourceAt.Add(24 * time.Hour), Current: base,
	}
	return record, request
}

func TestPosteriorReuseMutationWitness(t *testing.T) {
	record, request := posteriorFixture()
	visible := request.AsOf.Add(-time.Hour)
	future := request.AsOf.Add(time.Hour)
	cases := []struct {
		name   string
		change func(*Record, *Request) []Mutation
		want   Status
	}{
		{"unchanged", func(*Record, *Request) []Mutation { return nil }, Compatible},
		{"future insert", func(_ *Record, q *Request) []Mutation {
			q.Current.RuntimeVersion++
			q.Current.EvidenceEpoch++
			return []Mutation{{RuntimeVersion: 11, EvidenceEpochAfter: 5, Kind: Insert, AvailableAt: future}}
		}, Compatible},
		{"visible below cutoff", func(_ *Record, q *Request) []Mutation {
			q.Current.RuntimeVersion++
			q.Current.EvidenceEpoch++
			return []Mutation{{RuntimeVersion: 11, EvidenceEpochAfter: 5, Kind: Insert, AvailableAt: visible,
				ScoreBounded: true, ScoreUpper: .39, GraphChecked: true}}
		}, Compatible},
		{"visible competing insert", func(_ *Record, q *Request) []Mutation {
			q.Current.RuntimeVersion++
			q.Current.EvidenceEpoch++
			return []Mutation{{RuntimeVersion: 11, EvidenceEpochAfter: 5, Kind: Insert, AvailableAt: visible,
				ScoreBounded: true, ScoreUpper: .4, GraphChecked: true}}
		}, Blocked},
		{"visible graph dependency", func(_ *Record, q *Request) []Mutation {
			q.Current.RuntimeVersion++
			q.Current.EvidenceEpoch++
			return []Mutation{{RuntimeVersion: 11, EvidenceEpochAfter: 5, Kind: Insert, AvailableAt: visible,
				ScoreBounded: true, ScoreUpper: .39, GraphChecked: true, GraphTouches: true}}
		}, Blocked},
		{"missing score certificate", func(_ *Record, q *Request) []Mutation {
			q.Current.RuntimeVersion++
			q.Current.EvidenceEpoch++
			return []Mutation{{RuntimeVersion: 11, EvidenceEpochAfter: 5, Kind: Insert, AvailableAt: visible,
				GraphChecked: true}}
		}, Unknown},
		{"unrelated outcome", func(_ *Record, q *Request) []Mutation {
			q.Current.RuntimeVersion++
			return []Mutation{{RuntimeVersion: 11, EvidenceEpochAfter: 4, Kind: Outcome, AvailableAt: visible,
				AffectedKnown: true, AffectedKeys: []string{"event:b"}}}
		}, Compatible},
		{"related outcome", func(_ *Record, q *Request) []Mutation {
			q.Current.RuntimeVersion++
			return []Mutation{{RuntimeVersion: 11, EvidenceEpochAfter: 4, Kind: Outcome, AvailableAt: visible,
				AffectedKnown: true, AffectedKeys: []string{"event:a"}}}
		}, Blocked},
		{"unknown outcome dependencies", func(_ *Record, q *Request) []Mutation {
			q.Current.RuntimeVersion++
			return []Mutation{{RuntimeVersion: 11, EvidenceEpochAfter: 4, Kind: Outcome, AvailableAt: visible}}
		}, Unknown},
		{"outcome missing availability", func(_ *Record, q *Request) []Mutation {
			q.Current.RuntimeVersion++
			return []Mutation{{RuntimeVersion: 11, EvidenceEpochAfter: 4, Kind: Outcome,
				AffectedKnown: true, AffectedKeys: []string{"event:b"}}}
		}, Unknown},
		{"future related outcome", func(_ *Record, q *Request) []Mutation {
			q.Current.RuntimeVersion++
			return []Mutation{{RuntimeVersion: 11, EvidenceEpochAfter: 4, Kind: Outcome, AvailableAt: future,
				AffectedKnown: true, AffectedKeys: []string{"event:a"}}}
		}, Compatible},
		{"certificate publication", func(_ *Record, q *Request) []Mutation {
			q.Current.RuntimeVersion++
			return []Mutation{{RuntimeVersion: 11, EvidenceEpochAfter: 4, Kind: Certificate}}
		}, Compatible},
		{"hidden epoch change in certificate", func(_ *Record, q *Request) []Mutation {
			q.Current.RuntimeVersion++
			q.Current.EvidenceEpoch++
			return []Mutation{{RuntimeVersion: 11, EvidenceEpochAfter: 5, Kind: Certificate}}
		}, Unknown},
		{"insert without epoch increment", func(_ *Record, q *Request) []Mutation {
			q.Current.RuntimeVersion++
			return []Mutation{{RuntimeVersion: 11, EvidenceEpochAfter: 4, Kind: Insert, AvailableAt: future}}
		}, Unknown},
		{"missing runtime mutation", func(_ *Record, q *Request) []Mutation {
			q.Current.RuntimeVersion += 2
			return []Mutation{{RuntimeVersion: 12, EvidenceEpochAfter: 4, Kind: Certificate}}
		}, Unknown},
		{"unclassified mutation", func(_ *Record, q *Request) []Mutation {
			q.Current.RuntimeVersion++
			return []Mutation{{RuntimeVersion: 11, EvidenceEpochAfter: 4, Kind: Other}}
		}, Unknown},
		{"query changed", func(_ *Record, q *Request) []Mutation {
			q.QueryDigest = "other"
			return nil
		}, Blocked},
		{"frontier changed", func(_ *Record, q *Request) []Mutation {
			q.FrontierDigest = "other"
			return nil
		}, Blocked},
		{"source unavailable", func(r *Record, _ *Request) []Mutation {
			r.SourceAt = future
			return nil
		}, Blocked},
		{"graph version changed", func(_ *Record, q *Request) []Mutation {
			q.Current.GraphVersion++
			return nil
		}, Blocked},
		{"epoch changed with no runtime witness", func(_ *Record, q *Request) []Mutation {
			q.Current.EvidenceEpoch++
			return nil
		}, Unknown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, q := record, request
			mutations := tc.change(&r, &q)
			got := Check(r, q, mutations)
			if got.Status != tc.want {
				t.Fatalf("Check() = %v (%s), want %v", got.Status, got.Reason, tc.want)
			}
		})
	}
}

func TestPosteriorReuseHarmlessAppendPrefixes(t *testing.T) {
	record, request := posteriorFixture()
	mutations := make([]Mutation, 0, 128)
	for i := 0; i < 128; i++ {
		request.Current.RuntimeVersion++
		if i%2 == 0 {
			request.Current.EvidenceEpoch++
			mutations = append(mutations, Mutation{RuntimeVersion: request.Current.RuntimeVersion,
				EvidenceEpochAfter: request.Current.EvidenceEpoch, Kind: Insert,
				AvailableAt: request.AsOf.Add(time.Hour)})
		} else {
			mutations = append(mutations, Mutation{RuntimeVersion: request.Current.RuntimeVersion,
				EvidenceEpochAfter: request.Current.EvidenceEpoch, Kind: Certificate})
		}
		if got := Check(record, request, mutations); got.Status != Compatible {
			t.Fatalf("prefix %d: %v (%s)", i+1, got.Status, got.Reason)
		}
	}
	request.Current.RuntimeVersion++
	request.Current.EvidenceEpoch++
	mutations = append(mutations, Mutation{RuntimeVersion: request.Current.RuntimeVersion,
		EvidenceEpochAfter: request.Current.EvidenceEpoch, Kind: Insert,
		AvailableAt: request.AsOf, ScoreBounded: true, ScoreUpper: .95, GraphChecked: true})
	if got := Check(record, request, mutations); got.Status != Blocked {
		t.Fatalf("competing append: %v (%s)", got.Status, got.Reason)
	}
}

func TestPosteriorReuseV19FrontierChurn(t *testing.T) {
	record, request := posteriorFixture()
	record.CutoffLower = math.Cos(.74)
	request.Current.RuntimeVersion++
	request.Current.EvidenceEpoch++
	mutation := Mutation{RuntimeVersion: 11, EvidenceEpochAfter: 5, Kind: Insert,
		AvailableAt: request.AsOf, ScoreBounded: true, ScoreUpper: math.Cos(.00213),
		GraphChecked: true}
	if got := Check(record, request, []Mutation{mutation}); got.Status != Blocked {
		t.Fatalf("high-relevance append: %v (%s)", got.Status, got.Reason)
	}
}

func BenchmarkPosteriorReuse(b *testing.B) {
	for _, size := range []int{128, 10000} {
		b.Run(fmt.Sprintf("mutations=%d", size), func(b *testing.B) {
			record, request := posteriorFixture()
			mutations := make([]Mutation, size)
			for i := range mutations {
				mutations[i] = Mutation{RuntimeVersion: record.Base.RuntimeVersion + uint64(i) + 1,
					EvidenceEpochAfter: record.Base.EvidenceEpoch, Kind: Certificate}
			}
			request.Current.RuntimeVersion += uint64(size)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if got := Check(record, request, mutations); got.Status != Compatible {
					b.Fatal(got)
				}
			}
		})
	}
}
