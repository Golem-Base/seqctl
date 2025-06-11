// API Types matching Go structures

// Network represents a network in API responses
export interface Network {
  id: string;
  name: string;
  healthy: boolean;
  sequencers: Sequencer[];
  updated_at: string;
  _links: NetworkLinks;
}

// NetworkLinks represents HATEOAS links for a network
export interface NetworkLinks {
  self: Link;
  sequencers: Link;
}

// Sequencer represents a sequencer in API responses
export interface Sequencer {
  id: string;
  network_id: string;
  raft_addr: string;
  conductor_active: boolean;
  conductor_leader: boolean;
  conductor_paused: boolean;
  conductor_stopped: boolean;
  sequencer_healthy: boolean;
  sequencer_active: boolean;
  unsafe_l2: number;
  voting: boolean;
  updated_at: string;
  _links: SequencerLinks;
}

// SequencerLinks represents HATEOAS links for a sequencer
export interface SequencerLinks {
  self: Link;
  network: Link;
  pause?: Link;
  resume?: Link;
  transfer_leader?: Link;
  resign_leader?: Link;
  override_leader?: Link;
  halt?: Link;
  force_active?: Link;
  remove_member?: Link;
  update_member?: Link;
}

// Link represents a HATEOAS link
export interface Link {
  href: string;
  method?: string;
}

// ErrorResponse represents an error response following RFC 7807
export interface ErrorResponse {
  type: string;
  title: string;
  status: number;
  detail?: string;
  instance?: string;
  errors?: Record<string, any>;
}

// Request types

export interface TransferLeaderRequest {
  target_id: string;
  target_addr: string;
}

export interface OverrideLeaderRequest {
  override: boolean;
}

export interface ForceActiveRequest {
  block_hash?: string;
}

export interface RemoveMemberRequest {
  server_id: string;
}

export interface UpdateMembershipRequest {
  server_id: string;
  server_addr: string;
  voting: boolean;
}

// Helper types for state management

export type NetworksState = {
  networks: Network[];
  loading: boolean;
  error: string | null;
  refreshInterval: number;
};

export type SequencerAction =
  | "pause"
  | "resume"
  | "transfer-leader"
  | "resign-leader"
  | "override-leader"
  | "halt"
  | "force-active"
  | "remove-member"
  | "update-member";

export interface ActionState {
  sequencerId: string | null;
  action: SequencerAction | null;
  loading: boolean;
  error: string | null;
}
