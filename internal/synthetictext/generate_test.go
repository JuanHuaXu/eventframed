package synthetictext

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/packing"
	"github.com/JuanHuaXu/eventframed/internal/ranking"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

func TestBuildProducesValidPublicFactCorpus(t *testing.T) {
	records, manifest := Build()
	if manifest.Sessions != 32 || manifest.Turns != 384 || manifest.Queries != 128 || manifest.Topics != 10 {
		t.Fatalf("manifest counts = %+v", manifest)
	}
	if manifest.SplitSessions["design"] != 24 || manifest.SplitSessions["confirmation"] != 8 {
		t.Fatalf("split counts = %+v", manifest.SplitSessions)
	}
	allowedHosts := map[string]bool{
		"science.nasa.gov": true, "www.ncei.noaa.gov": true, "www.usgs.gov": true,
		"www.nist.gov": true, "prod-01-asg-www-climate.woc.noaa.gov": true,
		"www.jpl.nasa.gov": true, "spaceplace.nasa.gov": true, "www.nasa.gov": true,
		"www.si.edu": true, "www.nssl.noaa.gov": true,
	}
	seen := make(map[string]int)
	for _, record := range records {
		if record.Capture.IdempotencyKey != record.Capture.Turn.ID {
			t.Fatalf("turn %s has non-local idempotency key %q", record.Capture.Turn.ID, record.Capture.IdempotencyKey)
		}
		if err := record.Capture.Turn.Validate(); err != nil {
			t.Fatalf("turn %s: %v", record.Capture.Turn.ID, err)
		}
		turn := record.Capture.Turn
		event := frame.FromTurn(turn)
		if err := event.Validate(0); err != nil {
			t.Fatalf("5W1H event %s: %v", turn.ID, err)
		}
		if event.Attributes["semantic_extractor"] != "fivew1h-deterministic-v1" {
			t.Fatalf("turn %s bypassed post-contract extraction", turn.ID)
		}
		if previous := seen[turn.SessionID]; int(turn.Sequence) != previous+1 {
			t.Fatalf("session %s sequence %d after %d", turn.SessionID, turn.Sequence, previous)
		}
		seen[turn.SessionID] = int(turn.Sequence)
		for _, ref := range record.Sources {
			parsed, err := url.Parse(ref.URL)
			if err != nil || parsed.Scheme != "https" || !allowedHosts[parsed.Host] {
				t.Fatalf("record %s has unapproved source %q", turn.ID, ref.URL)
			}
		}
		if record.Oracle != nil {
			if turn.Sequence == 10 && record.Oracle.QueryBeforeCapture == recordsForSession(records, turn.SessionID, 5).Capture.Turn.UserText {
				t.Fatalf("confirmation query %s repeats the design query exactly", turn.ID)
			}
			if turn.Sequence == 10 && !containsString(record.Oracle.RelevantPriorIDs, turn.SessionID+"-t05") {
				t.Fatalf("confirmation oracle %s omits the prior correct answer", turn.ID)
			}
			for _, id := range append(append([]string{}, record.Oracle.RelevantPriorIDs...), record.Oracle.ObsoletePriorIDs...) {
				if !strings.HasPrefix(id, turn.SessionID+"-t") || id >= turn.ID {
					t.Fatalf("oracle for %s references non-prior event %s", turn.ID, id)
				}
			}
		}
		encoded, err := json.Marshal(record)
		if err != nil {
			t.Fatal(err)
		}
		lower := strings.ToLower(string(encoded))
		for _, forbidden := range []string{"/users/", "@gmail", "@outlook", "chatgpt.com", "github.com/", "orcid.org/", "clawdius", "juanhua"} {
			if strings.Contains(lower, forbidden) {
				t.Fatalf("record %s contains private-source marker %q", turn.ID, forbidden)
			}
		}
	}
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func recordsForSession(records []Record, sessionID string, sequence uint64) Record {
	for _, record := range records {
		if record.Capture.Turn.SessionID == sessionID && record.Capture.Turn.Sequence == sequence {
			return record
		}
	}
	return Record{}
}

func TestBuildReplaysThroughCaptureTurnContract(t *testing.T) {
	records, _ := Build()
	embedder, err := embed.NewHashEmbedder(64)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := service.New(memorystore.New(), embedder, service.Config{
		DefaultRecallK: 16, DefaultPackK: 4, DefaultTokenBudget: 4096,
		OverfetchMultiplier: 1, RankingPolicy: ranking.DefaultPolicy(),
		PackingPolicy: packing.Policy{MaxPack: 4}, ResidualMode: service.ResidualModeDisabled,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range records {
		if _, err := runtime.CaptureTurn(context.Background(), record.Capture); err != nil {
			t.Fatalf("capture %s: %v", record.Capture.Turn.ID, err)
		}
	}
}

func TestBuildIsDeterministic(t *testing.T) {
	first, firstManifest := Build()
	second, secondManifest := Build()
	a, _ := json.Marshal(struct {
		R any
		M any
	}{first, firstManifest})
	b, _ := json.Marshal(struct {
		R any
		M any
	}{second, secondManifest})
	if string(a) != string(b) {
		t.Fatal("public-fact corpus generator is not deterministic")
	}
}
