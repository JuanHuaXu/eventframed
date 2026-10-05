package researchledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"
)

// ServiceIdentity names a source observation, not a learner prediction ID.
// Stream is the ledger Key.Journal; Journal is the service journal identifier.
type ServiceIdentity struct{ Tenant, Stream, Contract, Journal, Event string }

const serviceIndexName = "research_service_identity_v1"
const servicePredicate = `kind='admit' AND json_type(CAST(payload AS TEXT),'$.Binding')='object'`
const serviceExpressions = `json_extract(identity,'$.Tenant'),json_extract(identity,'$.Journal'),json_extract(identity,'$.Contract'),json_extract(CAST(payload AS TEXT),'$.Binding.JournalID'),json_extract(CAST(payload AS TEXT),'$.Binding.EventID')`
const serviceIndexDDL = `CREATE UNIQUE INDEX research_service_identity_v1 ON research_log(` + serviceExpressions + `) WHERE ` + servicePredicate
const serviceLookupWhere = servicePredicate + ` AND json_extract(identity,'$.Tenant')=? AND json_extract(identity,'$.Journal')=? AND json_extract(identity,'$.Contract')=? AND json_extract(CAST(payload AS TEXT),'$.Binding.JournalID')=? AND json_extract(CAST(payload AS TEXT),'$.Binding.EventID')=?`
const serviceLookupSQL = `SELECT sequence,length(identity),CASE WHEN length(CAST(identity AS BLOB))<=131072 THEN identity ELSE NULL END,length(payload),typeof(payload),CASE WHEN typeof(payload)='blob' AND length(payload)<=1048576 THEN payload ELSE NULL END FROM research_log INDEXED BY research_service_identity_v1 WHERE ` + serviceLookupWhere + ` LIMIT 2`

// EnableServiceIdentity opts an owned log into persistent source uniqueness.
// Existing bound rows are structurally checked and duplicate history rejects
// atomically, without deleting records. Startup may scan/build an O(N) index;
// indexed lookup is separate. Call before exposing an owner, not on every read.
// The index enforces identity for bound records, not provenance or usefulness.
func (l *Ledger) EnableServiceIdentity(ctx context.Context) error {
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	invalid := []string{`json_type(CAST(payload AS TEXT),'$.Binding') IS NOT 'object'`}
	for _, field := range []struct{ source, path string }{{"identity", "$.Tenant"}, {"identity", "$.Journal"}, {"identity", "$.Event"}, {"identity", "$.Contract"}, {"CAST(payload AS TEXT)", "$.Binding.Tenant"}, {"CAST(payload AS TEXT)", "$.Binding.JournalID"}, {"CAST(payload AS TEXT)", "$.Binding.EventID"}} {
		expr := "json_extract(" + field.source + ",'" + field.path + "')"
		invalid = append(invalid, "json_type("+field.source+",'"+field.path+"') IS NOT 'text'", "length(CAST("+expr+" AS BLOB)) NOT BETWEEN 1 AND 4096")
	}
	invalid = append(invalid, `json_extract(CAST(payload AS TEXT),'$.Binding.Tenant') IS NOT json_extract(identity,'$.Tenant')`)
	var found int
	err = tx.QueryRowContext(ctx, `SELECT 1 FROM research_log WHERE kind='admit' AND json_type(CAST(payload AS TEXT),'$.Binding') IS NOT NULL AND (`+strings.Join(invalid, " OR ")+`) LIMIT 1`).Scan(&found)
	if err == nil {
		return errors.New("invalid existing service binding")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	var ddl string
	err = tx.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='index' AND name=?`, serviceIndexName).Scan(&ddl)
	if errors.Is(err, sql.ErrNoRows) {
		if _, err = tx.ExecContext(ctx, serviceIndexDDL); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if ddl != serviceIndexDDL {
		return errors.New("unexpected service identity index definition")
	}
	return tx.Commit()
}

// GetServiceAdmission resolves a source to its persisted original. Enable the
// index first and retain exclusive schema ownership. A missing index is an
// error, not a cache miss. Result materialization is bounded even for corruption;
// full original/forecast validation remains the consumer's responsibility.
func (l *Ledger) GetServiceAdmission(ctx context.Context, k ServiceIdentity) (Entry, error) {
	for _, v := range []string{k.Tenant, k.Stream, k.Contract, k.Journal, k.Event} {
		if len(v) == 0 || len(v) > 4096 || !utf8.ValidString(v) {
			return Entry{}, errors.New("invalid service lookup key")
		}
	}
	rows, err := l.db.QueryContext(ctx, serviceLookupSQL, k.Tenant, k.Stream, k.Contract, k.Journal, k.Event)
	if err != nil {
		return Entry{}, err
	}
	return readServiceAdmission(rows, k, 1<<20)
}

// The point and transactional batch paths share exactly the same envelope and
// source checks. limit also enforces the batch's remaining payload budget.
func readServiceAdmission(rows *sql.Rows, k ServiceIdentity, limit int) (Entry, error) {
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return Entry{}, err
		}
		return Entry{}, sql.ErrNoRows
	}
	var e Entry
	var rawIdentity sql.NullString
	var identityLength, payloadLength int64
	var storageType string
	err := rows.Scan(&e.Sequence, &identityLength, &rawIdentity, &payloadLength, &storageType, &e.Payload)
	if err != nil {
		return Entry{}, err
	}
	if !rawIdentity.Valid || identityLength <= 0 || len(rawIdentity.String) > 131072 || e.Sequence <= 0 || storageType != "blob" || payloadLength <= 0 || payloadLength > int64(limit) || int64(len(e.Payload)) != payloadLength {
		return Entry{}, errors.New("invalid service original envelope")
	}
	if err = json.Unmarshal([]byte(rawIdentity.String), &e.Key); err != nil {
		return Entry{}, err
	}
	if e.Key.Tenant != k.Tenant || e.Key.Journal != k.Stream || e.Key.Contract != k.Contract {
		return Entry{}, errors.New("service owner mismatch")
	}
	var source struct {
		Binding struct{ Tenant, JournalID, EventID string }
	}
	if err = json.Unmarshal(e.Payload, &source); err != nil {
		return Entry{}, err
	}
	if source.Binding.Tenant != k.Tenant || source.Binding.JournalID != k.Journal || source.Binding.EventID != k.Event {
		return Entry{}, errors.New("service source mismatch")
	}
	if rows.Next() {
		return Entry{}, errors.New("ambiguous service identity")
	}
	if err = rows.Err(); err != nil {
		return Entry{}, err
	}
	e.Kind = "admit"
	return e, nil
}
