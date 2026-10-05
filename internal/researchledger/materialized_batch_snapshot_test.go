package researchledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/url"
	"testing"
)

func TestMaterializedBatchSnapshotAndAliases(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/snapshot.sqlite"
	l, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if err := materializedSourceSchema(l); err != nil {
		t.Fatal(err)
	}
	first, late := sourceFixture(), sourceFixture()
	late.Event = "late"
	a, b := serviceEntry("1", first), serviceEntry("2", late)
	if _, err := materializedSourceAppend(ctx, l, []AppendRequest{a}, nil); err != nil {
		t.Fatal(err)
	}
	u := url.URL{Scheme: "file", Path: path}
	writer, err := sql.Open("sqlite", u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	requests := []ServiceIdentity{first, late, first}
	got, err := materializedSourceBatch(ctx, l, requests, func(i int) {
		if i == 0 {
			id, _ := json.Marshal(b.Key)
			source, err := canonicalSourceKey(b)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := writer.ExecContext(ctx, `INSERT INTO materialized_source_trial(identity,source,payload) VALUES(?,?,?)`, string(id), source, []byte(b.Payload)); err != nil {
				t.Fatal(err)
			}
		}
	})
	if err != nil || len(got) != 3 || !got[0].Found || got[1].Found || !got[2].Found {
		t.Fatal("mixed snapshot", err)
	}
	for i, r := range got {
		if r.Source != requests[i] {
			t.Fatal("reordered")
		}
	}
	got[0].Entry.Payload[0] = 'x'
	if got[2].Entry.Payload[0] == 'x' {
		t.Fatal("result alias")
	}
	got, err = materializedSourceBatch(ctx, l, requests, nil)
	if err != nil || !got[1].Found || got[0].Entry.Payload[0] == 'x' {
		t.Fatal("next snapshot", err)
	}
}
