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
    evidence_epoch: number;
  };
};

export type AdapterConfig = {
  socketPath: string;
  tenantId: string;
  recallK: number;
  packK: number;
  tokenBudget: number;
  capture: boolean;
};
