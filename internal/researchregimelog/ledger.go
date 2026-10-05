package researchregimelog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
)

// Token binds evidence, configuration, clock, and latent-support publication.
type Token struct {
	Version, SupportEpoch uint64
	LawID                 string
}

type Snapshot struct {
	Token
	AsOf                    int64
	Rows                    []Row
	Support                 [][]Key
	NextClean               []float64 // Before any new issue's support selection.
	LogEvidence, TVEnvelope float64
}

type Receipt struct {
	Ordinal      int
	Clean, First float64 // Latent Y and observed W1 under the issued mask.
	Token
}

type Branch struct {
	Probability    float64
	LogProbability float64
	Underflow      bool // Finite log mass that is too small for a positive float64.
	NextClean      []float64
}

type Query struct {
	Token
	Branches [2]Branch
}

// Ledger is single-owner and isolated from daemon, stores and authority gates.
// Unknown rows advance the original issue clock; reveal never inserts a row.
type Ledger struct {
	base  []float64
	cfg   Config
	rows  []Row
	clock int64
	state Result
	token Token
	cache []prefix
}

func New(base []float64, cfg Config) (*Ledger, error) {
	r, err := Run(base, cfg, nil)
	if err != nil {
		return nil, err
	}
	m := &Ledger{base: append([]float64(nil), base...), cfg: cfg, clock: -1, state: r}
	m.token = m.identity(nil, r.Support, -1, 0, 0)
	return m, nil
}

func (m *Ledger) identity(rows []Row, support [][]Key, clock int64, version, epoch uint64) Token {
	// Empty histories have one identity whether obtained live or by readback.
	if len(rows) == 0 {
		rows = nil
	}
	if len(support) == 0 {
		support = nil
	}
	payload := struct {
		Base           []float64
		Cfg            Config
		Rows           []Row
		Support        [][]Key
		Clock          int64
		Version, Epoch uint64
	}{m.base, m.cfg, rows, support, clock, version, epoch}
	// All numbers have passed finite/domain validation before publication.
	b, err := json.Marshal(payload)
	if err != nil {
		panic("validated ledger identity cannot be encoded")
	}
	h := sha256.Sum256(b)
	return Token{version, epoch, hex.EncodeToString(h[:])}
}

func copySupport(s [][]Key) [][]Key {
	out := make([][]Key, len(s))
	for i := range s {
		out[i] = append([]Key(nil), s[i]...)
	}
	return out
}

func (m *Ledger) Snapshot() Snapshot {
	return Snapshot{m.token, m.clock, append([]Row(nil), m.rows...), copySupport(m.state.Support),
		append([]float64(nil), m.state.Forecast...), m.state.LogEvidence, m.state.Envelope}
}

func (m *Ledger) publish(rows []Row, r Result, at int64, epoch uint64) {
	token := m.identity(rows, r.Support, at, m.token.Version+1, epoch)
	m.rows, m.state, m.clock, m.token = rows, r, at, token
}

// Append only the new as-of support; do not fit historical masks again.
func (m *Ledger) issueSupport() []Key {
	t := len(m.rows)
	next := make([]component, 0, len(m.state.parts)+Classes)
	if m.cfg.Reset < 1 {
		for _, c := range m.state.parts {
			c.logWeight += math.Log1p(-m.cfg.Reset)
			next = append(next, c)
		}
	}
	if m.cfg.Reset > 0 {
		for _, c := range initial(m.base) {
			c.logWeight += math.Log(m.cfg.Reset)
			c.start = t
			next = append(next, c)
		}
	}
	// There is no observed factor at issue. Member motion changes conditional
	// rates, not these component weights, so it cannot affect mask selection.
	if m.cfg.Cap > 0 && len(next) > m.cfg.Cap {
		next = protectedTop(next, m.cfg, t)
	}
	keys := make([]Key, len(next))
	for j, c := range next {
		keys[j] = Key{c.h, c.start}
	}
	return keys
}

