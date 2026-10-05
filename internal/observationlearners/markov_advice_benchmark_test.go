package observationlearners

import (
	"fmt"
	"testing"
)

type markovBenchAdapter struct{ *markovAdviceJournal }

func (g markovBenchAdapter) statsSnapshot() routedJournalStats { return g.journal.stats }
func (g markovBenchAdapter) selectorCount() uint64             { return g.received }

func BenchmarkMarkovAdviceJournal(b *testing.B) {
	a, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		b.Fatal(err)
	}
	c, err := newObservationExperts(jointFixture(true, true))
	if err != nil {
		b.Fatal(err)
	}
	for _, schedule := range []string{"immediate", "delay31", "reverse_batch31"} {
		for _, policy := range []string{"log", "markov"} {
			b.Run(fmt.Sprintf("%s/%s", schedule, policy), func(b *testing.B) {
				b.ReportAllocs()
				for n := 0; n < b.N; n++ {
					var g logV109Journal
					if policy == "log" {
						g = likelihoodV109Adapter{newLogAdviceJournal(false)}
					} else {
						g = markovBenchAdapter{newMarkovAdviceJournal()}
					}
					deliver := func(i int) {
						if err := g.deliver(uint64(i), i%3 == 0); err != nil {
							b.Fatal(err)
						}
					}
					for clock := 0; clock < 288; clock++ {
						if err := g.(interface{ setClock(uint64) error }).setClock(uint64(clock)); err != nil {
							b.Fatal(err)
						}
						if schedule == "delay31" && clock >= 31 && clock-31 < 256 {
							deliver(clock - 31)
						}
						// Descending batches give every preceding event delay1..31.
						if schedule == "reverse_batch31" && clock < 256 && clock%32 == 31 {
							for i := clock - 1; i >= clock-31; i-- {
								deliver(i)
							}
						}
						if clock >= 32 {
							if err := g.expireBefore(uint64(clock - 31)); err != nil {
								b.Fatal(err)
							}
						}
						if clock >= 256 {
							continue
						}
						if clock%32 == 0 {
							e := a
							if clock/32%2 == 1 {
								e = c
							}
							if err := g.publish(e); err != nil {
								b.Fatal(err)
							}
						}
						if _, err := g.predict(uint64(clock), &jointReader{x: uint16(clock), epoch: 1}, 1); err != nil {
							b.Fatal(err)
						}
						if schedule == "immediate" || (schedule == "reverse_batch31" && clock%32 == 31) {
							deliver(clock)
						}
					}
					if g.statsSnapshot().Pending != 0 || g.selectorCount() != 256 {
						b.Fatal("settlement")
					}
				}
			})
		}
	}
}
