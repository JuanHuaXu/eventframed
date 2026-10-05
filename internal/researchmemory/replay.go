package researchmemory

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

type RecordedFeedback struct {
	ID        uint64
	Useful    bool
	Available time.Time
}

// Discard uses the ledger's same terminal slot (historically called feedback),
// but a distinct canonical payload. It cannot be interpreted as Useful=false.
type RecordedDiscard struct {
	ID        uint64
	Discard   bool
	Available time.Time
}

// ResumeBackground transfers a rebuilt model and unresolved original records
// to a new worker. New operations are NOT automatically persisted: the future
// durable consumer must log them before acknowledgment. Replay grants no service
// dependency authority and requires the same exclusively owned quiescent log.
func ResumeBackground(ctx context.Context, log *researchledger.Ledger, tenant, stream string, epoch uint64, seed int64, capacity int) (*Background, error) {
	if capacity < 1 || capacity > 256 {
		return nil, errors.New("invalid background capacity")
	}
	a, e := ReplayLedger(ctx, log, tenant, stream, epoch, seed)
	if e != nil {
		return nil, e
	}
	return backgroundFromOwnedAdapter(a, capacity)
}

func decodeCanonical(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if e := decoder.Decode(target); e != nil {
		return e
	}
	canonical, e := json.Marshal(target)
	if e != nil {
		return e
	}
	// Owned log records use this encoder. Requiring its exact representation also
	// rejects omitted fields, duplicate keys, trailing values and ambiguous JSON.
	if !bytes.Equal(data, canonical) {
		return errors.New("noncanonical research record")
	}
	return nil
}

// ReplayLedger rebuilds one exclusively owned, quiescent research stream. Event
// keys here are decimal learner prediction IDs, NOT service event identifiers.
// The caller must stop writers and verify provenance/dependency authority first.
// No partial model is returned on failure, and no serving authority is granted.
func ReplayLedger(ctx context.Context, log *researchledger.Ledger, tenant, stream string, epoch uint64, seed int64) (*Adapter, error) {
	return replayLedger(ctx, log, tenant, stream, epoch, seed, false)
}

func replayLedger(ctx context.Context, log *researchledger.Ledger, tenant, stream string, epoch uint64, seed int64, requireBindings bool) (*Adapter, error) {
	if log == nil || tenant == "" || stream == "" {
		return nil, errors.New("invalid replay stream")
	}
	seal, err := log.Bootstrap(ctx)
	if err != nil {
		return nil, err
	}
	if seal != nil {
		return nil, errors.New("bound epoch requires bound replay")
	}
	return replayLedgerOnAdapter(ctx, log, tenant, stream, New(epoch, seed), requireBindings, false, nil)
}

// replayLedgerOnAdapter applies only this epoch's original records on top of a
// previously source-validated component model. The caller verifies the durable
// bootstrap seal before entering this function.
func replayLedgerOnAdapter(ctx context.Context, log *researchledger.Ledger, tenant, stream string, a *Adapter, requireBindings, requireWitness bool, validate func(context.Context, RecordedPrediction) error) (*Adapter, error) {
	if log == nil || a == nil || tenant == "" || stream == "" || a.epoch == 0 {
		return nil, errors.New("invalid bound replay stream")
	}
	var after int64
	for {
		rows, e := log.ReadAfter(ctx, after, 128)
		if e != nil {
			return nil, e
		}
		if len(rows) == 0 {
			return a, nil
		}
		for _, entry := range rows {
			if e = ctx.Err(); e != nil {
				return nil, e
			}
			if entry.Sequence <= after || entry.Key.Tenant != tenant || entry.Key.Journal != stream || entry.Key.Contract != RecordContract {
				return nil, errors.New("replay identity mismatch")
			}
			switch entry.Kind {
			case "admit":
				var r RecordedPrediction
				if e = decodeCanonical(entry.Payload, &r); e == nil {
					if (requireBindings && r.Binding == nil) || (requireWitness && (r.Witness == nil || r.Witness.FeatureContract != FeatureContract)) || entry.Key.Event != strconv.FormatUint(r.Prediction.ID, 10) || (r.Binding != nil && r.Binding.Tenant != tenant) {
						e = errors.New("prediction binding mismatch")
					} else if validate != nil {
						e = validate(ctx, r)
					}
					if e == nil {
						e = a.RestorePrediction(r)
					}
				}
			case "feedback":
				var r RecordedFeedback
				if e = decodeCanonical(entry.Payload, &r); e == nil {
					if entry.Key.Event != strconv.FormatUint(r.ID, 10) {
						e = errors.New("feedback binding mismatch")
					} else {
						e = a.Feedback(r.ID, r.Useful, a.epoch, r.Available)
					}
				} else {
					var discard RecordedDiscard
					e = decodeCanonical(entry.Payload, &discard)
					if e == nil {
						p, readErr := a.Record(discard.ID)
						if readErr != nil || !discard.Discard || entry.Key.Event != strconv.FormatUint(discard.ID, 10) || discard.Available.IsZero() || discard.Available.Before(p.At) {
							e = errors.New("invalid discard record")
						} else {
							a.Discard(discard.ID)
						}
					}
				}
			default:
				e = errors.New("unknown replay operation")
			}
			if e != nil {
				return nil, fmt.Errorf("replay sequence %d: %w", entry.Sequence, e)
			}
			after = entry.Sequence
		}
	}
}
