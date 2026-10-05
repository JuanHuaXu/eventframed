// This command is isolated research. Only arrived measurements enter a ledger.
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	old "github.com/JuanHuaXu/eventframed/internal/researchregimeincremental"
	newmodel "github.com/JuanHuaXu/eventframed/internal/researchregimeprotected"
)

type event struct {
	Ordinal, Member, First, Second, FirstAt, SecondAt int
	HasSecond                                         bool
	Truth, Observed                                   float64
}
type packet struct{ Ordinal, Which, Value, At int }
type fixture struct {
	Seed            uint64
	Generator       string
	Noise           float64
	Delayed         bool
	Members, Rounds int
	Base            []float64
	Events          []event
	Packets         [][]packet
}

func generate(seed uint64, gen string, eta float64, delayed bool, members, rounds int) fixture {
	f := fixture{Seed: seed, Generator: gen, Noise: eta, Delayed: delayed, Members: members, Rounds: rounds}
	for i := 0; i < members; i++ {
		f.Base = append(f.Base, []float64{.30, .45, .65, .85}[i%4])
	}
	d1, d2 := 0, 1
	if delayed {
		d1, d2 = members, 2*members
	}
	f.Packets = make([][]packet, members*rounds+d2+1)
	r := rand.New(rand.NewSource(int64(seed)))
	for i := 0; i < members*rounds; i++ {
		member, round := i%members, i/members
		blend := 0.
		switch gen {
		case "stationary":
		case "abrupt":
			if round >= rounds/2 {
				blend = 1
			}
		case "gradual":
			blend = math.Max(0, math.Min(1, float64(round-rounds/4)/float64(rounds/2)))
		case "recurring":
			if (round/(rounds/4))%2 == 1 {
				blend = 1
			}
		default:
			panic("undeclared generator")
		}
		p := (1-blend)*f.Base[member] + blend*(1-f.Base[member])
		y := 0
		if r.Float64() < p {
			y = 1
		}
		w1, w2 := y, y
		if r.Float64() < eta {
			w1 = 1 - w1
		}
		if r.Float64() < eta {
			w2 = 1 - w2
		}
		e := event{i, member, w1, w2, i + d1, i + d2, member < members/2, p, eta + (1-2*eta)*p}
		f.Events = append(f.Events, e)
		f.Packets[e.FirstAt] = append(f.Packets[e.FirstAt], packet{i, 1, w1, e.FirstAt})
		if e.HasSecond {
			f.Packets[e.SecondAt] = append(f.Packets[e.SecondAt], packet{i, 2, w2, e.SecondAt})
		}
	}
	return f
}

type issued struct {
	Ordinal      int
	Clean, First float64
	LawID        string
}
type operation struct {
	packet
	QueryError, RevealError string
}
type result struct {
	Arm                                                                            string
	Issued                                                                         []issued
	Operations                                                                     []operation
	CleanBrier, FirstBrier, RealizedFirstBrier                                     float64
	AcceptedFirst, AcceptedSecond, RejectedFirst, RejectedSecond, QueryAbstentions int
	CoreMS                                                                         float64
	AllocatedBytes                                                                 uint64
	FinalLawID                                                                     string
}
type adapter struct {
	issue   func(int, int64) (issued, error)
	reveal  func(int, int, int, int64) error
	pending func(int, int) error
	final   func() string
}

