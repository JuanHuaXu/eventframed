package observationlearners

import "testing"

// Research-only state machine. Evidence admission and surprise calculation
// belong to the caller; no outcome, hidden generator state, or future schedule
// enters this budget interface.
type queryBurstSchedule struct {
	clock, block, spent, until int
	started                    bool
}

func (s *queryBurstSchedule) step(clock int, surprise, available bool) bool {
	if clock < 0 || clock > 255 || (s.started && clock != s.clock+1) || (!s.started && clock != 0) {
		panic("query burst requires consecutive experiment clocks")
	}
	block := clock / 32
	if !s.started || block != s.block {
		s.spent = 0
	}
	if !s.started {
		s.until = -1
	}
	s.clock, s.block, s.started = clock, block, true
	if surprise {
		s.until = clock + 3
	}
	budget, end := 4, block*32+31
	if block == 0 {
		budget = 3
	}
	if block == 7 {
		end = 248
	}
	remaining := budget - s.spent
	if clock < 8 || clock > 248 || remaining <= 0 || !available {
		return false
	}
	if clock <= s.until || end-clock+1 <= remaining {
		s.spent++
		return true
	}
	return false
}

func TestQueryBurstBudgetContracts(t *testing.T) {
	// Exhaust all periodic trigger patterns up to8 bits; include empty pools.
	for mask := 0; mask < 256; mask++ {
		for empty := 0; empty < 3; empty++ {
			var s queryBurstSchedule
			var counts [8]int
			for clock := 0; clock < 256; clock++ {
				available := empty == 0 || (empty == 1 && clock%3 != 0)
				query := s.step(clock, mask&(1<<uint(clock%8)) != 0, available)
				if query {
					if !available || clock < 8 || clock > 248 {
						t.Fatal("invalid acquisition")
					}
					counts[clock/32]++
				}
			}
			for block, count := range counts {
				budget := 4
				if block == 0 {
					budget = 3
				}
				if count > budget || (empty == 0 && count != budget) || (empty == 2 && count != 0) {
					t.Fatalf("mask%d empty%d block%d count%d budget%d", mask, empty, block, count, budget)
				}
			}
		}
	}
}

func TestQueryBurstTimingAndExpiry(t *testing.T) {
	var quiet, shocked queryBurstSchedule
	var q, b []int
	for clock := 0; clock < 64; clock++ {
		if quiet.step(clock, false, true) {
			q = append(q, clock)
		}
		if shocked.step(clock, clock == 8, true) {
			b = append(b, clock)
		}
	}
	wantQuiet, wantBurst := []int{29, 30, 31, 60, 61, 62, 63}, []int{8, 9, 10, 60, 61, 62, 63}
	for i := range wantQuiet {
		if len(q) != 7 || len(b) != 7 || q[i] != wantQuiet[i] || b[i] != wantBurst[i] {
			t.Fatalf("quiet%v burst%v", q, b)
		}
	}
	// Empty pools neither spend nor carry a previous block's allowance.
	var s queryBurstSchedule
	count := 0
	for clock := 0; clock < 64; clock++ {
		if s.step(clock, true, clock >= 32) {
			count++
		}
	}
	if count != 4 {
		t.Fatalf("expired tokens carried: %d", count)
	}
}
