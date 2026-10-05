// Package researchcalendar is a research-only pre-packing calendar adapter.
// Its ordinal scores encode order, not probabilities or calibrated confidence.
package researchcalendar

import (
	"context"
	"errors"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

type binding struct{ tenant, original string }
type bindingKey struct{}

// Bind retains the original request in this context only; no shared mutable map.
func Bind(ctx context.Context, tenant, original string) context.Context {
	return context.WithValue(ctx, bindingKey{}, binding{tenant, original})
}

func Focus(q string) string {
	if strings.Count(q, ", rather than ") != 1 {
		return q
	}
	p, s, _ := strings.Cut(q, ", rather than ")
	if strings.TrimSpace(p) == "" || strings.TrimSpace(s) == "" {
		return q
	}
	return strings.TrimRight(strings.TrimSpace(p), "?") + "?"
}

var written = regexp.MustCompile(`(?i)\b\d{1,2} (January|February|March|April|May|June|July|August|September|October|November|December) \d{4}\b`)
var iso = regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}`)
var bound = regexp.MustCompile(`(?i)\b(before|after) (\d{4})\b`)
var exact = regexp.MustCompile(`(?i)\bon (\d{1,2} (?:January|February|March|April|May|June|July|August|September|October|November|December) \d{4})\b`)
var negation = regexp.MustCompile(`(?i)\b(not|never|unless|except)\b`)
var timeSuffix = regexp.MustCompile(`^\.\d|^\s+\d{2}:\d{2}`)

func Date(text string) (time.Time, bool) {
	w, i := written.FindAllString(text, -1), iso.FindAllStringIndex(text, -1)
	if len(w)+len(i) != 1 {
		return time.Time{}, false
	}
	var date time.Time
	var err error
	if len(w) == 1 {
		date, err = time.Parse("2 January 2006", w[0])
	} else {
		start, end := i[0][0], i[0][1]
		if start > 0 && !strings.ContainsRune(" \t\r\n(:\"'", rune(text[start-1])) {
			return time.Time{}, false
		}
		if end < len(text) && !strings.ContainsRune(" \t\r\n.,;!?)\"'", rune(text[end])) {
			return time.Time{}, false
		}
		if timeSuffix.MatchString(text[end:]) {
			return time.Time{}, false
		}
		date, err = time.Parse("2006-01-02", text[start:end])
	}
	return date, err == nil && date.Year() >= 1000
}

type rule struct {
	kind string
	date time.Time
	year int
}

func rules(q string) []rule {
	parts := strings.Split(q, ", rather than ")
	if len(parts) > 2 || negation.MatchString(q) {
		return nil
	}
	b, e := bound.FindAllStringSubmatch(parts[0], -1), exact.FindAllStringSubmatch(parts[0], -1)
	if len(b)+len(e) > 1 {
		return nil
	}
	var out []rule
	if len(b) == 1 {
		y, err := time.Parse("2006", b[0][2])
		if err != nil {
			return nil
		}
		out = append(out, rule{kind: strings.ToLower(b[0][1]), year: y.Year()})
	}
	if len(e) == 1 {
		d, ok := Date(e[0][1])
		if !ok {
			return nil
		}
		out = append(out, rule{kind: "on", date: d})
	}
	if len(parts) == 2 {
		s := strings.TrimSuffix(strings.TrimSpace(parts[1]), "?")
		m := written.FindString(s)
		if m != s {
			return nil
		}
		d, ok := Date(s)
		if !ok {
			return nil
		}
		out = append(out, rule{kind: "exclude", date: d})
	}
	return out
}

// Only the actual event's what field supplies the date. Other fields can repeat
// that date or contain ingestion timestamps and are not independent evidence.
func what(text string) (string, bool) {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || lines[0] != "representation: "+model.SemanticRepresentationVersion {
		return "", false
	}
	var value string
	count := 0
	for _, line := range lines[1:] {
		if strings.HasPrefix(line, "what: ") {
			count++
			value = strings.TrimPrefix(line, "what: ")
		}
	}
	return value, count == 1
}

func contradict(rs []rule, text string) bool {
	v, ok := what(text)
	if !ok {
		return false
	}
	d, ok := Date(v)
	if !ok {
		return false
	}
	for _, r := range rs {
		switch r.kind {
		case "before":
			if d.Year() >= r.year {
				return true
			}
		case "after":
			if d.Year() <= r.year {
				return true
			}
		case "on":
			if !d.Equal(r.date) {
				return true
			}
		case "exclude":
			if d.Equal(r.date) {
				return true
			}
		}
	}
	return false
}

type Ranker struct{}

func (Ranker) ContractName() string { return "research/calendar-ordinal-v1" }

func (Ranker) RankCandidates(ctx context.Context, req retrieval.RankRequest) ([]retrieval.Candidate, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	b, ok := ctx.Value(bindingKey{}).(binding)
	if !ok || b.tenant == "" || b.tenant != req.UserID || len(b.original) > 4096 || frame.QueryText(Focus(b.original)) != req.QueryText {
		return nil, errors.New("calendar query binding mismatch")
	}
	if len(req.Candidates) > 200 || req.K1 != len(req.Candidates) || req.K2 <= 0 || req.K2 > req.K1 {
		return nil, errors.New("invalid calendar frontier")
	}
	rs := rules(b.original)
	good, bad := make([]retrieval.Candidate, 0, len(req.Candidates)), make([]retrieval.Candidate, 0, len(req.Candidates))
	seen := make(map[string]bool, len(req.Candidates))
	for _, c := range req.Candidates {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if c.ID == "" || seen[c.ID] || len(c.Text) > 8192 || math.IsNaN(c.Score) || math.IsInf(c.Score, 0) || c.Score < 0 || c.Score > 1 {
			return nil, errors.New("invalid calendar candidate")
		}
		seen[c.ID] = true
		if contradict(rs, c.Text) {
			bad = append(bad, c)
		} else {
			good = append(good, c)
		}
	}
	if len(bad) == 0 || len(good) == 0 {
		return append([]retrieval.Candidate(nil), req.Candidates[:req.K2]...), nil
	}
	out := append(good, bad...)
	// Dense ordinal codes make the partition survive a subsequent score sort.
	// They are research ranking signals only; later runtime deltas may still move them.
	for i := range out {
		out[i].Score = float64(len(out)-i) / float64(len(out)+1)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return out[:req.K2], nil
}
