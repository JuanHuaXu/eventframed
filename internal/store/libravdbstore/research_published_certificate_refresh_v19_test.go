package libravdbstore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

// refreshCertificatesV19 assumes a separate auditor supplied these bounds.
// It tests publication cost and ordering, not the validity of that auditor.
func (g *incrementalSortGate) refreshCertificatesV19(ctx context.Context,
	selection model.SelectionSupportCertificate, omitted model.OmittedInfluenceCertificate,
	stopAt string,
) (model.Snapshot, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	marker, err := g.readMarker(ctx)
	if err != nil {
		return model.Snapshot{}, err
	}
	if g.verified == "" || marker.phase != "READY" || marker.digest != g.verified ||
		marker.count != g.wantRows || marker.lsn != g.verifiedLSN {
		return model.Snapshot{}, errors.New("certificate refresh gate is not current")
	}
	beforeLSN, err := g.store.db.LatestCommitLSN(ctx)
	if err != nil || beforeLSN != marker.lsn {
		g.verified = ""
		return model.Snapshot{}, fmt.Errorf("certificate refresh LSN moved: latest=%d marker=%d: %w", beforeLSN, marker.lsn, err)
	}
	before := g.store.Snapshot(ctx)
	if selection.PolicyVersion != before.PolicyVersion || omitted.PolicyVersion != before.PolicyVersion ||
		selection.EvidenceEpoch != before.EvidenceEpoch || omitted.EvidenceEpoch != before.EvidenceEpoch {
		return model.Snapshot{}, errors.New("synthetic audit outputs do not bind the current epoch")
	}
	if _, err := g.store.PublishSelectionCertificate(ctx, selection); err != nil {
		g.verified = ""
		return model.Snapshot{}, err
	}
	if _, err := g.store.PublishOmittedInfluenceCertificate(ctx, omitted); err != nil {
		g.verified = ""
		return model.Snapshot{}, err
	}
	after := g.store.Snapshot(ctx)
	afterLSN, err := g.store.db.LatestCommitLSN(ctx)
	if err != nil || afterLSN <= beforeLSN || after.RuntimeVersion != before.RuntimeVersion+2 ||
		after.EvidenceEpoch != before.EvidenceEpoch || after.PolicyVersion != before.PolicyVersion {
		g.verified = ""
		return model.Snapshot{}, fmt.Errorf("certificate refresh state mismatch: LSN %d -> %d: %v", beforeLSN, afterLSN, err)
	}
	savedSelection, selectionErr := g.store.GetSelectionCertificate(ctx, selection.TenantID)
	savedOmitted, omittedErr := g.store.GetOmittedInfluenceCertificate(ctx, omitted.TenantID)
	if selectionErr != nil || omittedErr != nil || !reflect.DeepEqual(savedSelection, selection) || !reflect.DeepEqual(savedOmitted, omitted) {
		g.verified = ""
		return model.Snapshot{}, fmt.Errorf("certificate refresh readback: selection=%v omitted=%v", selectionErr, omittedErr)
	}
	if stopAt == "after_db" {
		g.verified = ""
		return model.Snapshot{}, errors.New("injected interruption after certificate DB commits")
	}
	tx, err := g.sidecar.BeginTx(ctx, nil)
	if err != nil {
		g.verified = ""
		return model.Snapshot{}, err
	}
	defer tx.Rollback()
	updated, err := tx.ExecContext(ctx,
		"UPDATE marker SET lsn=? WHERE id=1 AND phase='READY' AND lsn=? AND row_count=? AND digest=?",
		afterLSN, marker.lsn, marker.count, marker.digest)
	if err != nil {
		g.verified = ""
		return model.Snapshot{}, err
	}
	if n, err := updated.RowsAffected(); err != nil || n != 1 {
		g.verified = ""
		return model.Snapshot{}, fmt.Errorf("certificate marker update affected %d rows: %v", n, err)
	}
	if err := tx.Commit(); err != nil {
		g.verified = ""
		return model.Snapshot{}, err
	}
	if stopAt == "after_sqlite" {
		g.verified = ""
		return model.Snapshot{}, errors.New("injected interruption after certificate marker commit")
	}
	g.verifiedLSN = afterLSN
	return after, nil
}

func (s *publishedOutcomeStoreV17) publishSyntheticCertificatesV19(ctx context.Context,
	selection model.SelectionSupportCertificate, omitted model.OmittedInfluenceCertificate,
) (model.Snapshot, error) {
	s.owner.Lock()
	defer s.owner.Unlock()
	snapshot, err := s.gate.refreshCertificatesV19(ctx, selection, omitted, "")
	if err != nil {
		s.current.Store(nil)
		return model.Snapshot{}, err
	}
	if err := s.publishLocked(ctx); err != nil {
		return model.Snapshot{}, err
	}
	return snapshot, nil
}

