package frame

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

// IdentityCard is a versioned lookup result, not evidence for the truth of a
// statement. Opaque profile prose is never mined for identity aliases.
type IdentityCard struct {
	UserID    string
	Name      string
	Aliases   []string
	Version   int32
	UpdatedAt time.Time
	Hash      string
}

type identityNames struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	Aliases     []string `json:"aliases"`
	Accounts    []struct {
		Platform string `json:"platform"`
		ID       string `json:"id"`
	} `json:"accounts"`
}

func ParseIdentityCard(userID, payload string, version int32, updatedAt time.Time) (IdentityCard, error) {
	var raw map[string]json.RawMessage
	if len(payload) > 256<<10 || json.Unmarshal([]byte(payload), &raw) != nil || raw == nil {
		return IdentityCard{}, fmt.Errorf("invalid user-card object")
	}
	card := IdentityCard{UserID: userID, Version: version, UpdatedAt: updatedAt}
	digest := sha256.Sum256([]byte(payload))
	card.Hash = hex.EncodeToString(digest[:])
	add := func(names identityNames) {
		if card.Name == "" {
			card.Name = bounded(names.Name, 128)
		}
		if card.Name == "" {
			card.Name = bounded(names.DisplayName, 128)
		}
		for _, alias := range append([]string{names.Name, names.DisplayName}, names.Aliases...) {
			if alias = strings.TrimSpace(alias); alias != "" && len(alias) <= 128 && len(card.Aliases) < 32 {
				card.Aliases = append(card.Aliases, alias)
			}
		}
		for _, account := range names.Accounts {
			// Bare account IDs are not globally unique; require a platform prefix.
			if account.Platform != "" && account.ID != "" && len(account.Platform)+len(account.ID) <= 127 && len(card.Aliases) < 32 {
				card.Aliases = append(card.Aliases, account.Platform+":"+account.ID)
			}
		}
	}
	var names identityNames
	if json.Unmarshal([]byte(payload), &names) == nil {
		add(names)
	}
	var identities []json.RawMessage
	if json.Unmarshal(raw["identity"], &identities) == nil {
		for _, identity := range identities[:min(len(identities), 8)] {
			var names identityNames
			if json.Unmarshal(identity, &names) == nil {
				add(names)
			}
		}
	}
	return card, nil
}

type WhoResolution struct {
	Status         string    `json:"status"`
	Method         string    `json:"method,omitempty"`
	Mention        string    `json:"mention,omitempty"`
	UserID         string    `json:"user_card_user_id,omitempty"`
	CardVersion    int32     `json:"card_version,omitempty"`
	CardHash       string    `json:"card_hash,omitempty"`
	CardUpdatedAt  time.Time `json:"card_updated_at,omitempty"`
	ContextEventID string    `json:"context_event_id,omitempty"`
}

// UnresolvedReference identifies a retained token, not a guessed referent.
// Fields name bounded extraction inputs; mentions are not raw-text offsets.
type UnresolvedReference struct {
	Field   string `json:"field"`
	Role    string `json:"role"`
	Mention string `json:"mention"`
	Reason  string `json:"reason"`
}

type UnresolvedReferences struct {
	Status    string                `json:"status"`
	Scope     string                `json:"scope"`
	Count     int                   `json:"count"`
	Truncated bool                  `json:"truncated"`
	Entries   []UnresolvedReference `json:"entries"`
}

func (r *UnresolvedReferences) record(field, role, mention, reason string) {
	r.Count++
	r.Status = "needs-resolution"
	if len(r.Entries) >= 64 {
		r.Truncated = true
		return
	}
	r.Entries = append(r.Entries, UnresolvedReference{field, role, mention, reason})
}

var subjectName = regexp.MustCompile(`\b(\p{Lu}[\p{L}'-]*(?:\s+\p{Lu}[\p{L}'-]*){0,3})\s+(?:will|should|must|needs?|wants?|decided|reported|said|owns|created|implemented|deployed|joined|left|is|was)\b`)
var namedMention = regexp.MustCompile(`\b\p{Lu}[\p{L}'-]*(?:\s+\p{Lu}[\p{L}'-]*){0,3}\b`)
var subjectPronoun = regexp.MustCompile(`(?i)^\s*(I(?:['\x{2019}](?:m|ve|ll|d))?|you(?:['\x{2019}](?:re|ve|ll|d))?|he|she|they|it|we)\b`)
var referenceWords = regexp.MustCompile(`\b(?i:I am|I(?:['\x{2019}](?:m|ve|ll|d))?|me|my|mine|you are|you(?:['\x{2019}](?:re|ve|ll|d))?|your|yours|he|him|his|she|her|hers|we|us|our|ours|they|them|their|theirs|it|its|myself|yourself|yourselves|himself|herself|itself|ourselves|themselves)\b`)
var referenceURL = regexp.MustCompile("(?i)\\b[a-z][a-z0-9+.-]*://[^\\s<>\"`]+")
var subjectAction = regexp.MustCompile(`(?i)^(?:will|should|must|needs?|wants?|decided|reported|said|owns|created|implemented|deployed|joined|left|is|was)\b`)

