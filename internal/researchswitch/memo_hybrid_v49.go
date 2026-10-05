package researchswitch

import (
	"errors"

	"github.com/JuanHuaXu/eventframed/internal/researchdispersion"
	"github.com/JuanHuaXu/eventframed/internal/researchmoment"
)

// This constructor copies V47 admission and ownership exactly. Only the
// deterministic moment prior initialization uses constructor-local memoization.
func newMemoLocalV49(base []float64, cfg Config, memberTrials int, epoch uint64) (*LocalPool, error) {
	if len(cfg.Prior) != 3 || len(base) < 2 || len(base) > 200 || memberTrials < 1 || memberTrials > researchdispersion.MaxTrials || cfg.Trials < 1 || cfg.Trials > MaxTrials || cfg.Trials > len(base)*memberTrials || cfg.Pending < 1 || cfg.Pending > cfg.Trials {
		return nil, errors.New("invalid local pool caps or arity")
	}
	local := cfg
	local.Trials, local.Pending = memberTrials, min(cfg.Pending, memberTrials)
	first, err := New(local, epoch)
	if err != nil {
		return nil, err
	}
	full, err := researchdispersion.NewWindowObserver(base, epoch, cfg.Pending, "full")
	if err != nil {
		return nil, err
	}
	adaptive, err := researchdispersion.NewWindowObserver(base, epoch, cfg.Pending, "adaptive")
	if err != nil {
		return nil, err
	}
	moment, err := researchmoment.NewMemoV49(base, epoch, cfg.Pending, researchmoment.Config{Family: "rich", Prior: "moment", Strength: 2, Hazard: 1. / 16, Shared: true})
	if err != nil {
		return nil, err
	}
	mixes := make([]*Model, len(base))
	mixes[0] = first
	for i := 1; i < len(base); i++ {
		mixes[i], err = New(local, epoch)
		if err != nil {
			return nil, err
		}
	}
	cfg.Prior = append([]float64(nil), cfg.Prior...)
	return &LocalPool{full: full, adaptive: adaptive, moment: moment, mixes: mixes, base: append([]float64(nil), base...), rows: make([]poolRow, cfg.Trials), issued: make([]int, len(base)), cfg: cfg, memberTrials: memberTrials, epoch: epoch}, nil
}

// NewMemoHybridV49 returns the original hybrid type. No serving/update method
// is forked. Epoch rebuilds deliberately retain the original constructor;
// this experiment claims only initial-construction savings, not reset savings.
func NewMemoHybridV49(base []float64, cfg Config, memberTrials int, globalHazard, scopeHazard float64, epoch uint64) (*HybridPool, error) {
	globalCfg := cfg
	globalCfg.Hazard = globalHazard
	global, err := newCompactV48(compactConfigV48(globalCfg), epoch)
	if err != nil {
		return nil, err
	}
	scope, err := newCompactV48(compactConfigV48{Prior: []float64{.9, .1}, Hazard: scopeHazard, Trials: cfg.Trials, Pending: cfg.Pending}, epoch)
	if err != nil {
		return nil, err
	}
	local, err := newMemoLocalV49(base, cfg, memberTrials, epoch)
	if err != nil {
		return nil, err
	}
	return &HybridPool{local, global, scope, globalHazard, scopeHazard}, nil
}
