package researchcalendar

import (
	"context"
	"regexp"
	"strings"

	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

type TaskPlan struct {
	PriorityPlan
	Role             string
	TargetRelation   string
	NegatedSelection bool
	Reasons          map[string]string
}

var launchMention = regexp.MustCompile(`(?i)\b(launch|launched|launching)\b`)
var arrivalMention = regexp.MustCompile(`(?i)\b(arrival|arrived|arrive|arriving)\b`)
var alternativeMention = regexp.MustCompile(`(?i)\b(or|unless|except)\b`)
var refutationMention = regexp.MustCompile(`(?i)\b(disproves?|refutes?|false|falsifies?)\b`)

func taskRelation(s string) string {
	l, a := launchMention.MatchString(s), arrivalMention.MatchString(s)
	if l == a {
		return ""
	}
	if l {
		return "launch"
	}
	return "arrival"
}

// This bounded task interpreter is an explicit research hypothesis, not a
// general semantic parser. It changes order only; unsupported structure carries
// no exclusion authority. Matching a target relation is separate from whether
// its date makes the proposition under investigation true.
func PlanTaskRole(ctx context.Context, req retrieval.RankRequest) (TaskPlan, error) {
	base, err := PlanPriority(ctx, req) // Reuse the frozen input/binding validation only.
	if err != nil {
		return TaskPlan{}, err
	}
	b := ctx.Value(bindingKey{}).(binding)
	q := strings.ToLower(strings.TrimSpace(b.original))
	p := TaskPlan{PriorityPlan: base, Role: "lookup", TargetRelation: taskRelation(q), Reasons: map[string]string{}}
	p.Method = "research/task-role-v1"
	if strings.HasPrefix(q, "which ") || strings.HasPrefix(q, "what publication status ") {
		p.Role = "selection"
	}
	if refutationMention.MatchString(q) || strings.HasPrefix(q, "was ") || strings.HasPrefix(q, "is ") || strings.HasPrefix(q, "did ") || strings.HasPrefix(q, "explain ") {
		p.Role = "assessment"
	}
	normalized := q
	for _, token := range []string{"did not occur", "isn't", "aren't", "isn’t", "aren’t"} {
		if strings.Contains(normalized, token) {
			p.NegatedSelection = true
			normalized = strings.ReplaceAll(normalized, token, "")
		}
	}
	rs := rules(normalized)
	if alternativeMention.MatchString(q) || (p.Role == "selection" && negation.MatchString(normalized)) {
		p.Role = "unsupported"
		p.TargetRelation = ""
		rs = nil
	}
	if p.NegatedSelection && len(rs) != 1 {
		p.Role = "unsupported"
		p.TargetRelation = ""
		rs = nil
	}
	good, bad := []int{}, []int{}
	for i, c := range req.Candidates {
		if err := ctx.Err(); err != nil {
			return TaskPlan{}, err
		}
		d := &p.Decisions[i]
		d.State = "unknown"
		d.CalendarDate = ""
		text, ok := what(c.Text)
		if !ok || p.Role == "unsupported" {
			good = append(good, i)
			continue
		}
		relation := taskRelation(text)
		if p.TargetRelation != "" && relation != "" {
			d.State = "compatible"
			p.Reasons[c.ID] = "target relation matches"
			if relation != p.TargetRelation {
				d.State = "contradicted"
				p.Reasons[c.ID] = "different target relation"
			}
		}
		if p.Role == "selection" && len(rs) > 0 {
			date, valid := Date(text)
			if valid {
				d.CalendarDate = date.Format("2006-01-02")
				matches := !contradict(rs, c.Text)
				if p.NegatedSelection {
					matches = !matches
				}
				if !matches {
					d.State = "contradicted"
					p.Reasons[c.ID] = "selection predicate fails"
				} else if d.State != "contradicted" {
					d.State = "compatible"
					p.Reasons[c.ID] = "selection predicate matches"
				}
			}
		}
		if d.State == "contradicted" {
			bad = append(bad, i)
		} else {
			good = append(good, i)
		}
	}
	p.AllContradicted = len(bad) == len(req.Candidates)
	p.Order = append(good, bad...)
	if p.AllContradicted {
		for i := range p.Order {
			p.Order[i] = i
		}
	}
	return p, nil
}