func syntheticCertificatesV19(snapshot model.Snapshot, now time.Time) (model.SelectionSupportCertificate, model.OmittedInfluenceCertificate) {
	selection := model.SelectionSupportCertificate{
		ID: "selection-v19", TenantID: "tenant-a", PolicyVersion: snapshot.PolicyVersion,
		EvidenceEpoch: snapshot.EvidenceEpoch, MinSelectionProbability: .2,
		SimultaneousCoverage: .95, Procedure: "synthetic assumption, not empirical audit",
		Issuer: "research-fixture", ExternalAudit: true,
		ValidFrom: now.Add(-time.Minute), ValidUntil: now.Add(time.Hour),
	}
	omitted := model.OmittedInfluenceCertificate{
		ID: "omitted-v19", TenantID: "tenant-a", PolicyVersion: snapshot.PolicyVersion,
		EvidenceEpoch: snapshot.EvidenceEpoch, DivergenceUCB: .02, DivergenceLimit: .05,
		AuditProbability: 1, SimultaneousCoverage: .95,
		Procedure: "synthetic assumption, not empirical audit",
		Issuer:    "research-fixture", ExternalAudit: true, ValidUntil: now.Add(time.Hour),
	}
	return selection, omitted
}

func TestResearchCertificateRefreshGateV19(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_CERTIFICATE_REFRESH_V19") != "1" {
		t.Skip("opt-in synthetic certificate publication contract")
	}
	ctx := context.Background()
	for _, stopAt := range []string{"", "after_db", "after_sqlite"} {
		t.Run("interruption-"+stopAt, func(t *testing.T) {
			root := t.TempDir()
			g := createIncrementalSortGate(t, root)
			defer func() {
				if g != nil {
					g.close()
				}
			}()
			before, ready := g.capture(ctx)
			if !ready {
				t.Fatal("initial marker not READY")
			}
			selection, omitted := syntheticCertificatesV19(g.store.Snapshot(ctx), time.Now().UTC())
			_, err := g.refreshCertificatesV19(ctx, selection, omitted, stopAt)
			if stopAt == "" && err != nil || stopAt != "" && err == nil {
				t.Fatal("certificate refresh/interruption", stopAt, err)
			}
			lsn, ready := g.capture(ctx)
			if stopAt == "" && (!ready || lsn <= before) || stopAt != "" && ready {
				t.Fatal("wrong in-memory READY state", stopAt, before, lsn, ready)
			}
			if err := g.close(); err != nil {
				t.Fatal(err)
			}
			g = nil
			g, err = openIncrementalSortGate(root)
			if err != nil {
				t.Fatal(err)
			}
			_, ready = g.capture(ctx)
			if ready != (stopAt != "after_db") {
				t.Fatal("wrong reopened READY state", stopAt, ready)
			}
		})
	}
	t.Run("direct_bypass_and_stale_epoch", func(t *testing.T) {
		g := createIncrementalSortGate(t, t.TempDir())
		defer g.close()
		old := g.store.Snapshot(ctx)
		selection, omitted := syntheticCertificatesV19(old, time.Now().UTC())
		if err := g.appendBatch(ctx, []ResearchEventWrite{pinnedWrite("new-epoch-v19", time.Now().UTC())}, ""); err != nil {
			t.Fatal(err)
		}
		if _, err := g.refreshCertificatesV19(ctx, selection, omitted, ""); err == nil {
			t.Fatal("stale synthetic audit output was accepted")
		}
		selection, _ = syntheticCertificatesV19(g.store.Snapshot(ctx), time.Now().UTC())
		if _, err := g.store.PublishSelectionCertificate(ctx, selection); err != nil {
			t.Fatal(err)
		}
		_, omitted = syntheticCertificatesV19(g.store.Snapshot(ctx), time.Now().UTC())
		if _, err := g.refreshCertificatesV19(ctx, selection, omitted, ""); err == nil {
			t.Fatal("direct metadata bypass was laundered")
		}
	})
}

func TestResearchCertificateRefreshLoadV19(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_CERTIFICATE_REFRESH_LOAD_V19") != "1" {
		t.Skip("opt-in synthetic certificate-supply load screen")
	}
	raceCorrectnessOnly := os.Getenv("EVENTFRAME_RACE_CORRECTNESS_ONLY") == "1"
	trials := 2
	if raceCorrectnessOnly {
		trials = 1
	}
	for trial := 1; trial <= trials; trial++ {
		for _, refresh := range []bool{false, true} {
			t.Run(fmt.Sprintf("trial-%d-refresh-%v", trial, refresh), func(t *testing.T) {
				runPublishedOutcomeLoadV18(t, true, raceCorrectnessOnly, refresh, refresh)
			})
		}
	}
}