// FromTurnWithIdentities resolves only a small, explicit identity roster and a
// validated preceding turn. Missing, conflicting, quoted, or plural references
// abstain. It never edits the raw turn or rewrites a user card.
func FromTurnWithIdentities(turn model.TurnCapture, cards []IdentityCard, previous *model.Event, aliasesComplete bool) model.Event {
	event := FromTurn(turn)
	event.Attributes["semantic_extractor"] = "fivew1h-identity-v2"
	if turn.UserID != "" {
		event.Attributes["speaker_user_id"] = turn.UserID
	}
	if turn.AgentID != "" {
		event.Attributes["speaker_agent_id"] = turn.AgentID
	}
	resolution := WhoResolution{Status: "unresolved"}
	clean := unquoted(turn.UserText)
	var focus string
	var contextID string
	// A previous extracted subject is used only if it names one explicit subject
	// in user evidence. Generated assistant text is not an identity authority.
	if previous != nil {
		previousUser := previous.Content
		if strings.HasPrefix(previousUser, "User: ") {
			previousUser = strings.SplitN(strings.TrimPrefix(previousUser, "User: "), "\n\nAssistant: ", 2)[0]
			names := explicitSubjects(unquoted(previousUser))
			if len(names) == 1 {
				focus, contextID = names[0], previous.ID
			}
		}
	}
	nameMatches := subjectName.FindAllStringSubmatchIndex(clean, -1)
	event.Who = field("unresolved participant", model.SourceInferred, 0, "no unambiguous subject")
	for _, m := range nameMatches {
		if nonName(clean[m[2]:m[3]]) {
			continue
		}
		event.Who = field(turn.UserText[m[2]:m[3]], model.SourceObserved, .88, span("user", m[2], m[3]))
		resolution.Mention = event.Who.Value
		resolution.Method = "explicit-subject"
		break
	}
	// Structured aliases also cover lower-case names and platform-qualified
	// handles, which the capitalized-name extractor cannot recognize.
	for _, card := range cards {
		if !aliasesComplete {
			break
		}
		for _, alias := range append([]string{card.UserID}, card.Aliases...) {
			trimmed := strings.TrimLeft(clean, " \t\r\n")
			if len(trimmed) <= len(alias) || !aliasEqual(trimmed[:len(alias)], alias) || trimmed[len(alias)] != ' ' {
				continue
			}
			if !subjectAction.MatchString(strings.TrimSpace(trimmed[len(alias):])) {
				continue
			}
			event.Who = field(alias, model.SourceObserved, .88, span("user", len(clean)-len(trimmed), len(clean)-len(trimmed)+len(alias)))
			resolution.Mention, resolution.Method = alias, "explicit-subject"
			break
		}
	}
	if m := subjectPronoun.FindStringSubmatchIndex(clean); m != nil && referenceBoundary(clean, m[2], m[3]) {
		mention := clean[m[2]:m[3]]
		resolution = WhoResolution{Status: "unresolved", Mention: mention}
		event.Who = field("unresolved participant", model.SourceInferred, 0, span("user", m[2], m[3]))
		switch strings.ToLower(strings.ReplaceAll(mention, "\u2019", "'")) {
		case "i", "i'm", "i've", "i'll", "i'd":
			if turn.UserID != "" {
				event.Who = field(actorLabel("", "user", turn.TenantID, turn.UserID), model.SourceObserved, 1, "adapter user_id")
				resolution.UserID, resolution.Method, resolution.Status = turn.UserID, "adapter-speaker", "resolved-account"
			}
		case "you", "you're", "you've", "you'll", "you'd":
			if turn.AgentID != "" {
				event.Who = field(actorLabel("", "agent", turn.TenantID, turn.AgentID), model.SourceObserved, 1, "adapter agent_id")
				resolution.Method, resolution.Status = "adapter-addressee", "resolved-agent"
			}
		case "he", "she":
			if focus != "" {
				event.Who = field(focus, model.SourceInferred, .75, span("user", m[2], m[3]))
				resolution.Method, resolution.ContextEventID = "preceding-explicit-subject", contextID
			}
		default: // "we", "they", and "it" need a group/object referent, not a guessed person.
			resolution.Status = "ambiguous-reference"
		}
	}
	if resolution.Method == "explicit-subject" || resolution.Method == "preceding-explicit-subject" {
		var matches []IdentityCard
		if aliasesComplete {
			matches = matchingCards(event.Who.Value, cards)
		} else {
			resolution.Status = "alias-roster-incomplete"
		}
		if len(matches) == 1 {
			resolution.UserID = matches[0].UserID
		}
		if len(matches) > 1 {
			resolution.Status = "ambiguous-alias"
		}
	}
	for _, card := range cards {
		if card.UserID != resolution.UserID {
			continue
		}
		resolution.Status = "linked-card"
		resolution.CardVersion, resolution.CardHash, resolution.CardUpdatedAt = card.Version, card.Hash, card.UpdatedAt
		event.Who.Value = actorLabel(card.Name, "user", turn.TenantID, card.UserID)
		break
	}
	if resolution.Method != "" && resolution.Status == "unresolved" {
		resolution.Status = "resolved-name-unlinked"
	}
	userName := ""
	if turn.UserID != "" {
		userName = actorLabel("", "user", turn.TenantID, turn.UserID)
	}
	for _, card := range cards {
		if card.UserID == turn.UserID && card.Name != "" {
			userName = card.Name
		}
	}
	agentName := ""
	if turn.AgentID != "" {
		agentName = actorLabel("", "agent", turn.TenantID, turn.AgentID)
	}
	request := firstStatement(sourceText{name: "user", text: turn.UserText, source: model.SourceObserved})
	outcome := firstStatement(sourceText{name: "assistant", text: turn.AssistantText, source: model.SourceSynthetic})
	assistantClean := unquoted(turn.AssistantText)
	unresolved := UnresolvedReferences{Status: "no-unresolved-detected", Scope: "extracted-field-inputs", Entries: []UnresolvedReference{}}
	resolve := func(text, speaker, addressee, focus, field, role, evidence string) string {
		raw, masked := turn.UserText, clean
		if role == "assistant" {
			raw, masked = turn.AssistantText, assistantClean
		}
		mask := referenceInputMask(text, evidence, role, raw, masked)
		return resolveReferences(text, mask, speaker, addressee, focus, func(mention, reason string) {
			unresolved.record(field, role, mention, reason)
		})
	}
	if resolution.Mention != "" && resolution.Method == "" {
		_ = resolve(resolution.Mention, userName, agentName, "", "who", "user", event.Who.Evidence)
	}
	event.What.Value = bounded("request: "+resolve(request.value, userName, agentName, focus, "what.request", "user", request.evidence)+"; outcome: "+resolve(outcome.value, agentName, userName, "", "what.outcome", "assistant", outcome.evidence), maxFieldBytes)
	for i, f := range []*model.Field{&event.Where, &event.When, &event.Why, &event.How} {
		fieldName := [...]string{"where", "when", "why", "how"}[i]
		input, prefix := f.Value, ""
		// A fallback wrapper is not source text. Mask quotes before prepending it,
		// so a leading Markdown blockquote remains excluded from resolution.
		if f.Source == model.SourceInferred {
			if i == 2 && strings.HasPrefix(input, "to address the user request: ") {
				prefix = "to address the user request: "
			} else if i == 3 && strings.HasPrefix(input, "through the agent response: ") {
				prefix = "through the agent response: "
			}
			input = strings.TrimPrefix(input, prefix)
		}
		if strings.HasPrefix(f.Evidence, "user[bytes:") {
			fieldFocus := focus
			var start, end int
			if _, err := fmt.Sscanf(f.Evidence, "user[bytes:%d:%d]", &start, &end); err == nil && start >= 0 && start <= len(clean) {
				preceding := explicitSubjects(clean[:start])
				if len(preceding) == 1 {
					fieldFocus = preceding[0]
				} else if len(preceding) > 1 {
					fieldFocus = ""
				}
			}
			value := prefix + resolve(input, userName, agentName, fieldFocus, fieldName, "user", f.Evidence)
			if value != f.Value {
				f.Value, f.Source, f.Confidence = bounded(value, maxFieldBytes), model.SourceInferred, min(f.Confidence, .75)
			}
		} else if strings.HasPrefix(f.Evidence, "assistant[bytes:") {
			f.Value = bounded(prefix+resolve(input, agentName, userName, "", fieldName, "assistant", f.Evidence), maxFieldBytes)
		}
	}
	payload, _ := json.Marshal(resolution)
	event.Attributes["who_resolution"] = string(payload)
	payload, _ = json.Marshal(unresolved)
	event.Attributes["unresolved_references"] = string(payload)
	return event
}

