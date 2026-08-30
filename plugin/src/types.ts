export const PROTOCOL_VERSION = "eventframe.v1alpha1" as const;

export type EventFrame = {
  id: string;
  tenant_id: string;
  session_id: string;
  sequence: number;
  kind: string;
  content: string;
  occurred_at: string;
  observed_at: string;
  available_at: string;
  priority: number;
  tags?: string[];
  provenance: {
    producer: string;
    source_event_ids?: string[];
    retrieved_ids?: string[];
    tool_call_id?: string;
    run_id?: string;
  };
};

// CapturedTurn is the OpenClaw transport envelope. It intentionally contains no
// 5W1H fields; eventframed performs semantic enrichment after contract receipt.
export type CapturedTurn = {
  id: string;
  tenant_id: string;
  session_id: string;
  sequence: number;
  run_id?: string;
  agent_id?: string;
  user_text: string;
  assistant_text: string;
  retrieved_ids?: string[];
  occurred_at: string;
  observed_at: string;
  available_at: string;
};

export type ContextCandidate = {
  event: EventFrame;
  similarity: number;
  graph_compatibility: number;
  graph_applied: boolean;
  retrieval_score: number;
  rank_delta: number;
  rank_delta_scale?: number;
  rank_delta_answer_certainty?: number;
  rank_delta_correction_reliability?: number;
  rank_delta_confidence?: number;
  rank_delta_basis?: "rank-boundary+correction-reliability";
  retrieval_contract: string;
  score: number;
  estimated_tokens: number;
};

export type ContextPacket = {
  protocol_version: string;
  candidates: ContextCandidate[];
  recalled: number;
  eligible: number;
  packed: number;
  used_tokens: number;
  packet_confidence?: number;
  packet_answer_certainty?: number;
  nomination_contract: string;
  retrieval_contract: string;
  snapshot: {
    runtime_version: number;
    policy_version: number;
    contract_version: number;
    graph_version: number;
    posterior_version: number;
    residual_version: number;
    abstraction_version: number;
    agency_version: number;
    evidence_epoch: number;
  };
  bayesian_shadow?: {
    mode: "shadow";
    journal_id: string;
    journal_durable: boolean;
    nominated: number;
    activated: number;
    deep_reviewed: number;
    selection_support_certified: false;
    decisions: Array<{
      event_id: string;
      activation_score: number;
      activated: boolean;
      cheap_update: boolean;
      deep_review: boolean;
      evidence_ready: boolean;
      audit_selected: boolean;
      audit_probability: number;
      posterior_key: string;
    }>;
  };
};

export type OutcomeSignal = "useful" | "not_useful" | "cited" | "successful_downstream" | "correction" | "rejected";

export type OutcomeObservation = {
  tenant_id: string;
  journal_id: string;
  event_id: string;
  idempotency_key: string;
  signal: OutcomeSignal;
  observed_at?: string;
  available_at?: string;
};

export type AdapterConfig = {
  socketPath: string;
  tenantId: string;
  recallK: number;
  packK: number;
  tokenBudget: number;
  capture: boolean;
  tracePath?: string;
  agencyEnabled: boolean;
  agencyKillSwitch: boolean;
  agencyPublicKeyPath: string;
  agencyAuthorityTokenPath: string;
  agencyCapabilities: string[];
  agencyConsentActions: AgencyAction[];
  agencyConsumerId: string;
  agencyPollIntervalMs: number;
  agencyMaxClaims: number;
  agencyMaxChainDepth: number;
  agencyQuietHoursStartUtc?: number;
  agencyQuietHoursEndUtc?: number;
  agencyCriticalThreshold: number;
  agencyAllowedSessionPrefixes: string[];
};

export type AgencyAction = "wake" | "notify" | "schedule";

export type AgencyProposal = {
  id: string;
  tenant_id: string;
  session_id: string;
  action: AgencyAction;
  reason: string;
  evidence_ids: string[];
  expected_utility: number;
  priority: number;
  required_capability: string;
  not_before: string;
  scheduled_for?: string;
  expires_at: string;
  idempotency_key: string;
  causal_chain_id: string;
  parent_proposal_id?: string;
  causal_chain_depth: number;
  created_at: string;
  contract_version: number;
};

export type AgencyProposalRecord = {
  proposal: AgencyProposal;
  signed: { payload: string; signature: string; key_id: string };
  status: "pending" | "claimed" | "approved" | "rejected" | "expired";
  claimed_by?: string;
  lease_until?: string;
  resolution_reason?: string;
  execution_ref?: string;
  resolved_at?: string;
};

export type AgencyClaimResponse = {
  protocol_version: string;
  records: AgencyProposalRecord[];
  snapshot: ContextPacket["snapshot"];
};
