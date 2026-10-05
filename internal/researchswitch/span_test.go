package researchswitch

import (
	"math"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/researchswitchref"
)

func TestSparseSpansMatchEveryFullHistoryPrefix(t *testing.T) {
	for _, alpha := range []float64{0, math.SmallestNonzeroFloat64, 1e-12, 1. / 150, .23, 1} {
		m := testModel(t, alpha, MaxTrials, MaxTrials)
		qs := make([][]float64, MaxTrials)
		tickets := make([]Ticket, MaxTrials)
		for j := range qs {
			qs[j] = []float64{.1 + .01*float64(j%70), .2 + .01*float64(j%60), .3 + .01*float64(j%50)}
			var err error
			tickets[j], err = m.Issue(qs[j], int64(j))
			if err != nil {
				t.Fatal(err)
			}
		}
		observed := map[int]bool{}
		// Newest-first, then oldest, interior, adjacent and cancelled gaps.
		for k, j := range []int{4095, 0, 2048, 150, 3000, 149, 2047, 1} {
			y := k%2 == 0
			r, err := m.Resolve(tickets[j], y, int64(MaxTrials+k))
			if err != nil || r.Forecast != tickets[j].Forecast() {
				t.Fatal("original sparse receipt", err)
			}
			observed[j] = y
			prefixes, err := researchswitchref.Prefixes(m.prior[:m.n], alpha, qs, observed)
			if err != nil {
				t.Fatal(err)
			}
			for pos, want := range prefixes {
				got := m.probabilities(m.logsAt(pos))
				for h := range want {
					near(t, got[h], want[h])
				}
			}
		}
		before := m.logsAt(MaxTrials - 1)
		if err := m.Cancel(tickets[777], MaxTrials+20); err != nil {
			t.Fatal(err)
		}
		if before != m.logsAt(MaxTrials-1) {
			t.Fatal("cancellation changed unit emission")
		}
		rows, known := append([]row(nil), m.rows...), append([]int(nil), m.known...)
		m.rows[0].advice[0] = math.NaN() // Later emission must fence older replay.
		rows[0].advice[0] = math.NaN()
		clock, pending := m.clock, m.pending
		// Slot2 cannot encounter slot0, so corrupt a LATER known anchor too.
		m.rows[150].advice[0] = math.NaN()
		if _, err := m.Resolve(tickets[2], true, MaxTrials+21); err == nil {
			t.Fatal("invalid later emission accepted")
		}
		if !reflect.DeepEqual(known, m.known) || m.clock != clock || m.pending != pending {
			t.Fatal("failed insertion changed index or clock")
		}
		for j := range rows {
			if m.rows[j].logPosterior != rows[j].logPosterior || m.rows[j].status != rows[j].status {
				t.Fatal("failed insertion published messages")
			}
		}
		if err := m.BeginEpoch(2, MaxTrials+22); err != nil || len(m.known) != 0 || m.used != 0 {
			t.Fatal("epoch retained sparse anchors", err)
		}
	}
}