func create(arm string, base []float64) (adapter, error) {
	if arm == "v80_top36" {
		m, err := old.New(base, old.Config{Reset: 1. / 16, Hazard: .25, Cap: 36})
		if err != nil {
			return adapter{}, err
		}
		return adapter{
			issue: func(i int, at int64) (issued, error) {
				r, e := m.Issue(i, at)
				return issued{r.Ordinal, r.Clean, r.First, r.LawID}, e
			},
			reveal:  func(i, w, v int, at int64) error { return m.Reveal(m.Snapshot().Token, i, w, v, at) },
			pending: func(i, w int) error { _, e := m.Pending(m.Snapshot().Token, i, w); return e },
			final:   func() string { return m.Snapshot().LawID },
		}, nil
	}
	cap := 36
	if arm == "v81_reset9" {
		cap = 9
	} else if arm != "v81_protected36" {
		return adapter{}, fmt.Errorf("unknown arm %s", arm)
	}
	m, err := newmodel.New(base, newmodel.Config{Reset: 1. / 16, Hazard: .25, Cap: cap})
	if err != nil {
		return adapter{}, err
	}
	return adapter{
		issue: func(i int, at int64) (issued, error) {
			r, e := m.Issue(i, at)
			return issued{r.Ordinal, r.Clean, r.First, r.LawID}, e
		},
		reveal:  func(i, w, v int, at int64) error { return m.Reveal(m.Snapshot().Token, i, w, v, at) },
		pending: func(i, w int) error { _, e := m.Pending(m.Snapshot().Token, i, w); return e },
		final:   func() string { return m.Snapshot().LawID },
	}, nil
}
func run(f fixture, arm string) (result, error) {
	out := result{Arm: arm, Issued: make([]issued, 0, len(f.Events)), Operations: make([]operation, 0, 2*len(f.Events))}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	begin := time.Now()
	m, err := create(arm, f.Base)
	if err != nil {
		return out, err
	}
	for at, ps := range f.Packets {
		if at < len(f.Events) {
			// The learner never receives the fixture's future outcome or rate.
			x, e := m.issue(f.Events[at].Member, int64(at))
			if e != nil {
				return out, e
			}
			if x.Ordinal != at || !(x.Clean >= 0 && x.Clean <= 1 && x.First >= 0 && x.First <= 1) {
				return out, fmt.Errorf("invalid issue %d", at)
			}
			out.Issued = append(out.Issued, x)
		}
		for _, p := range ps {
			if p.At != at || p.Ordinal > at {
				return out, fmt.Errorf("early packet")
			}
			x := operation{packet: p}
			if p.Which == 2 {
				if e := m.pending(p.Ordinal, 2); e != nil {
					x.QueryError = e.Error()
					out.QueryAbstentions++
				}
			}
			if e := m.reveal(p.Ordinal, p.Which, p.Value, int64(at)); e != nil {
				x.RevealError = e.Error()
				if p.Which == 1 {
					out.RejectedFirst++
				} else {
					out.RejectedSecond++
				}
			} else if p.Which == 1 {
				out.AcceptedFirst++
			} else {
				out.AcceptedSecond++
			}
			out.Operations = append(out.Operations, x)
		}
	}
	out.FinalLawID = m.final()
	out.CoreMS = float64(time.Since(begin)) / 1e6
	runtime.ReadMemStats(&after)
	out.AllocatedBytes = after.TotalAlloc - before.TotalAlloc
	for i, x := range out.Issued {
		e := f.Events[i]
		out.CleanBrier += e.Truth*(1-e.Truth) + math.Pow(x.Clean-e.Truth, 2)
		out.FirstBrier += e.Observed*(1-e.Observed) + math.Pow(x.First-e.Observed, 2)
		out.RealizedFirstBrier += math.Pow(x.First-float64(e.First), 2)
	}
	n := float64(len(out.Issued))
	out.CleanBrier /= n
	out.FirstBrier /= n
	out.RealizedFirstBrier /= n
	return out, nil
}
func save(p string, x any) error {
	b, e := json.MarshalIndent(x, "", "  ")
	if e != nil {
		return e
	}
	f, e := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	if _, e = f.Write(append(b, '\n')); e != nil {
		return e
	}
	return f.Sync()
}
func main() {
	if len(os.Args) != 2 {
		panic("one fresh output directory required")
	}
	root := os.Args[1]
	if e := os.Mkdir(root, 0700); e != nil {
		panic(e)
	}
	var summaries []any
	caseID := 0
	for _, seed := range []uint64{2026105501, 2026105503} {
		for _, gen := range []string{"stationary", "abrupt", "gradual", "recurring"} {
			for _, noise := range []float64{0, .1, .2} {
				for _, delayed := range []bool{false, true} {
					f := generate(seed, gen, noise, delayed, 50, 16)
					arms := []string{"v80_top36", "v81_protected36", "v81_reset9"}
					var rs []result
					for a := 0; a < 3; a++ {
						name := arms[(a+caseID)%3]
						r, e := run(f, name)
						if e != nil {
							panic(e)
						}
						rs = append(rs, r)
					}
					baseline := 0.
					for _, e := range f.Events {
						baseline += e.Truth*(1-e.Truth) + math.Pow(f.Base[e.Member]-e.Truth, 2)
					}
					baseline /= float64(len(f.Events))
					p := filepath.Join(root, fmt.Sprintf("case-%03d.json", caseID))
					if e := save(p, struct {
						Fixture          fixture
						Results          []result
						StaticCleanBrier float64
					}{f, rs, baseline}); e != nil {
						panic(e)
					}
					for _, r := range rs {
						summaries = append(summaries, struct {
							Case                                            int
							Seed                                            uint64
							Generator                                       string
							Noise                                           float64
							Delayed                                         bool
							Arm                                             string
							CleanBrier, FirstBrier, CoreMS                  float64
							AllocatedBytes                                  uint64
							RejectedFirst, RejectedSecond, QueryAbstentions int
						}{caseID, seed, gen, noise, delayed, r.Arm, r.CleanBrier, r.FirstBrier, r.CoreMS, r.AllocatedBytes, r.RejectedFirst, r.RejectedSecond, r.QueryAbstentions})
					}
					fmt.Printf("CASE %d %s noise%.1f delayed%v complete\n", caseID, gen, noise, delayed)
					caseID++
				}
			}
		}
	}
	if e := save(filepath.Join(root, "summary.json"), summaries); e != nil {
		panic(e)
	}
}
