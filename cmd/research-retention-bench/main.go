// Public-API performance readback. No base model, database or service is opened.
package main

import (
	"encoding/json"
	retention "github.com/JuanHuaXu/eventframed/internal/researchretention"
	"os"
	"testing"
)

var weightSink retention.Weights
var forecastSink float64
var selectorSink *retention.Selector

type timing struct {
	Name                    string
	Repetition, Iterations  int
	NSPerOp                 int64
	BytesPerOp, AllocsPerOp int64
}

func fixture(members int) (*retention.Selector, retention.Ticket, retention.Forecasts) {
	var f retention.Forecasts
	for k, q := range [3]float64{.2, .5, .8} {
		f[k].Clean = q
		// eta=.1: P00=(1-q)*.81+q*.01, P01=P10=.09.
		f[k].Joint = [4]float64{(1-q)*.81 + q*.01, .09, .09, q*.81 + (1-q)*.01}
	}
	s, e := retention.New(members, 1)
	if e != nil {
		panic(e)
	}
	var first retention.Ticket
	for n := 0; n < retention.MaxTrials; n++ {
		ticket, e := s.Issue(0, int64(2*n), f)
		if e != nil {
			panic(e)
		}
		if n == 0 {
			first = ticket
		}
		if _, e = s.Resolve(ticket, n%2 == 0, int64(2*n+1)); e != nil {
			panic(e)
		}
	}
	return s, first, f
}
func main() {
	s, first, f := fixture(retention.MaxMembers)
	methods := []struct {
		name string
		fn   func(*testing.B)
	}{
		{"weights_with_observable_sink", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				w, e := s.Weights(0)
				if e != nil {
					panic(e)
				}
				weightSink = w
			}
		}},
		{"smoothed_query_with_observable_sink", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				q, e := s.QuerySecond(first)
				if e != nil {
					panic(e)
				}
				forecastSink = q
			}
		}},
		{"constructor_150", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				x, e := retention.New(150, 1)
				if e != nil {
					panic(e)
				}
				selectorSink = x
			}
		}},
		{"complete_64_trial_member_loop_150_member_capacity", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				x, e := retention.New(150, 1)
				if e != nil {
					panic(e)
				}
				var oldest retention.Ticket
				for j := 0; j < retention.MaxTrials; j++ {
					ticket, e := x.Issue(0, int64(2*j), f)
					if e != nil {
						panic(e)
					}
					if j == 0 {
						oldest = ticket
					}
					forecastSink = ticket.Forecast()
					if _, e = x.Resolve(ticket, j%2 == 0, int64(2*j+1)); e != nil {
						panic(e)
					}
				}
				second, e := x.RequestSecond(oldest, 128)
				if e != nil {
					panic(e)
				}
				forecastSink = second.Forecast()
				if _, e = x.Resolve(second, false, 129); e != nil {
					panic(e)
				}
				selectorSink = x
			}
		}},
	}
	var rows []timing
	for _, m := range methods {
		for repeat := 1; repeat <= 3; repeat++ {
			result := testing.Benchmark(m.fn)
			rows = append(rows, timing{m.name, repeat, result.N, result.NsPerOp(), result.AllocedBytesPerOp(), result.AllocsPerOp()})
		}
	}
	out := struct {
		Scope    string
		Timings  []timing
		Weights  retention.Weights
		Forecast float64
	}{
		"Only public selector API; explicit observable sinks. Complete-loop row includes selector construction,64 issue/first resolves,origin request/second resolve. Excludes all three base-model costs,nomination,persistence and loaded serving.", rows, weightSink, forecastSink}
	if e := json.NewEncoder(os.Stdout).Encode(out); e != nil {
		panic(e)
	}
}
