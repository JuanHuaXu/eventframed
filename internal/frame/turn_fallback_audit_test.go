package frame

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

const turnFallbackAuditHead = "1a7edb62b6be4031fd01ebeab8071b17303a7815"
const turnFallbackAuditCurrentSHA256 = "88527b6032ebdd1f9d50102134a16fbcddabbf093e0aed708b5e76644fe51cb3"
const turnFallbackAuditQuerySHA256 = "34aec61ade88b154686e34efd99096fd4689d04f45d5d1b9de44120b2d4c315f"
const turnFallbackAuditIdentitySHA256 = "78e20396658bdb15940e62cece042af4334f4f74e85f497dbc0e483ed919689d"

func turnFallbackAuditCapture(user, assistant string) model.TurnCapture {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	return model.TurnCapture{
		ID: "synthetic-turn", TenantID: "synthetic-tenant", SessionID: "synthetic-session", Sequence: 1,
		RunID: "synthetic-run", UserID: "account:operator", AgentID: "audit-agent",
		UserText: user, AssistantText: assistant, RetrievedIDs: []string{"public-example-a", "public-example-b"},
		OccurredAt: now, ObservedAt: now.Add(time.Second), AvailableAt: now.Add(time.Second),
	}
}

func TestTurnFallbackAuditParticipantControls(t *testing.T) {
	for _, tc := range []struct {
		name, text, agent, value, token string
		confidence                      float64
	}{
		{"collective-we", "we plan a release.", "audit-agent", "unresolved group", "we", 0},
		{"collective-We-no-name-pattern", "We plan a release.", "audit-agent", "unresolved group", "We", 0},
		{"collective-us", "Please help us prepare.", "audit-agent", "unresolved group", "us", 0},
		{"collective-our", "Please review our plan.", "audit-agent", "unresolved group", "our", 0},
		{"collective-ours", "The plan remains ours.", "", "unresolved group", "ours", 0},
		{"collective-before-singular", "I plan a release for our team.", "audit-agent", "unresolved group", "our", 0},
		{"speaker-fallback", "I plan a release.", "audit-agent", "user", "I", .72},
		{"agent-fallback", "you can review the plan.", "audit-agent", "agent:audit-agent", "you", .70},
		{"no-agent-no-addressee-inference", "you can review the plan.", "", "user and agent", "", .55},
		{"participants-metadata", "Please proceed.", "audit-agent", "user and agent:audit-agent", "", .55},
		{"participants-no-agent", "Please proceed.", "", "user and agent", "", .55},
	} {
		t.Run(tc.name, func(t *testing.T) {
			turn := turnFallbackAuditCapture(tc.text, "Acknowledged.")
			turn.AgentID = tc.agent
			got := FromTurn(turn).Who
			evidence := "turn participants"
			if tc.token != "" {
				start := strings.Index(tc.text, tc.token)
				evidence = span("user", start, start+len(tc.token))
			}
			want := field(tc.value, model.SourceInferred, tc.confidence, evidence)
			if got != want {
				t.Fatalf("Who = %+v; want %+v", got, want)
			}
		})
	}
}

func TestTurnFallbackAuditIdentityControls(t *testing.T) {
	card := IdentityCard{UserID: "account:operator", Name: "Operator", Version: 1}
	previous := FromTurn(turnFallbackAuditCapture("Alex will deploy.", "Acknowledged."))
	for _, tc := range []struct{ text, who, status, method string }{
		{"I will deploy.", "Operator [user:account:operator]", "linked-card", "adapter-speaker"},
		{"I am Another Person.", "Operator [user:account:operator]", "linked-card", "adapter-speaker"},
		{"You will review.", "agent:audit-agent", "resolved-agent", "adapter-addressee"},
		{"We will deploy.", "unresolved participant", "ambiguous-reference", ""},
		{"we plan a release.", "unresolved participant", "ambiguous-reference", ""},
		{"We plan a release.", "unresolved participant", "ambiguous-reference", ""},
	} {
		t.Run(tc.text, func(t *testing.T) {
			turn := turnFallbackAuditCapture(tc.text, "Acknowledged.")
			turn.ParticipantUserIDs = []string{"account:other"}
			got := FromTurnWithIdentities(turn, []IdentityCard{card}, &previous, true)
			var resolution WhoResolution
			if err := json.Unmarshal([]byte(got.Attributes["who_resolution"]), &resolution); err != nil {
				t.Fatal(err)
			}
			if got.Who.Value != tc.who || resolution.Status != tc.status || resolution.Method != tc.method {
				t.Fatalf("Who = %+v; resolution = %+v", got.Who, resolution)
			}
			if tc.method == "" {
				var unresolved UnresolvedReferences
				if err := json.Unmarshal([]byte(got.Attributes["unresolved_references"]), &unresolved); err != nil {
					t.Fatal(err)
				}
				if got.Who.Confidence != 0 || resolution.UserID != "" || unresolved.Count == 0 || !strings.Contains(got.What.Value, tc.text) {
					t.Fatalf("group guessed or original reference lost: %+v / %+v", got, unresolved)
				}
			} else if got.Who.Source != model.SourceObserved || got.Who.Confidence != 1 {
				t.Fatalf("explicit adapter identity downgraded: %+v", got.Who)
			}
			if got.Attributes["speaker_user_id"] != turn.UserID || got.Attributes["speaker_agent_id"] != turn.AgentID {
				t.Fatal("speaker metadata lost")
			}
		})
	}
}

