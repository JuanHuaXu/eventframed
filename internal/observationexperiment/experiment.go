// Package observationexperiment supplies synthetic outcomes only to fitting and
// scoring. The controller sees EventFrames through a budgeted Reader interface.
package observationexperiment

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/observation"
)

const FitSeed int64 = 2026091301
const DesignSeed int64 = 2026091302
const ConfirmationSeed int64 = 2026091303
const Trials = 256

var Families = []string{"local_detail", "episode_relation", "process_relation", "mixed_scope", "irrelevant_novelty", "missing_evidence", "regime_shift"}

func target(family string, x uint16, shifted bool, rng *rand.Rand) bool {
	bit := func(i int) bool { return x&(1<<i) != 0 }
	var y bool
	switch family {
	case "local_detail":
		y = bit(2)
	case "episode_relation":
		y = bit(3) != bit(4)
	case "process_relation":
		y = (bit(6) != bit(7)) != bit(8)
	case "mixed_scope":
		n := 0
		for _, i := range []int{0, 4, 8} {
			if bit(i) {
				n++
			}
		}
		y = n >= 2
	case "irrelevant_novelty":
		y = bit(0)
	case "missing_evidence":
		return rng.Intn(2) == 1
	case "regime_shift":
		if shifted {
			y = bit(2)
		} else {
			y = (bit(6) != bit(7)) != bit(8)
		}
	default:
		panic("unknown generator family")
	}
	if rng.Float64() < .05 {
		y = !y
	}
	return y
}

func FitFamily(index int) (*observation.Model, error) {
	if index < 0 || index >= len(Families) {
		return nil, fmt.Errorf("invalid family")
	}
	rng := rand.New(rand.NewSource(FitSeed + int64(index)*10000))
	samples := make([]observation.Sample, 4096)
	for i := range samples {
		x := uint16(rng.Intn(observation.Universe))
		samples[i] = observation.Sample{Bits: x, Outcome: target(Families[index], x, false, rng)}
	}
	return observation.Fit(samples)
}

// Frames are simulated observations, not facts about real entities. Why and
// Content deliberately suggest ungrounded explanations and are never predictors.
func Frames(x uint16, id string) *observation.EventReader {
	asof := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	r := &observation.EventReader{Tenant: "mmm-synthetic", AsOf: asof, Version: 1}
	field := func(value string) model.Field {
		return model.Field{Value: value, Source: model.SourceObserved, Confidence: 1, Evidence: "synthetic simulator observation"}
	}
	value := func(bit int) string {
		if x&(1<<bit) != 0 {
			return "1"
		}
		return "0"
	}
	for scope := 0; scope < 3; scope++ {
		for stage := 0; stage < 2; stage++ {
			when := asof.Add(-time.Duration((3-scope)*10+2-stage) * time.Minute)
			e := model.Event{ID: fmt.Sprintf("%s-%d-%d", id, scope, stage), TenantID: r.Tenant, SessionID: id, Sequence: uint64(stage + 1), Kind: "synthetic_observation",
				Content: "Novel but unverified explanation: this must succeed", OccurredAt: when, ObservedAt: when, AvailableAt: when,
				Who: field("simulated observer"), What: field(value(3*scope + 2)), Where: field(fmt.Sprintf("scope-%d", scope)), When: field(when.Format(time.RFC3339)),
				How: field(value(3*scope + 1)), Why: model.Field{Value: "success is certain", Source: model.SourceInferred, Confidence: 1, Evidence: "unsupported fixture hypothesis"},
				Priority: .5, Provenance: model.Provenance{Producer: "eventframe-mmm-simulator"}}
			if stage == 1 {
				e.What = field(value(3 * scope))
			}
			r.Frames[scope][stage] = e
		}
	}
	return r
}

type Row struct {
	Split   string             `json:"split"`
	Family  string             `json:"family"`
	Trial   int                `json:"trial"`
	Shifted bool               `json:"shifted"`
	Policy  string             `json:"policy"`
	Outcome bool               `json:"outcome"`
	Result  observation.Result `json:"result"`
	Brier   float64            `json:"brier"`
	LogLoss float64            `json:"log_loss"`
	Correct bool               `json:"correct"`
}
type Summary struct {
	Split           string  `json:"split"`
	Family          string  `json:"family"`
	Policy          string  `json:"policy"`
	N               int     `json:"n"`
	Brier           float64 `json:"brier"`
	LogLoss         float64 `json:"log_loss"`
	Accuracy        float64 `json:"accuracy"`
	MeanCost        float64 `json:"mean_cost"`
	MeanObserved    float64 `json:"mean_observed"`
	ConfidentErrors int     `json:"confident_errors"`
}
type Comparison struct {
	Split     string  `json:"split"`
	Family    string  `json:"family"`
	Control   string  `json:"control"`
	N         int     `json:"n"`
	BrierGain float64 `json:"brier_gain"`
	Lower     float64 `json:"lower"`
	Upper     float64 `json:"upper"`
	HarmFlag  bool    `json:"harm_flag"`
}
type Output struct {
	Schema                string            `json:"schema"`
	Protocol              string            `json:"protocol"`
	Hashes                map[string]string `json:"sha256"`
	FitSeed               int64             `json:"fit_seed"`
	DesignSeed            int64             `json:"design_seed"`
	ConfirmationSeed      int64             `json:"confirmation_seed"`
	Runtime               string            `json:"runtime"`
	Rows                  []Row             `json:"rows,omitempty"`
	Summaries             []Summary         `json:"summaries"`
	Comparisons           []Comparison      `json:"comparisons"`
	PrimaryPassed         bool              `json:"primary_passed"`
	AnyConfirmationHarm   bool              `json:"any_confirmation_harm"`
	ModelSupportUnchanged bool              `json:"model_support_unchanged"`
}