func actorLabel(name, kind, tenantID, principalID string) string {
	key := kind + ":" + principalID
	if len(key) > 160 {
		// Preserve uniqueness in the bounded point field; the full namespace key
		// remains in resolution metadata rather than being truncated away.
		digest := sha256.Sum256([]byte(kind + "\x00" + tenantID + "\x00" + principalID))
		key = kind + ":sha256:" + hex.EncodeToString(digest[:])
	}
	if name == "" {
		return key
	}
	return name + " [" + key + "]"
}

func matchingCards(name string, cards []IdentityCard) []IdentityCard {
	var matches []IdentityCard
	for _, card := range cards {
		for _, alias := range append([]string{card.UserID}, card.Aliases...) {
			if aliasEqual(name, alias) {
				matches = append(matches, card)
				break
			}
		}
	}
	return matches
}

func aliasEqual(a, b string) bool {
	if strings.ContainsAny(b, ":@|/") {
		return a == b
	}
	return strings.EqualFold(a, b)
}

func explicitSubjects(text string) []string {
	var names []string
	seen := make(map[string]bool)
	for _, m := range subjectName.FindAllStringSubmatch(text, -1) {
		name := m[1]
		if nonName(name) {
			continue
		}
		key := strings.ToLower(name)
		if !seen[key] {
			names = append(names, name)
			seen[key] = true
		}
	}
	if len(names) == 0 {
		return nil
	}
	// An object or coordinated name can also be a pronoun antecedent. Refusing
	// additional capitalized mentions is conservative (a city/product can also
	// cause abstention), but avoids silently choosing Alex over Blair.
	for _, name := range namedMention.FindAllString(text, -1) {
		if nonName(name) {
			continue
		}
		key := strings.ToLower(name)
		if !seen[key] {
			names = append(names, name)
			seen[key] = true
		}
	}
	return names
}