func turnFallbackAuditQuotes(text string) []struct{ name, text string } {
	return []struct{ name, text string }{
		{"double", `"` + text + `"`},
		{"single", "'" + text + "'"},
		{"curly-double", "\u201c" + text + "\u201d"},
		{"curly-single", "\u2018" + text + "\u2019"},
		{"inline-code", "`" + text + "`"},
		{"fenced-code", "```text\n" + text + "\n```"},
		{"blockquote", "> " + text + "\n"},
		{"indented-blockquote", " \t> " + text + "\n"},
	}
}

func TestTurnFallbackAuditQuotedClaims(t *testing.T) {
	claim := "Alex will deploy in Toronto tomorrow because cache failed using console."
	for _, quote := range turnFallbackAuditQuotes(claim) {
		for _, role := range []string{"user", "assistant"} {
			t.Run(quote.name+"/"+role, func(t *testing.T) {
				turn := turnFallbackAuditCapture("Please review.", "Acknowledged.")
				if role == "user" {
					turn.UserText = quote.text
				} else {
					turn.AssistantText = quote.text
				}
				got, prior := FromTurn(turn), turnFallbackAuditHeadFromTurn(turn)
				if got.Who != field("user and agent:audit-agent", model.SourceInferred, .55, "turn participants") ||
					got.Where.Evidence != "session metadata" || got.When.Evidence != "turn timestamp" ||
					got.Why.Source != model.SourceInferred || got.How.Source != model.SourceInferred {
					t.Fatalf("quoted claim selected as direct evidence: %+v", got)
				}
				if prior.Who.Value != "Alex" || prior.Where.Value != "Toronto" || prior.When.Value != "tomorrow" {
					t.Fatalf("prior control did not exercise the changed path: %+v", prior)
				}
				if got.Content != prior.Content || got.What != prior.What || got.Provenance.Producer != prior.Provenance.Producer {
					t.Fatal("raw metadata or first-statement behavior changed")
				}
				identity := FromTurnWithIdentities(turn, nil, nil, true)
				if identity.Who.Confidence != 0 || identity.Who.Value != "unresolved participant" || identity.Content != got.Content ||
					identity.Where != got.Where || identity.When != got.When || identity.Why != got.Why || identity.How != got.How {
					t.Fatalf("identity enrichment promoted quoted claims: %+v", identity)
				}
			})
		}
	}
}

func TestTurnFallbackAuditQuotedPronouns(t *testing.T) {
	for _, quote := range turnFallbackAuditQuotes("we you my") {
		t.Run(quote.name, func(t *testing.T) {
			turn := turnFallbackAuditCapture(quote.text, "Acknowledged.")
			got := FromTurn(turn)
			if got.Who != field("user and agent:audit-agent", model.SourceInferred, .55, "turn participants") {
				t.Fatalf("quoted pronoun affected fallback: %+v", got.Who)
			}
			identity := FromTurnWithIdentities(turn, nil, nil, true)
			var unresolved UnresolvedReferences
			if err := json.Unmarshal([]byte(identity.Attributes["unresolved_references"]), &unresolved); err != nil {
				t.Fatal(err)
			}
			if identity.Who.Confidence != 0 || unresolved.Count != 0 || identity.Content != got.Content {
				t.Fatalf("quoted references resolved or flagged: %+v / %+v", identity.Who, unresolved)
			}
			turn.UserText = quote.text + "\nI plan a release."
			if got := FromTurn(turn).Who; got.Value != "user" || got.Confidence != .72 {
				t.Fatalf("quoted collective hid live singular fallback: %+v", got)
			}
		})
	}
}

