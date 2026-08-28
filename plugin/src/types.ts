export const PROTOCOL_VERSION = "eventframe.v1alpha1" as const;

export type FieldSource = "observed" | "inferred" | "synthetic";

export type EventField = {
  value: string;
  source: FieldSource;
  confidence: number;
  evidence?: string;
};

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
  who: EventField;
  what: EventField;
  where: EventField;
  when: EventField;
  why: EventField;
  how: EventField;
  priority: number;
  tags?: string[];
  provenance: {
    producer: string;
    source_event_ids?: string[];
    retrieved_ids?: string[];
    tool_call_id?: string;
    run_id?: string;
  };
  attributes?: Record<string, string>;
  embedding_model?: string;
};

export type ContextCandidate = {
  event: EventFrame;
  similarity: number;
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
    nominated: number;
    activated: number;
    selection_support_certified: false;
    decisions: Array<{
      event_id: string;
      activation_score: number;
      activated: boolean;
      evidence_ready: boolean;
      audit_selected: boolean;
      audit_probability: number;
      posterior_key: string;
    }>;
  };
};

export type AdapterConfig = {
  socketPath: string;
  tenantId: string;
  recallK: number;
  packK: number;
  tokenBudget: number;
  capture: boolean;
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
