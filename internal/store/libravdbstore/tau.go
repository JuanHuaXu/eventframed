package libravdbstore

import (
	"context"
	"errors"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

func (s *Store) GetAgencyProposal(ctx context.Context, tenantID, proposalID string) (model.AgencyProposalRecord, error) {
	s.writeMu.RLock()
	defer s.writeMu.RUnlock()
	r, err := s.agency.Get(ctx, agencyProposalRecordID(tenantID, proposalID))
	if errors.Is(err, libra.ErrRecordNotFound) {
		return model.AgencyProposalRecord{}, store.ErrAgencyNotFound
	}
	if err != nil {
		return model.AgencyProposalRecord{}, err
	}
	return decodeAgencyRecord(r.Metadata)
}