func TestTurnFallbackAuditUnquotedEvidenceAfterQuotes(t *testing.T) {
	turn := turnFallbackAuditCapture("\u201cBlair will deploy in Paris today.\u201d Alex will deploy in Toronto tomorrow because cache failed using console.", "Acknowledged.")
	got := FromTurn(turn)
	for _, tc := range []struct {
		field model.Field
		value string
	}{{got.Who, "Alex"}, {got.Where, "Toronto"}, {got.When, "tomorrow"}, {got.Why, "cache failed"}, {got.How, "using console"}} {
		start := strings.Index(turn.UserText, tc.value)
		if tc.field.Value != tc.value || tc.field.Source != model.SourceObserved || tc.field.Evidence != span("user", start, start+len(tc.value)) {
			t.Fatalf("byte-aligned unquoted field = %+v; want %q", tc.field, tc.value)
		}
	}
	turn.UserText, turn.AssistantText = "Please review.", turn.UserText
	got = FromTurn(turn)
	if got.Who.Value != "Alex" || got.Who.Source != model.SourceSynthetic || got.Who.Confidence != .76 || !strings.HasPrefix(got.Who.Evidence, "assistant[bytes:") {
		t.Fatalf("assistant evidence/provenance changed: %+v", got.Who)
	}
}

func TestTurnFallbackAuditQuoteMaskAlignment(t *testing.T) {
	for _, quote := range turnFallbackAuditQuotes("\u00e9 we in Toronto tomorrow") {
		t.Run(quote.name, func(t *testing.T) {
			text := quote.text + "\nI plan a release."
			mask := unquoted(text)
			if len(mask) != len(text) || !strings.HasSuffix(mask, "\nI plan a release.") {
				t.Fatalf("mask changed byte positions/live text: %q", mask)
			}
			for i := 0; i < len(quote.text); i++ {
				want := byte(' ')
				if quote.text[i] == '\n' {
					want = '\n'
				}
				if mask[i] != want {
					t.Fatalf("quote byte %d survived: %q", i, mask)
				}
			}
		})
	}
	for _, text := range []string{"I'm ready.", "O'Brien will deploy.", "I plan a release."} {
		if mask := unquoted(text); mask != text {
			t.Fatalf("unquoted contraction/name masked: %q -> %q", text, mask)
		}
	}
}

// Desired-contract regressions stay active after the approved repair.
func TestTurnFallbackAuditKnownBugRegressions(t *testing.T) {
	t.Run("capitalized-collective-bypasses-fallback", func(t *testing.T) {
		turn := turnFallbackAuditCapture("We will deploy.", "Acknowledged.")
		got, prior := FromTurn(turn), turnFallbackAuditHeadFromTurn(turn)
		t.Logf("prior Who=%+v; current Who=%+v", prior.Who, got.Who)
		if got.Who.Confidence != 0 || got.Who.Value != "unresolved group" {
			t.Errorf("ambiguous collective selected as a named actor: %+v", got.Who)
		}
	})
	for _, tc := range []struct{ name, text, fieldName string }{
		{"quoted-only-reason", `Please deploy because "cache failed".`, "why"},
		{"quoted-only-method", "Please deploy using `console`.", "how"},
		{"quoted-only-temporal-clause", `Please deploy after "approval".`, "when"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, role := range []string{"user", "assistant"} {
				t.Run(role, func(t *testing.T) {
					turn := turnFallbackAuditCapture("Please review.", "Acknowledged.")
					if role == "user" {
						turn.UserText = tc.text
					} else {
						turn.AssistantText = tc.text
					}
					prior := turnFallbackAuditHeadFromTurn(turn)
					t.Logf("prior Why=%+v; How=%+v; When=%+v", prior.Why, prior.How, prior.When)
					for _, identity := range []bool{false, true} {
						got := FromTurn(turn)
						if identity {
							got = FromTurnWithIdentities(turn, nil, nil, true)
						}
						f := map[string]model.Field{"why": got.Why, "how": got.How, "when": got.When}[tc.fieldName]
						t.Logf("identity=%t; field=%+v; mask=%q", identity, f, unquoted(tc.text))
						isFallback := f.Source == model.SourceInferred
						if tc.fieldName == "when" {
							isFallback = f.Evidence == "turn timestamp"
						}
						if !isFallback {
							t.Errorf("identity=%t: quoted-only clause promoted to direct %s: %+v", identity, tc.fieldName, f)
						}
					}
				})
			}
		})
	}
}