func (m *Ledger) Issue(member int, at int64) (Receipt, error) {
	if member < 0 || member >= len(m.base) || at < 0 || at < m.clock || len(m.rows) >= MaxRows {
		return Receipt{}, errors.New("ledger issue binding/clock/cap")
	}
	rows := append(append([]Row(nil), m.rows...), Row{member, -1, -1})
	support := make([][]Key, len(m.state.Support)+1)
	copy(support, m.state.Support)
	support[len(m.state.Support)] = m.issueSupport()
	r, cache, err := m.replay(rows, support, len(m.rows), true)
	if err != nil {
		return Receipt{}, err
	}
	clean, first := 0., 0.
	for _, c := range r.parts {
		p := atomProbability(rate(c.odds[member]))
		clean += math.Exp(c.logWeight) * p
		eta := noise[c.h%3]
		first += math.Exp(c.logWeight) * (eta + (1-2*eta)*p)
	}
	if !finite(clean) || !finite(first) || clean < 0 || clean > 1 || first < 0 || first > 1 {
		return Receipt{}, errors.New("ledger issued law bounds")
	}
	m.publish(rows, r, at, m.token.SupportEpoch+1)
	m.cache = cache
	return Receipt{len(rows) - 1, clean, first, m.token}, nil
}

func (m *Ledger) bind(token Token, ordinal, which int) error {
	if token != m.token || ordinal < 0 || ordinal >= len(m.rows) || which < 1 || which > 2 {
		return errors.New("ledger stale law or evidence binding")
	}
	r := m.rows[ordinal]
	if (which == 1 && r.First != -1) || (which == 2 && (r.First == -1 || r.Second != -1)) {
		return errors.New("ledger evidence phase")
	}
	return nil
}

func (m *Ledger) evidenceRows(ordinal, which, value int) []Row {
	rows := append([]Row(nil), m.rows...)
	if which == 1 {
		rows[ordinal].First = value
	} else {
		rows[ordinal].Second = value
	}
	return rows
}

// Pending evaluates both branches under one fixed published law. Any rejected
// branch causes abstention; an arbitrary inference error is never assigned mass0.
func (m *Ledger) Pending(token Token, ordinal, which int) (Query, error) {
	if err := m.bind(token, ordinal, which); err != nil {
		return Query{}, err
	}
	q := Query{Token: m.token}
	logTotal := math.Inf(-1)
	for value := 0; value < 2; value++ {
		r, _, err := m.replay(m.evidenceRows(ordinal, which, value), m.state.Support, ordinal, false)
		if err != nil {
			return Query{}, err
		}
		lp := (r.LogEvidence - m.state.LogEvidence) + (r.logCompensation - m.state.logCompensation)
		p := math.Exp(lp)
		if !finite(lp) || lp > 2e-11 || !finite(p) || p > 1+2e-11 {
			return Query{}, errors.New("ledger branch probability bounds")
		}
		q.Branches[value] = Branch{Probability: p, LogProbability: lp, Underflow: p == 0, NextClean: append([]float64(nil), r.Forecast...)}
		logTotal = logAdd(logTotal, lp)
	}
	if math.Abs(logTotal) > 2e-11 {
		return Query{}, errors.New("ledger branch mass inconsistency")
	}
	return q, nil
}

func (m *Ledger) Reveal(token Token, ordinal, which, value int, at int64) error {
	if at < 0 || at < m.clock || value < 0 || value > 1 {
		return errors.New("ledger reveal clock/value")
	}
	if err := m.bind(token, ordinal, which); err != nil {
		return err
	}
	rows := m.evidenceRows(ordinal, which, value)
	r, cache, err := m.replay(rows, m.state.Support, ordinal, true)
	if err != nil {
		return err
	}
	m.publish(rows, r, at, m.token.SupportEpoch)
	m.cache = cache
	return nil
}

// Refresh is explicitly NOT conditioning: it refits masks from available
// evidence, creates a new support epoch, and invalidates outstanding queries.
func (m *Ledger) Refresh(token Token, at int64) error {
	if token != m.token || at < 0 || at < m.clock {
		return errors.New("ledger refresh binding/clock")
	}
	r, err := Run(m.base, m.cfg, m.rows)
	if err != nil {
		return err
	}
	r, cache, err := m.rebuild(m.rows, r.Support)
	if err != nil {
		return err
	}
	m.publish(m.rows, r, at, m.token.SupportEpoch+1)
	m.cache = cache
	return nil
}