func nonName(name string) bool {
	switch strings.ToLower(name) {
	case "i", "you", "he", "she", "they", "we", "it", "his", "her", "my", "our", "their", "your", "please", "hello", "hi", "then", "however", "also", "because", "tomorrow", "yesterday", "today":
		return true
	}
	return false
}

// Rebuild the quote mask using original evidence, before whitespace collapse.
// Normalizing the mask independently would lose alignment with retained text.
func referenceInputMask(text, evidence, role, raw, masked string) string {
	var start, end int
	if _, err := fmt.Sscanf(evidence, role+"[bytes:%d:%d]", &start, &end); err != nil || start < 0 || end < start || end > len(raw) {
		return unquoted(text)
	}
	var value, mask strings.Builder
	for offset := start; offset < end; {
		r, size := utf8.DecodeRuneInString(raw[offset:end])
		if unicode.IsSpace(r) {
			offset += size
			continue
		}
		wordStart := offset
		for offset < end {
			r, size = utf8.DecodeRuneInString(raw[offset:end])
			if unicode.IsSpace(r) {
				break
			}
			offset += size
			if value.Len()+offset-wordStart > maxFieldBytes {
				break
			}
		}
		if value.Len() != 0 {
			value.WriteByte(' ')
			mask.WriteByte(' ')
		}
		value.WriteString(raw[wordStart:offset])
		mask.WriteString(masked[wordStart:offset])
		if value.Len() > maxFieldBytes {
			break
		}
	}
	plain, clean := value.String(), mask.String()
	if text == plain {
		return clean
	}
	if strings.HasSuffix(text, "...") {
		prefix := strings.TrimSuffix(text, "...")
		if strings.HasPrefix(plain, prefix) {
			cut := len(prefix)
			result := clean[:cut]
			if cut < len(plain) {
				next, _ := utf8.DecodeRuneInString(plain[cut:])
				if referenceWordRune(next) {
					// Truncation must not turn part of an identifier into a pronoun.
					wordStart := strings.LastIndexByte(prefix, ' ') + 1
					result = result[:wordStart] + strings.Repeat(" ", cut-wordStart)
				}
			}
			return result + "..."
		}
	}
	return unquoted(text)
}

func referenceWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsMark(r) || r == '_'
}