func TestTurnFallbackAuditMatchedControls(t *testing.T) {
	for _, profile := range []string{"early", "late", "no-match", "collective"} {
		for _, size := range []int{256, 2048, 16384} {
			t.Run(fmt.Sprintf("%s/%d", profile, size), func(t *testing.T) {
				turn := turnFallbackAuditFixture(profile, size)
				if err := turn.Validate(); err != nil {
					t.Fatal(err)
				}
				got, prior := FromTurn(turn), turnFallbackAuditHeadFromTurn(turn)
				if profile == "collective" {
					if prior.Who.Value != "user and agent" || prior.Who.Confidence != .72 || got.Who.Value != "unresolved group" || got.Who.Confidence != 0 {
						t.Fatalf("collective treatment not exercised: prior=%+v; current=%+v", prior.Who, got.Who)
					}
					prior.Who = got.Who
				}
				if !reflect.DeepEqual(got, prior) {
					t.Fatalf("unexpected full-event difference:\nprior=%+v\ncurrent=%+v", prior, got)
				}
			})
		}
	}
}

func turnFallbackAuditQueryFields(text string) map[string]string {
	fields := make(map[string]string)
	for _, line := range strings.Split(QueryText(text), "\n") {
		name, value, ok := strings.Cut(line, ": ")
		if ok {
			fields[name] = value
		}
	}
	return fields
}

func turnFallbackAuditCheckDirectSpans(t *testing.T, event model.Event, sources map[string]string) {
	t.Helper()
	for _, f := range []model.Field{event.Who, event.Where, event.When, event.Why, event.How} {
		if f.Source != model.SourceObserved && f.Source != model.SourceSynthetic {
			continue // Inferred summaries retain raw text; they are not promoted captures.
		}
		role, _, ok := strings.Cut(f.Evidence, "[bytes:")
		if !ok {
			continue
		}
		raw, ok := sources[role]
		if !ok {
			t.Fatalf("unexpected evidence role: %+v", f)
		}
		var start, end int
		if _, err := fmt.Sscanf(f.Evidence, role+"[bytes:%d:%d]", &start, &end); err != nil || start < 0 || end < start || end > len(raw) {
			t.Fatalf("invalid byte evidence: %+v", f)
		}
		if raw[start:end] != unquoted(raw)[start:end] {
			t.Errorf("promoted capture contains masked bytes: %+v", f)
		}
	}
}

func TestTurnFallbackAuditMaskedSpanAdmission(t *testing.T) {
	for _, tc := range []struct{ text, blocked string }{
		{`Alex "quoted" Blair will deploy.`, "who"},
		{`Please deploy in north "quoted" zone tomorrow.`, "where"},
		{`Please deploy next "quoted" week.`, "when"},
		{`Please deploy after "approval".`, "when"},
		{`Please deploy because "cache failed".`, "why"},
		{`Please deploy because cache "quoted" failed.`, "why"},
		{`Please deploy using console "quoted" and script.`, "how"},
		{"Please deploy using `console`.", "how"},
		{"Please deploy using console\n> they need me\nand script.", "how"},
	} {
		t.Run(tc.text, func(t *testing.T) {
			turn := turnFallbackAuditCapture(tc.text, "Acknowledged.")
			for _, role := range []string{"user", "assistant"} {
				t.Run(role, func(t *testing.T) {
					input := turn
					if role == "assistant" {
						input.UserText, input.AssistantText = "Please review.", tc.text
					}
					got, prior := FromTurn(input), turnFallbackAuditHeadFromTurn(input)
					turnFallbackAuditCheckDirectSpans(t, got, map[string]string{"user": input.UserText, "assistant": input.AssistantText})
					if got.Content != prior.Content || got.What != prior.What {
						t.Fatal("mask admission changed raw metadata/What")
					}
				})
			}
			query := turnFallbackAuditQueryFields(tc.text)
			if _, ok := query[tc.blocked]; ok {
				t.Errorf("masked span became query %s: %q", tc.blocked, query[tc.blocked])
			}
			if query["what"] != "request: "+firstStatement(sourceText{text: tc.text}).value {
				t.Fatal("query What lost retained raw text")
			}
			if tc.blocked == "who" {
				return // Offline Who is a supplied role, not a text-derived actor.
			}
			offline := FromText(tc.text, "offline-role", turn.SessionID, turn.OccurredAt, model.SourceObserved)
			turnFallbackAuditCheckDirectSpans(t, offline, map[string]string{"message": tc.text})
			f := map[string]model.Field{"where": offline.Where, "when": offline.When, "why": offline.Why, "how": offline.How}[tc.blocked]
			if (tc.blocked == "where" && f.Evidence != "session metadata") || (tc.blocked == "when" && f.Evidence != "message timestamp") ||
				((tc.blocked == "why" || tc.blocked == "how") && f != (model.Field{})) {
				t.Errorf("masked offline clause did not abstain: %+v", f)
			}
			if offline.Content != tc.text || offline.What.Value != firstStatement(sourceText{text: tc.text}).value || offline.Who.Value != "offline-role" {
				t.Fatal("offline raw text/What/role changed")
			}
		})
	}
	for _, quote := range turnFallbackAuditQuotes("untrusted payload") {
		t.Run("quote-form/"+quote.name, func(t *testing.T) {
			text := "Please deploy using\n" + quote.text + "\n."
			turn := turnFallbackAuditCapture(text, "Acknowledged.")
			got := FromTurn(turn)
			turnFallbackAuditCheckDirectSpans(t, got, map[string]string{"user": text, "assistant": turn.AssistantText})
			if got.How.Source != model.SourceInferred || turnFallbackAuditQueryFields(text)["how"] != "" ||
				FromText(text, "offline-role", turn.SessionID, turn.OccurredAt, model.SourceObserved).How != (model.Field{}) {
				t.Fatal("masked-only method was admitted")
			}
		})
	}
}

