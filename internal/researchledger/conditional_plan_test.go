package researchledger

import (
	"context"
	"encoding/json"
	"testing"
)

// Diagnostic VM plan, not an assertion that opcode counts predict latency.
func TestConditionalInsertPlan(t *testing.T) {
	l, err := Open(t.TempDir() + "/plan.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if err = l.EnableServiceIdentity(context.Background()); err != nil {
		t.Fatal(err)
	}
	r := serviceEntry("1", sourceFixture())
	identity, _ := json.Marshal(r.Key)
	for _, conditional := range []bool{false, true} {
		query := "EXPLAIN INSERT INTO research_log(identity,kind,payload) VALUES(?,?,?)"
		args := []any{string(identity), r.Kind, []byte(r.Payload)}
		if conditional {
			query = "EXPLAIN INSERT INTO research_log(identity,kind,payload) SELECT ?,?,? WHERE NOT EXISTS (SELECT 1 FROM research_log WHERE identity=? AND kind=?)"
			args = append(args, string(identity), r.Kind)
		}
		rows, err := l.db.Query(query, args...)
		if err != nil {
			t.Fatal(err)
		}
		counts := map[string]int{}
		for rows.Next() {
			var addr, p1, p2, p3, p5 int
			var opcode string
			var p4, comment any
			if err = rows.Scan(&addr, &opcode, &p1, &p2, &p3, &p4, &p5, &comment); err != nil {
				t.Fatal(err)
			}
			counts[opcode]++
		}
		if err = rows.Err(); err != nil {
			t.Fatal(err)
		}
		rows.Close()
		t.Logf("conditional%t opcodes%v", conditional, counts)
	}
}
