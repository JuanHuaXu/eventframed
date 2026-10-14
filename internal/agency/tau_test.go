package agency

import (
	"github.com/JuanHuaXu/eventframed/internal/model"
	"math"
	"testing"
	"time"
)

func tauFixture() (model.TauForecast, model.TauTurnPolicy, time.Time) {
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	f := model.TauForecast{ID: "law-1", Origin: now, IssuedAt: now, ValidUntil: now.Add(time.Minute), HorizonMS: 60000, NoEvent: .2, TVRadius: .1, Points: []model.TauPoint{{Mark: "complete", DelayMS: 15000, Probability: .8}}}
	p := model.TauTurnPolicy{TargetKind: "complete", Benefit: 1, FalseCost: .5, TurnCost: .1, MinProbability: .6, Margin: .1, LeadMS: 2000, WindowMS: 5000, SlackMS: 3000, MinDelayMS: 1000, CooldownMS: 60000}
	return f, p, now
}

func TestTauUtilityAndJitter(t *testing.T) {
	f, p, now := tauFixture()
	d, err := PlanTau(f, p, now)
	if err != nil || !d.Schedule || math.Abs(d.LowerUtility-.45) > 2e-12 || !d.DeliverAt.Equal(now.Add(10*time.Second)) {
		t.Fatalf("decision=%+v err=%v", d, err)
	}
	f.NoEvent = 0
	f.TVRadius = 0
	f.Points = nil
	for _, delay := range []int64{12000, 15000, 17000, 18000} {
		f.Points = append(f.Points, model.TauPoint{Mark: "complete", DelayMS: delay, Probability: .25})
	}
	p.FalseCost = 0
	p.TurnCost = 0
	p.MinProbability = .4
	d, err = PlanTau(f, p, now.Add(8*time.Second))
	if err != nil || !d.DeliverAt.Equal(now.Add(10*time.Second)) || d.Probability != .5 {
		t.Fatalf("jitter intersection: %+v %v", d, err)
	}
	f, p, now = tauFixture()
	f.Points = []model.TauPoint{{Mark: "other", DelayMS: 15000, Probability: .8}}
	d, err = PlanTau(f, p, now)
	if err != nil || d.Schedule {
		t.Fatal("wrong mark triggered")
	}
	f, p, now = tauFixture()
	p.TurnCost = 1
	d, err = PlanTau(f, p, now)
	if err != nil || d.Schedule {
		t.Fatal("negative utility triggered")
	}
	f, p, now = tauFixture()
	f.TVRadius = .8
	d, err = PlanTau(f, p, now)
	if err != nil || d.Schedule {
		t.Fatal("uncertainty allowance ignored")
	}
	f, p, now = tauFixture()
	f.NoEvent = 1
	f.Points[0].Probability = 0
	d, err = PlanTau(f, p, now)
	if err != nil || d.Schedule {
		t.Fatal("no-event triggered")
	}
}

func TestTauRejectsInvalidInputs(t *testing.T) {
	mutations := []func(*model.TauForecast, *model.TauTurnPolicy){
		func(f *model.TauForecast, p *model.TauTurnPolicy) { f.Points[0].Probability = .9 },
		func(f *model.TauForecast, p *model.TauTurnPolicy) { f.TVRadius = math.NaN() },
		func(f *model.TauForecast, p *model.TauTurnPolicy) { f.IssuedAt = f.IssuedAt.Add(time.Second) },
		func(f *model.TauForecast, p *model.TauTurnPolicy) {
			f.Points = append(f.Points, model.TauPoint{Mark: "complete", DelayMS: 15000})
		},
		func(f *model.TauForecast, p *model.TauTurnPolicy) { p.WindowMS = p.SlackMS - 1 },
		func(f *model.TauForecast, p *model.TauTurnPolicy) { p.Margin = 0 },
	}
	for i, mutate := range mutations {
		f, p, now := tauFixture()
		mutate(&f, &p)
		if _, err := PlanTau(f, p, now); err == nil {
			t.Fatalf("mutation %d accepted", i)
		}
	}
}