func TestTurnFallbackAuditNamedActorAdmission(t *testing.T) {
	for _, actor := range []string{"I", "Me", "My", "Mine", "You", "Your", "Yours", "He", "Him", "His", "She", "Her", "Hers", "It", "Its", "They", "Them", "Their", "Theirs", "We", "Us", "Our", "Ours", "Ourselves", "Myself", "Yourself", "Yourselves", "Himself", "Herself", "Itself", "Themselves", "We All", "Our Team", "You And I", "All Of Us", "Who", "Whom", "Whose", "This", "That", "These", "Those", "Anyone", "Anybody", "Anything", "Everyone", "Everybody", "Everything", "Someone", "Somebody", "Something", "Nobody", "Nothing", "None"} {
		t.Run(actor, func(t *testing.T) {
			text := actor + " will deploy."
			turn := turnFallbackAuditCapture(text, "Acknowledged.")
			got := FromTurn(turn)
			if got.Who != participantFallback(turn, sourceText{name: "user", text: text}) {
				t.Errorf("pronoun/collective capture became a named actor: %+v", got.Who)
			}
			if who := turnFallbackAuditQueryFields(text)["who"]; who != "" {
				t.Errorf("pronoun/collective query actor = %q", who)
			}
			turn.UserText, turn.AssistantText = "Please review.", text
			if who := FromTurn(turn).Who; who.Source == model.SourceSynthetic {
				t.Errorf("assistant pronoun capture became a named actor: %+v", who)
			}
		})
	}
	for _, actor := range []string{"Alex", "Example Operator", "Operations Team", "O'Brien", "Jean-Luc", "Iris", "Usain", "Theodore", "I\u00e9", "Alex \u00c9lodie"} {
		t.Run("genuine/"+actor, func(t *testing.T) {
			text := actor + " will deploy."
			turn := turnFallbackAuditCapture(text, "Acknowledged.")
			got := FromTurn(turn)
			if got.Who != field(actor, model.SourceObserved, .88, span("user", 0, len(actor))) || turnFallbackAuditQueryFields(text)["who"] != actor {
				t.Fatalf("genuine name rejected: %+v / %s", got.Who, QueryText(text))
			}
			turn.UserText, turn.AssistantText = "Please review.", text
			if who := FromTurn(turn).Who; who.Value != actor || who.Source != model.SourceSynthetic || who.Confidence != .76 {
				t.Errorf("genuine assistant name rejected: %+v", who)
			}
		})
	}
}

func TestTurnFallbackAuditAdjacentUnquotedControls(t *testing.T) {
	const text = "Alex will deploy in Toronto tomorrow because cache failed using console."
	turn := turnFallbackAuditCapture(text, "Acknowledged.")
	query := turnFallbackAuditQueryFields(text)
	offline := FromText(text, "offline-role", turn.SessionID, turn.OccurredAt, model.SourceObserved)
	for _, tc := range []struct {
		name, value string
		f           model.Field
	}{{"where", "Toronto", offline.Where}, {"when", "tomorrow", offline.When}, {"why", "cache failed", offline.Why}, {"how", "using console", offline.How}} {
		start := strings.Index(text, tc.value)
		if query[tc.name] != tc.value || tc.f.Value != tc.value || tc.f.Source != model.SourceObserved || tc.f.Evidence != span("message", start, start+len(tc.value)) {
			t.Errorf("adjacent %s control lost: query=%q / offline=%+v", tc.name, query[tc.name], tc.f)
		}
	}
	for _, text := range []string{`Please deploy because "untrusted" cache failed.`, "Please deploy because cache failed \t ."} {
		if why := FromTurn(turnFallbackAuditCapture(text, "Acknowledged.")).Why; why.Value != "cache failed" || why.Source != model.SourceObserved {
			t.Errorf("untouched capture was over-rejected: %+v", why)
		}
	}
	turn.UserText, turn.AssistantText = `Please deploy because "untrusted".`, "Blair reported progress because tests passed."
	if why := FromTurn(turn).Why; why.Value != "tests passed" || why.Source != model.SourceSynthetic || why.Confidence != .78 {
		t.Errorf("rejected user capture blocked valid assistant evidence: %+v", why)
	}
}