func referenceBoundary(text string, start, end int) bool {
	connector := func(r rune) bool { return strings.ContainsRune("-/:@.", r) }
	if start > 0 {
		r, size := utf8.DecodeLastRuneInString(text[:start])
		if referenceWordRune(r) || r == '/' {
			return false
		}
		if connector(r) && start > size {
			before, _ := utf8.DecodeLastRuneInString(text[:start-size])
			if referenceWordRune(before) || before == '/' {
				return false
			}
		}
	}
	if end < len(text) {
		r, size := utf8.DecodeRuneInString(text[end:])
		if referenceWordRune(r) {
			return false
		}
		if connector(r) && end+size < len(text) {
			after, _ := utf8.DecodeRuneInString(text[end+size:])
			if referenceWordRune(after) || after == '/' {
				return false
			}
		}
	}
	return true
}

func resolveReferences(text, clean, speaker, addressee, focus string, unresolved func(string, string)) string {
	if urls := referenceURL.FindAllStringIndex(clean, -1); len(urls) != 0 {
		// URL queries/fragments can contain standalone-looking pronouns too.
		// Mask the whole literal without changing the retained value or offsets.
		masked := []byte(clean)
		for _, url := range urls {
			for i := url[0]; i < url[1]; i++ {
				masked[i] = ' '
			}
		}
		clean = string(masked)
	}
	indices := referenceWords.FindAllStringIndex(clean, -1)
	var out strings.Builder
	start := 0
	for _, index := range indices {
		if !referenceBoundary(text, index[0], index[1]) {
			continue
		}
		word := strings.ToLower(strings.ReplaceAll(text[index[0]:index[1]], "\u2019", "'"))
		replacement := ""
		reason := "missing-or-ambiguous-antecedent"
		switch word {
		case "i", "i am", "i'm", "i've", "i'll", "i'd", "me", "my", "mine":
			replacement = speaker
			reason = "missing-speaker-identity"
		case "you", "you are", "you're", "you've", "you'll", "you'd", "your", "yours":
			replacement = addressee
			reason = "missing-addressee-identity"
		case "we", "us", "our", "ours", "they", "them", "their", "theirs", "it", "its":
			reason = "unsupported-group-or-object-reference"
		case "myself", "yourself", "yourselves", "himself", "herself", "itself", "ourselves", "themselves":
			reason = "unsupported-reflexive-reference"
		case "her":
			reason = "ambiguous-object-or-possessive-grammar"
		case "he", "him", "his", "she", "hers":
			preceding := explicitSubjects(clean[:index[0]])
			if len(preceding) == 1 {
				replacement = preceding[0]
			} else if len(preceding) == 0 {
				replacement = focus
			}
		}
		if replacement == "" {
			unresolved(text[index[0]:index[1]], reason)
			continue
		}
		switch word {
		case "i am", "i'm", "you are", "you're":
			replacement += " is"
		case "i've", "you've":
			replacement += " has"
		case "i'll", "you'll":
			replacement += " will"
		case "i'd", "you'd":
			replacement += "'d" // Preserve the unresolved had/would distinction.
		}
		if word == "my" || word == "mine" || word == "your" || word == "yours" || word == "his" || word == "hers" {
			replacement += "'s"
		}
		out.WriteString(text[start:index[0]])
		out.WriteString(replacement)
		start = index[1]
	}
	out.WriteString(text[start:])
	return out.String()
}

// Mask quotations/code while retaining byte positions for original evidence.
func unquoted(text string) string {
	masked := []byte(text)
	mask := func(start, end int) {
		for i := start; i < end; i++ {
			if masked[i] != '\n' {
				masked[i] = ' '
			}
		}
	}
	offset := 0
	for _, line := range strings.SplitAfter(text, "\n") {
		if strings.HasPrefix(strings.TrimLeft(line, " \t"), ">") {
			mask(offset, offset+len(line))
		}
		offset += len(line)
	}
	clean := string(masked)
	quote := ""
	previous, escaped := rune(' '), false
	for i := 0; i < len(clean); {
		r, width := utf8.DecodeRuneInString(clean[i:])
		if quote != "" {
			if !escaped && strings.HasPrefix(clean[i:], quote) {
				mask(i, i+len(quote))
				i += len(quote)
				quote = ""
				previous = ' '
				continue
			}
			mask(i, i+width)
		} else if r == '"' || r == '`' || r == '\u201c' || r == '\u2018' || (r == '\'' && !unicode.IsLetter(previous) && !unicode.IsDigit(previous)) {
			quote = string(r)
			if r == '\u201c' {
				quote = "\u201d"
			}
			if r == '\u2018' {
				quote = "\u2019"
			}
			if strings.HasPrefix(clean[i:], "```") {
				quote, width = "```", 3
			}
			mask(i, i+width)
		}
		if r == '\\' {
			escaped = !escaped
		} else {
			escaped = false
		}
		previous = r
		i += width
	}
	return string(masked)
}
