package researchledger

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// Keep the rejected candidate as a negative control, not a supported schema.
func TestPackedIndexDuplicateWitness(t *testing.T) {
	for _, packed := range []bool{false, true} {
		l, err := Open(t.TempDir() + "/duplicate.sqlite")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { l.Close() })
		ctx := context.Background()
		if packed {
			_, err = l.db.Exec(`CREATE UNIQUE INDEX packed_source_trial ON research_log(json_extract(identity,'$.Tenant','$.Journal','$.Contract'),json_extract(CAST(payload AS TEXT),'$.Binding.JournalID','$.Binding.EventID')) WHERE ` + servicePredicate)
		} else {
			err = l.EnableServiceIdentity(ctx)
		}
		if err != nil {
			t.Fatal(err)
		}
		key := sourceFixture()
		key.Journal = "alpha"
		a, b := serviceEntry("1", key), serviceEntry("2", key)
		b.Payload = []byte(strings.Replace(string(b.Payload), `"alpha"`, `"\u0061lpha"`, 1))
		_, err = l.AppendBatchPrepared(ctx, []AppendRequest{a, b})
		if (err == nil) != packed {
			t.Fatal("unexpected duplicate acceptance", packed, err)
		}
		rows, err := l.ReadAfter(ctx, 0, 8)
		want := 0
		if packed {
			want = 2
		}
		if err != nil || len(rows) != want {
			t.Fatal("unexpected committed rows", err, len(rows))
		}
	}
}

// Audit before benchmarking: text-valued JSON arrays may retain lexical escapes
// that scalar extraction decodes. Collection success is not candidate success.
func TestPackedIndexSemanticsExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_PACKED_SEMANTICS")
	if path == "" {
		t.Skip("opt-in candidate semantic falsifier")
	}
	l, err := Open(t.TempDir() + "/semantics.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	type result struct {
		Name, A, B               string
		ScalarEqual, PackedEqual bool
	}
	rows := []result{
		{Name: "ordinary", A: `{"a":"alpha","b":"beta"}`, B: `{"a":"alpha","b":"beta"}`},
		{Name: "unicode escape", A: `{"a":"alpha","b":"beta"}`, B: `{"a":"\u0061lpha","b":"beta"}`},
		{Name: "slash escape", A: `{"a":"a/b","b":"beta"}`, B: `{"a":"a\/b","b":"beta"}`},
		{Name: "whitespace", A: `{"a":"alpha","b":"beta"}`, B: `{ "b" : "beta", "a" : "alpha" }`},
		{Name: "delimiter boundary", A: `{"a":"a,b","b":"c"}`, B: `{"a":"a","b":"b,c"}`},
		{Name: "control escape", A: `{"a":"a\n","b":"beta"}`, B: `{"a":"a\u000a","b":"beta"}`},
	}
	for i := range rows {
		r := &rows[i]
		var a, b map[string]string
		if err := json.Unmarshal([]byte(r.A), &a); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(r.B), &b); err != nil {
			t.Fatal(err)
		}
		if err := l.db.QueryRow(`SELECT json_extract(?,'$.a')=json_extract(?,'$.a') AND json_extract(?,'$.b')=json_extract(?,'$.b'), json_extract(?,'$.a','$.b')=json_extract(?,'$.a','$.b')`, r.A, r.B, r.A, r.B, r.A, r.B).Scan(&r.ScalarEqual, &r.PackedEqual); err != nil {
			t.Fatal(err)
		}
		if r.ScalarEqual != (a["a"] == b["a"] && a["b"] == b["b"]) {
			t.Fatal("scalar reference disagrees with decoded strings")
		}
		t.Logf("%s scalar=%v packed=%v", r.Name, r.ScalarEqual, r.PackedEqual)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(rows); err != nil {
		t.Fatal(err)
	}
}