func TestTurnFallbackAuditContractedActors(t *testing.T) {
	for _, actor := range []string{"I'm", "I've", "I'll", "I'd", "We're", "We've", "We'll", "We'd", "You're", "You've", "You'll", "You'd", "He's", "He'll", "He'd", "She's", "She'll", "She'd", "It's", "It'll", "It'd", "They're", "They've", "They'll", "They'd"} {
		t.Run(actor, func(t *testing.T) {
			text := actor + " deployed."
			turn := turnFallbackAuditCapture(text, "Acknowledged.")
			if who := FromTurn(turn).Who; who != participantFallback(turn, sourceText{name: "user", text: text}) || turnFallbackAuditQueryFields(text)["who"] != "" {
				t.Fatalf("contracted pronoun became a named actor: %+v / %s", who, QueryText(text))
			}
		})
	}
}

func TestTurnFallbackAuditBaselineFidelity(t *testing.T) {
	turnFallbackAuditCheckBaseline(t)
}

// AST comparisons check complete copied bodies, not just selected fields. Only
// unchanged declarations are compared with HEAD; the final candidate is pinned.
// This prevents a later edit from silently making the benchmark unmatched.
func turnFallbackAuditCheckBaseline(t testing.TB) {
	t.Helper()
	_, testPath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate audit source")
	}
	dir := filepath.Dir(testPath)
	root := filepath.Clean(filepath.Join(dir, "../.."))
	head, err := exec.Command("git", "-C", root, "show", turnFallbackAuditHead+":internal/frame/turn.go").Output()
	if err != nil {
		t.Fatal(err)
	}
	revision, err := exec.Command("git", "-C", root, "rev-parse", "--verify", turnFallbackAuditHead+"^{commit}").Output()
	if err != nil || strings.TrimSpace(string(revision)) != turnFallbackAuditHead {
		t.Fatalf("immutable audit control commit unavailable/mismatched: %s / %v", revision, err)
	}
	objectType, err := exec.Command("git", "-C", root, "cat-file", "-t", turnFallbackAuditHead).Output()
	if err != nil || strings.TrimSpace(string(objectType)) != "commit" {
		t.Fatalf("immutable audit control has wrong object type: %s / %v", objectType, err)
	}
	current, err := os.ReadFile(filepath.Join(dir, "turn.go"))
	if err != nil {
		t.Fatal(err)
	}
	if digest := fmt.Sprintf("%x", sha256.Sum256(current)); digest != turnFallbackAuditCurrentSHA256 {
		t.Fatalf("audit patch changed: current turn.go SHA256=%s", digest)
	}
	for path, expected := range map[string]string{"query.go": turnFallbackAuditQuerySHA256, "identity.go": turnFallbackAuditIdentitySHA256} {
		data, err := os.ReadFile(filepath.Join(dir, path))
		if err != nil {
			t.Fatal(err)
		}
		if digest := fmt.Sprintf("%x", sha256.Sum256(data)); digest != expected {
			t.Fatalf("audit candidate/quote implementation changed: %s SHA256=%s", path, digest)
		}
	}
	readAST := func(path string, source []byte) *ast.File {
		f, err := parser.ParseFile(token.NewFileSet(), path, source, 0)
		if err != nil {
			t.Fatal(err)
		}
		return f
	}
	headAST := readAST("HEAD-turn.go", head)
	currentAST := readAST("turn.go", current)
	testSource, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatal(err)
	}
	testAST := readAST(testPath, testSource)
	renames := map[string]string{
		"FromTurn":            "turnFallbackAuditHeadFromTurn",
		"firstField":          "turnFallbackAuditHeadFirstField",
		"participantFallback": "turnFallbackAuditHeadParticipantFallback",
	}
	canonical := func(node ast.Node) string {
		var out bytes.Buffer
		if err := format.Node(&out, token.NewFileSet(), node); err != nil {
			t.Fatal(err)
		}
		return out.String()
	}
	var sharedHead, sharedCurrent []string
	for _, decl := range headAST.Decls {
		fn, isFunc := decl.(*ast.FuncDecl)
		if !isFunc || renames[fn.Name.Name] == "" {
			sharedHead = append(sharedHead, canonical(decl))
			continue
		}
		name := renames[fn.Name.Name]
		ast.Inspect(fn, func(node ast.Node) bool {
			if id, ok := node.(*ast.Ident); ok && renames[id.Name] != "" {
				id.Name = renames[id.Name]
			}
			return true
		})
		found := false
		for _, candidate := range testAST.Decls {
			copy, ok := candidate.(*ast.FuncDecl)
			if ok && copy.Name.Name == name {
				found = true
				if canonical(fn) != canonical(copy) {
					t.Fatalf("reconstructed %s differs from complete HEAD body", name)
				}
			}
		}
		if !found {
			t.Fatalf("missing reconstruction %s", name)
		}
	}
	for _, decl := range currentAST.Decls {
		fn, isFunc := decl.(*ast.FuncDecl)
		if !isFunc || (renames[fn.Name.Name] == "" && fn.Name.Name != "validNamedActor") {
			sharedCurrent = append(sharedCurrent, canonical(decl))
		}
	}
	if !reflect.DeepEqual(sharedHead, sharedCurrent) {
		t.Fatal("shared extraction helpers/patterns differ from HEAD")
	}
}

