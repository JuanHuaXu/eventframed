// Package researchsparse is a frozen research scorer, not an online trainer or
// authority to change production forecasts. It implements the ASCII pilot map.
package researchsparse

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math"
	"sort"
	"strings"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

const Dimension = 264
const Contract = "research-sparse-ascii-v1"

var names = [6]string{"who", "what", "where", "when", "why", "how"}
var stops = func() map[string]bool {
	m := map[string]bool{}
	for _, s := range strings.Fields("a an the is was of to at in on and what which when according retained records date s") {
		m[s] = true
	}
	return m
}()

// Features is immutable outside this package; metadata and content never enter
// extraction. Non-ASCII inputs reject rather than silently change Python's
// Unicode case/token semantics. A multilingual contract needs separate testing.
type Features struct {
	values [Dimension]float64
	valid  bool
}

func (f Features) Values() [Dimension]float64 { return f.values }
func words(s string, limit int) (map[string]bool, error) {
	if len(s) > limit {
		return nil, errors.New("research text exceeds cap")
	}
	for i := range s {
		if s[i] >= 128 {
			return nil, errors.New("research ASCII contract violation")
		}
	}
	m := map[string]bool{}
	for _, v := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9') }) {
		if !stops[v] {
			m[v] = true
		}
	}
	return m, nil
}

func Extract(query string, event model.Event) (Features, error) {
	f := Features{}
	q, e := words(query, 1024)
	if e != nil {
		return f, e
	}
	if len(q) == 0 || len(q) > 64 {
		return f, errors.New("invalid research query token count")
	}
	terms := make([]string, 0, len(q))
	for t := range q {
		terms = append(terms, t)
	}
	sort.Strings(terms)
	fields := [6]model.Field{event.Who, event.What, event.Where, event.When, event.Why, event.How}
	union := map[string]bool{}
	f.values[0] = 1
	for j, field := range fields {
		w, e := words(field.Value, 2048)
		if e != nil {
			return Features{}, e
		}
		for t := range w {
			union[t] = true
		}
		matched := 0
		for _, term := range terms {
			flag := "0"
			if w[term] {
				flag = "1"
				matched++
			}
			h := sha256.Sum256([]byte(names[j] + ":" + flag + ":" + term))
			k := 8 + binary.BigEndian.Uint32(h[:4])%256
			sign := -1.
			if h[4]&1 != 0 {
				sign = 1
			}
			f.values[k] += sign / float64(len(terms))
		}
		f.values[j+1] = float64(matched) / float64(len(terms))
	}
	matched := 0
	for _, term := range terms {
		if union[term] {
			matched++
		}
	}
	f.values[7] = float64(matched) / float64(len(terms))
	f.valid = true
	return f, nil
}

type Frozen struct {
	weights [Dimension]float64
	ready   bool
}

func New(weights []float64) (Frozen, error) {
	f := Frozen{}
	if len(weights) != Dimension {
		return f, errors.New("invalid research weight dimension")
	}
	for _, w := range weights {
		if math.IsNaN(w) || math.IsInf(w, 0) || math.Abs(w) > 1e6 {
			return f, errors.New("invalid research weight")
		}
	}
	copy(f.weights[:], weights)
	f.ready = true
	return f, nil
}
func (f Frozen) Score(x Features) (float64, error) {
	if !f.ready || !x.valid {
		return 0, errors.New("uninitialized research scorer")
	}
	z := 0.
	for k, v := range x.values {
		z += f.weights[k] * v
	}
	z = math.Max(-40, math.Min(40, z))
	return 1 / (1 + math.Exp(-z)), nil
}