func Run() (Output, error) {
	out := Output{Schema: "eventframe.mmm-observation.v1", Protocol: "docs/experiments/mmm-attention-v1-protocol.md", FitSeed: FitSeed, DesignSeed: DesignSeed, ConfirmationSeed: ConfirmationSeed, ModelSupportUnchanged: true}
	for f, family := range Families {
		m, err := FitFamily(f)
		if err != nil {
			return out, err
		}
		support := m.Support()
		for splitIndex, split := range []string{"design", "confirmation"} {
			seed := DesignSeed
			if splitIndex == 1 {
				seed = ConfirmationSeed
			}
			rng := rand.New(rand.NewSource(seed + int64(f)*10000))
			for trial := 0; trial < Trials; trial++ {
				x := uint16(rng.Intn(observation.Universe))
				shifted := family == "regime_shift" && trial >= Trials/2
				y := target(family, x, shifted, rng)
				reader := Frames(x, fmt.Sprintf("%s-%s-%d", split, family, trial))
				for _, policy := range observation.Policies {
					result, err := observation.Run(m, reader, 1, policy, seed+int64(f)*10000+int64(trial))
					if err != nil {
						return out, err
					}
					truth := 0.0
					if y {
						truth = 1
					}
					p := result.Probability
					loss := -math.Log1p(-p)
					if y {
						loss = -math.Log(p)
					}
					out.Rows = append(out.Rows, Row{split, family, trial, shifted, policy, y, result, (p - truth) * (p - truth), loss, (p >= .5) == y})
				}
			}
		}
		if m.Support() != support {
			return out, fmt.Errorf("observation changed fitting support")
		}
	}
	Summarize(&out)
	return out, nil
}

func Summarize(out *Output) {
	for _, split := range []string{"design", "confirmation"} {
		for _, family := range append(append([]string{}, Families...), "stationary_informative", "post_shift") {
			matches := func(r Row) bool {
				if r.Split != split {
					return false
				}
				if family == "stationary_informative" {
					return r.Family != "missing_evidence" && r.Family != "regime_shift"
				}
				if family == "post_shift" {
					return r.Family == "regime_shift" && r.Shifted
				}
				return r.Family == family
			}
			mmm := map[string]float64{}
			for _, r := range out.Rows {
				if matches(r) && r.Policy == "mmm" {
					mmm[fmt.Sprintf("%s/%d", r.Family, r.Trial)] = r.Brier
				}
			}
			for _, policy := range observation.Policies {
				s := Summary{Split: split, Family: family, Policy: policy}
				var gains []float64
				for _, r := range out.Rows {
					if matches(r) && r.Policy == policy {
						s.N++
						s.Brier += r.Brier
						s.LogLoss += r.LogLoss
						s.MeanCost += float64(r.Result.Cost)
						s.MeanObserved += float64(r.Result.Observed)
						if r.Correct {
							s.Accuracy++
						} else if r.Result.Probability <= .1 || r.Result.Probability >= .9 {
							s.ConfidentErrors++
						}
						gains = append(gains, r.Brier-mmm[fmt.Sprintf("%s/%d", r.Family, r.Trial)])
					}
				}
				n := float64(s.N)
				s.Brier /= n
				s.LogLoss /= n
				s.Accuracy /= n
				s.MeanCost /= n
				s.MeanObserved /= n
				out.Summaries = append(out.Summaries, s)
				if policy == "breadth" || policy == "depth_first" {
					mean := 0.0
					for _, g := range gains {
						mean += g / n
					}
					variance := 0.0
					for _, g := range gains {
						variance += (g - mean) * (g - mean) / (n - 1)
					}
					margin := 3.4 * math.Sqrt(variance/n)
					c := Comparison{split, family, policy, s.N, mean, mean - margin, mean + margin, mean < -0.01}
					out.Comparisons = append(out.Comparisons, c)
					if split == "confirmation" && c.HarmFlag {
						out.AnyConfirmationHarm = true
					}
				}
			}
		}
	}
	out.PrimaryPassed = true
	for _, c := range out.Comparisons {
		if c.Split == "confirmation" && c.Family == "stationary_informative" {
			out.PrimaryPassed = out.PrimaryPassed && c.BrierGain >= .02 && c.Lower > 0
		}
	}
}