func turnFallbackAuditFixture(profile string, size int) model.TurnCapture {
	const claim = "Alex will deploy in Toronto tomorrow because cache failed using console.\n"
	const assistantClaim = "Blair reported progress in Ottawa today because tests passed using script.\n"
	const filler = "synthetic fixture sample alpha beta gamma delta epsilon.\n"
	pad := func(prefix, suffix string) string {
		n := size - len(prefix) - len(suffix)
		if n < 0 {
			panic("audit fixture too small")
		}
		return prefix + strings.Repeat(filler, (n+len(filler)-1)/len(filler))[:n] + suffix
	}
	var user, assistant string
	switch profile {
	case "early":
		user, assistant = pad(claim, ""), pad(assistantClaim, "")
	case "late":
		user, assistant = pad("", "\n"+claim), pad("", "\n"+assistantClaim)
	case "no-match":
		user, assistant = pad("Please review.\n", ""), pad("Acknowledged.\n", "")
	case "quoted":
		user = pad("Please review.\n", "\n\""+strings.TrimSpace(claim)+"\"")
		assistant = pad("Acknowledged.\n", "\n\""+strings.TrimSpace(assistantClaim)+"\"")
	case "collective":
		user, assistant = pad("we plan a release.\n", ""), pad("Acknowledged.\n", "")
	default:
		panic("unknown audit fixture profile")
	}
	return turnFallbackAuditCapture(user, assistant)
}

var turnFallbackAuditSink model.Event

func BenchmarkTurnFallbackAuditMatched(b *testing.B) {
	turnFallbackAuditBenchmark(b, false, 1)
}

func BenchmarkTurnFallbackAuditMatchedReverse(b *testing.B) {
	turnFallbackAuditBenchmark(b, true, 1)
}

func BenchmarkTurnFallbackAuditMatchedAlternating(b *testing.B) {
	turnFallbackAuditBenchmark(b, false, 3)
}

func turnFallbackAuditBenchmark(b *testing.B, reverse bool, rounds int) {
	turnFallbackAuditCheckBaseline(b)
	versions := []struct {
		name    string
		extract func(model.TurnCapture) model.Event
	}{{"HEAD", turnFallbackAuditHeadFromTurn}, {"current", FromTurn}}
	if reverse {
		versions[0], versions[1] = versions[1], versions[0]
	}
	for _, profile := range []string{"early", "late", "no-match", "quoted", "collective"} {
		for _, size := range []int{256, 2048, 16384} {
			turn := turnFallbackAuditFixture(profile, size)
			if err := turn.Validate(); err != nil {
				b.Fatal(err)
			}
			for round := 0; round < rounds; round++ {
				order := [2]int{0, 1}
				if round%2 == 1 {
					order[0], order[1] = order[1], order[0]
				}
				for _, index := range order {
					version := versions[index]
					name := fmt.Sprintf("%s/%dB-per-role/%s", profile, size, version.name)
					if rounds > 1 {
						name = fmt.Sprintf("%s/%dB-per-role/round-%d/%s", profile, size, round+1, version.name)
					}
					b.Run(name, func(b *testing.B) {
						b.ReportAllocs()
						b.SetBytes(int64(len(turn.UserText) + len(turn.AssistantText)))
						b.ResetTimer()
						for b.Loop() {
							turnFallbackAuditSink = version.extract(turn)
						}
					})
				}
			}
		}
	}
}