func TestTauTieAndStrictMargin(t *testing.T) {
	f0, p0, now0 := tauFixture()
	p0.Margin = .45
	if d, err := PlanTau(f0, p0, now0); err != nil || d.Schedule {
		t.Fatalf("decimal strict margin: %+v %v", d, err)
	}
	f, p, now := tauFixture()
	f.Points = []model.TauPoint{{Mark: "complete", DelayMS: 15000, Probability: .4}, {Mark: "complete", DelayMS: 30000, Probability: .4}}
	p.MinProbability = .1
	p.Margin = .01
	d, err := PlanTau(f, p, now)
	if err != nil {
		t.Fatal(err)
	}
	p.FalseCost = 0
	p.TurnCost = 0
	d, err = PlanTau(f, p, now)
	if err != nil || !d.DeliverAt.Equal(now.Add(10*time.Second)) {
		t.Fatalf("tie=%+v %v", d, err)
	}
	p.Margin = d.LowerUtility
	d, err = PlanTau(f, p, now)
	if err != nil || d.Schedule {
		t.Fatal("margin equality must abstain")
	}
}

func TestTauSuppliedFeasibleCandidates(t *testing.T) {
	f, p, now := tauFixture()
	empty := []int64{}
	p.CandidateDelaysMS = &empty
	if d, err := PlanTau(f, p, now); err != nil || d.Schedule {
		t.Fatalf("explicit empty candidates triggered: %+v %v", d, err)
	}
	offsets := []int64{9000, 10000}
	p.CandidateDelaysMS = &offsets
	f.ValidUntil = now.Add(12 * time.Second)
	d, err := PlanTau(f, p, now)
	if err != nil || !d.Schedule || !d.DeliverAt.Equal(now.Add(9*time.Second)) {
		t.Fatalf("candidate feasibility: %+v %v", d, err)
	}
	f, p, now = tauFixture()
	f.HorizonMS = 7200000
	f.ValidUntil = now.Add(2 * time.Hour)
	f.Points = []model.TauPoint{{Mark: "complete", DelayMS: 15000, Probability: .6}, {Mark: "complete", DelayMS: 3605000, Probability: .2}}
	f.NoEvent = .2
	p.MinProbability = .1
	p.FalseCost = 0
	p.TurnCost = 0
	p.Margin = .01
	start, end := 12, 13
	p.QuietStartUTC = &start
	p.QuietEndUTC = &end
	d, err = PlanTau(f, p, now)
	if err != nil || !d.Schedule || !d.DeliverAt.Equal(now.Add(time.Hour)) {
		t.Fatalf("quiet candidate: %+v %v", d, err)
	}
}

func TestTauFitClosedEpisodes(t *testing.T) {
	_, p, now := tauFixture()
	events := map[string]model.Event{
		"a": {ID: "a", SessionID: "one", OccurredAt: now.Add(-2 * time.Minute), AvailableAt: now.Add(-2 * time.Minute)},
		"b": {ID: "b", SessionID: "one", Kind: "complete", OccurredAt: now.Add(-2*time.Minute + 15*time.Second), AvailableAt: now.Add(-time.Minute)},
		"c": {ID: "c", SessionID: "two", OccurredAt: now.Add(-2 * time.Minute), AvailableAt: now.Add(-2 * time.Minute)},
		"d": {ID: "d", SessionID: "two", Kind: "tau_no_event", OccurredAt: now.Add(-time.Minute), AvailableAt: now.Add(-time.Minute)},
	}
	eps := []model.TauEpisode{{AnchorID: "a", OutcomeID: "b"}, {AnchorID: "c", OutcomeID: "d", NoEvent: true}}
	f, err := FitTau("fit", now, now, 60000, []int64{15000, 60000}, eps, events, p.TargetKind)
	if err != nil || f.NoEvent != .4 || f.Points[0].Probability != .4 {
		t.Fatalf("fit=%+v %v", f, err)
	}
	eps[1] = eps[0]
	if _, err := FitTau("fit", now, now, 60000, []int64{60000}, eps, events, p.TargetKind); err == nil {
		t.Fatal("reused episode accepted")
	}
	eps = eps[:1]
	future := events["b"]
	future.AvailableAt = now.Add(time.Second)
	events["b"] = future
	if _, err := FitTau("fit", now, now, 60000, []int64{60000}, eps, events, p.TargetKind); err == nil {
		t.Fatal("future evidence accepted")
	}
}

func BenchmarkTauPlan32(b *testing.B) {
	f, p, now := tauFixture()
	f.NoEvent = 0
	f.Points = nil
	for i := 1; i <= 32; i++ {
		f.Points = append(f.Points, model.TauPoint{Mark: "complete", DelayMS: int64(i * 1000), Probability: 1.0 / 32})
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := PlanTau(f, p, now); err != nil {
			b.Fatal(err)
		}
	}
}
