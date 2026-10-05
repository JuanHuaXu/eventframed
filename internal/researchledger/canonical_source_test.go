package researchledger

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"unicode/utf8"
)

// Test-only preparation for a materialized index. Original bytes are never
// rewritten. Reject duplicate object keys rather than choosing first/last wins.
func sourceObject(raw []byte) (map[string]json.RawMessage, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return nil, errors.New("source object required")
	}
	out := map[string]json.RawMessage{}
	for d.More() {
		token, err = d.Token()
		if err != nil {
			return nil, err
		}
		name, ok := token.(string)
		if !ok {
			return nil, errors.New("invalid object key")
		}
		if _, exists := out[name]; exists {
			return nil, errors.New("duplicate source object key")
		}
		var value json.RawMessage
		if err := d.Decode(&value); err != nil {
			return nil, err
		}
		out[name] = value
	}
	if _, err := d.Token(); err != nil {
		return nil, err
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, errors.New("trailing source data")
	}
	return out, nil
}

func canonicalSourceKey(r AppendRequest) (string, error) {
	if r.Kind != "admit" || len(r.Payload) > 1<<20 || !utf8.Valid(r.Payload) {
		return "", errors.New("invalid source record")
	}
	fields := []string{r.Key.Tenant, r.Key.Journal, r.Key.Contract, r.Key.Event}
	outer, err := sourceObject(r.Payload)
	if err != nil {
		return "", err
	}
	for name := range outer {
		if name != "Binding" && strings.EqualFold(name, "Binding") {
			return "", errors.New("ambiguous binding alias")
		}
	}
	binding, err := sourceObject(outer["Binding"])
	if err != nil {
		return "", err
	}
	var tenant, journal, event string
	for name, target := range map[string]*string{"Tenant": &tenant, "JournalID": &journal, "EventID": &event} {
		for actual := range binding {
			if actual != name && strings.EqualFold(actual, name) {
				return "", errors.New("ambiguous source field alias")
			}
		}
		if err := json.Unmarshal(binding[name], target); err != nil {
			return "", err
		}
	}
	if tenant != r.Key.Tenant {
		return "", errors.New("source tenant mismatch")
	}
	fields = append(fields, tenant, journal, event)
	for _, s := range fields {
		// Conservatively exclude replacement characters and NUL in this prototype;
		// no claim of compatibility with every legacy accepted identity is made.
		if len(s) == 0 || len(s) > 4096 || !utf8.ValidString(s) || strings.ContainsRune(s, utf8.RuneError) || strings.ContainsRune(s, 0) {
			return "", errors.New("invalid source field")
		}
	}
	key, err := json.Marshal([]string{r.Key.Tenant, r.Key.Journal, r.Key.Contract, journal, event})
	return string(key), err
}

func TestCanonicalSourceKey(t *testing.T) {
	base := serviceEntry("1", ServiceIdentity{"tenant", "stream", "contract", "alpha", "a/b"})
	want, err := canonicalSourceKey(base)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range [][2]string{{`"alpha"`, `"\u0061lpha"`}, {`"a/b"`, `"a\/b"`}, {`"Binding"`, `"\u0042inding"`}} {
		r := base
		r.Payload = []byte(strings.Replace(string(base.Payload), change[0], change[1], 1))
		before := string(r.Payload)
		got, err := canonicalSourceKey(r)
		if err != nil || got != want || string(r.Payload) != before {
			t.Fatal("lexical parity", err, got)
		}
	}
	// Every source coordinate matters; the learner prediction ID does not.
	for i := 0; i < 5; i++ {
		k := ServiceIdentity{"tenant", "stream", "contract", "alpha", "a/b"}
		p := []*string{&k.Tenant, &k.Stream, &k.Contract, &k.Journal, &k.Event}
		*p[i] += "changed"
		got, err := canonicalSourceKey(serviceEntry("1", k))
		if err != nil || got == want {
			t.Fatal("coordinate collision", i, err)
		}
	}
	r := base
	r.Key.Event = "different learner ID"
	if got, err := canonicalSourceKey(r); err != nil || got != want {
		t.Fatal("learner ID changed source", err)
	}
	bad := []string{
		`{"Binding":{"Tenant":"tenant","JournalID":"alpha","EventID":"x","EventID":"y"}}`,
		`{"Binding":{"Tenant":"tenant","JournalID":"alpha","EventID":"x","\u0045ventID":"y"}}`,
		`{"Binding":{},"Binding":{}}`,
		`{"Binding":{"Tenant":"tenant","JournalID":"alpha","EventID":"x","eventid":"y"}}`,
		`{"Binding":{"Tenant":"tenant","JournalID":"alpha","EventID":"x"},"binding":{}}`,
		`{"binding":{"Tenant":"tenant","JournalID":"alpha","EventID":"x"}}`,
		`{"Binding":{"Tenant":"other","JournalID":"alpha","EventID":"x"}}`,
		`{"Binding":{"Tenant":"tenant","JournalID":"alpha","EventID":null}}`,
		`{"Binding":{"Tenant":"tenant","JournalID":"alpha","EventID":"\ud800"}}`,
		`{"Binding":{"Tenant":"tenant","JournalID":"alpha","EventID":"\u0000"}}`,
		string(base.Payload) + ` {}`,
	}
	for _, payload := range bad {
		r := base
		r.Payload = []byte(payload)
		if _, err := canonicalSourceKey(r); err == nil {
			t.Fatal("accepted ambiguity", payload)
		}
	}
}

func BenchmarkCanonicalSourceKey(b *testing.B) {
	r := phaseFixture(1, 1)[0]
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := canonicalSourceKey(r); err != nil {
			b.Fatal(err)
		}
	}
}