// These three functions are HEAD's complete implementations, with only their
// names and calls to each other changed. In particular HEAD firstField does NOT
// use unquoted; giving it the new mask would remove part of the actual patch.
func turnFallbackAuditHeadFromTurn(turn model.TurnCapture) model.Event {
	user := sourceText{name: "user", text: turn.UserText, source: model.SourceObserved}
	assistant := sourceText{name: "assistant", text: turn.AssistantText, source: model.SourceSynthetic}
	sources := []sourceText{user, assistant}
	request := firstStatement(user)
	outcome := firstStatement(assistant)
	return model.Event{
		ID: turn.ID, TenantID: turn.TenantID, SessionID: turn.SessionID, Sequence: turn.Sequence,
		Kind: "agent_turn", Content: "User: " + turn.UserText + "\n\nAssistant: " + turn.AssistantText,
		OccurredAt: turn.OccurredAt, ObservedAt: turn.ObservedAt, AvailableAt: turn.AvailableAt,
		Who:      turnFallbackAuditHeadFirstField(sources, whoPatterns, .88, nil, turnFallbackAuditHeadParticipantFallback(turn, user)),
		What:     field("request: "+request.value+"; outcome: "+outcome.value, model.SourceInferred, .82, request.evidence+"; "+outcome.evidence),
		Where:    turnFallbackAuditHeadFirstField(sources, wherePatterns, .86, validLocation, field("session:"+turn.SessionID, model.SourceObserved, 1, "session metadata")),
		When:     turnFallbackAuditHeadFirstField(sources, whenPatterns, .90, nil, field(turn.OccurredAt.Format("2006-01-02T15:04:05.999999999Z07:00"), model.SourceObserved, 1, "turn timestamp")),
		Why:      turnFallbackAuditHeadFirstField(sources, whyPatterns, .90, nil, field("to address the user request: "+request.value, model.SourceInferred, .62, request.evidence)),
		How:      turnFallbackAuditHeadFirstField(sources, howPatterns, .86, nil, field("through the agent response: "+outcome.value, model.SourceInferred, .62, outcome.evidence)),
		Priority: .5, Tags: []string{"conversation", "agent-turn"},
		Provenance: model.Provenance{Producer: "openclaw-eventframe-memory", RetrievedIDs: append([]string(nil), turn.RetrievedIDs...), RunID: turn.RunID},
		Attributes: map[string]string{"user_source": "observed", "assistant_source": "synthetic", "semantic_extractor": "fivew1h-deterministic-v1"},
	}
}

func turnFallbackAuditHeadFirstField(sources []sourceText, patterns []pattern, confidence float64, accept func(string) bool, fallback model.Field) model.Field {
	for _, source := range sources {
		for _, candidate := range patterns {
			indices := candidate.expression.FindStringSubmatchIndex(source.text)
			group := candidate.group * 2
			if len(indices) <= group+1 || indices[group] < 0 {
				continue
			}
			start, end := indices[group], indices[group+1]
			value := strings.TrimSpace(source.text[start:end])
			if value == "" || (accept != nil && !accept(value)) {
				continue
			}
			if source.source == model.SourceSynthetic {
				confidence = max(0, confidence-.12)
			}
			return field(value, source.source, confidence, span(source.name, start, end))
		}
	}
	return fallback
}

func turnFallbackAuditHeadParticipantFallback(turn model.TurnCapture, user sourceText) model.Field {
	if indices := collectivePattern.FindStringIndex(user.text); indices != nil {
		return field("user and agent", model.SourceInferred, .72, span(user.name, indices[0], indices[1]))
	}
	if indices := firstPersonPattern.FindStringIndex(user.text); indices != nil {
		return field("user", model.SourceInferred, .72, span(user.name, indices[0], indices[1]))
	}
	if turn.AgentID != "" {
		if indices := secondPersonPattern.FindStringIndex(user.text); indices != nil {
			return field("agent:"+turn.AgentID, model.SourceInferred, .70, span(user.name, indices[0], indices[1]))
		}
		return field("user and agent:"+turn.AgentID, model.SourceInferred, .55, "turn participants")
	}
	return field("user and agent", model.SourceInferred, .55, "turn participants")
}
