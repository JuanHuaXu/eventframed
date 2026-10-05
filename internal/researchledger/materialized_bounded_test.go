package researchledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

const materializedBoundedSQL = `SELECT sequence,length(identity),CASE WHEN length(CAST(identity AS BLOB))<=131072 THEN identity ELSE NULL END,length(payload),typeof(payload),CASE WHEN typeof(payload)='blob' AND length(payload)<=1048576 THEN payload ELSE NULL END FROM materialized_source_trial WHERE source=? LIMIT 2`

// Same projection limits and row validator as the original indexed reader.
func materializedSourceReadBounded(ctx context.Context, l *Ledger, k ServiceIdentity) (Entry, error) {
	for _, v := range []string{k.Tenant, k.Stream, k.Contract, k.Journal, k.Event} {
		if len(v) == 0 || len(v) > 4096 || !utf8.ValidString(v) {
			return Entry{}, errors.New("invalid service lookup key")
		}
	}
	key, err := json.Marshal([]string{k.Tenant, k.Stream, k.Contract, k.Journal, k.Event})
	if err != nil {
		return Entry{}, err
	}
	rows, err := l.db.QueryContext(ctx, materializedBoundedSQL, string(key))
	if err != nil {
		return Entry{}, err
	}
	return readServiceAdmission(rows, k, 1<<20)
}

func TestMaterializedBoundedRead(t *testing.T) {
	ctx := context.Background()
	for _, damage := range []string{"none", "oversized payload", "oversized identity", "text payload", "source mismatch", "owner mismatch"} {
		t.Run(damage, func(t *testing.T) {
			l, err := Open(t.TempDir() + "/bounded.sqlite")
			if err != nil {
				t.Fatal(err)
			}
			defer l.Close()
			if err := materializedSourceSchema(l); err != nil {
				t.Fatal(err)
			}
			k := sourceFixture()
			r := serviceEntry("1", k)
			if _, err := materializedSourceAppend(ctx, l, []AppendRequest{r}, nil); err != nil {
				t.Fatal(err)
			}
			statements := map[string]string{
				"oversized payload":  `UPDATE materialized_source_trial SET payload=zeroblob(1048577)`,
				"oversized identity": `UPDATE materialized_source_trial SET identity=CAST(zeroblob(131073) AS TEXT)`,
				"text payload":       `UPDATE materialized_source_trial SET payload=CAST(payload AS TEXT)`,
				"source mismatch":    `UPDATE materialized_source_trial SET payload=CAST('{"Binding":{"Tenant":"tenant","JournalID":"journal","EventID":"different"}}' AS BLOB)`,
				"owner mismatch":     `UPDATE materialized_source_trial SET identity='{"Tenant":"other"}'`,
			}
			if q := statements[damage]; q != "" {
				if _, err := l.db.Exec(q); err != nil {
					t.Fatal(err)
				}
			}
			e, err := materializedSourceReadBounded(ctx, l, k)
			if (err == nil) != (damage == "none") {
				t.Fatal("damage handling", err)
			}
			if damage == "none" {
				if e.Key != r.Key || string(e.Payload) != string(r.Payload) {
					t.Fatal("original mismatch")
				}
				missing := k
				missing.Event = "missing"
				if _, err := materializedSourceReadBounded(ctx, l, missing); !errors.Is(err, sql.ErrNoRows) {
					t.Fatal("miss", err)
				}
				for _, bad := range []string{"", strings.Repeat("x", 4097), string([]byte{255})} {
					invalid := k
					invalid.Event = bad
					if _, err := materializedSourceReadBounded(ctx, l, invalid); err == nil {
						t.Fatal("bad key")
					}
				}
				c, cancel := context.WithCancel(ctx)
				cancel()
				if _, err := materializedSourceReadBounded(c, l, k); err == nil {
					t.Fatal("canceled read")
				}
			}
			if strings.HasPrefix(damage, "oversized") {
				key, _ := json.Marshal([]string{k.Tenant, k.Stream, k.Contract, k.Journal, k.Event})
				var seq, n, p int64
				var identity sql.NullString
				var kind string
				var payload []byte
				if err := l.db.QueryRow(materializedBoundedSQL, string(key)).Scan(&seq, &n, &identity, &p, &kind, &payload); err != nil {
					t.Fatal(err)
				}
				if damage == "oversized payload" && payload != nil || damage == "oversized identity" && identity.Valid {
					t.Fatal("oversized column materialized")
				}
			}
		})
	}
}
