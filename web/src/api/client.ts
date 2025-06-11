import type {
  ErrorResponse,
  ForceActiveRequest,
  Network,
  OverrideLeaderRequest,
  RemoveMemberRequest,
  Sequencer,
  TransferLeaderRequest,
  UpdateMembershipRequest,
} from "./types";

export class ApiError extends Error {
  constructor(
    public status: number,
    public data: ErrorResponse,
  ) {
    super(data.detail || data.title);
    this.name = "ApiError";
  }
}

export class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = "") {
    this.baseUrl = baseUrl;
  }

  private async request<T>(
    path: string,
    options: RequestInit = {},
  ): Promise<T> {
    const url = `${this.baseUrl}${path}`;

    try {
      const response = await fetch(url, {
        ...options,
        headers: {
          "Content-Type": "application/json",
          ...options.headers,
        },
      });

      if (!response.ok) {
        const errorData = await response.json() as ErrorResponse;
        throw new ApiError(response.status, errorData);
      }

      // Handle 204 No Content
      if (response.status === 204) {
        return {} as T;
      }

      return await response.json() as T;
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }
      throw new Error(`Network error: ${error}`);
    }
  }

  // Network endpoints

  async listNetworks(): Promise<Network[]> {
    return this.request<Network[]>("/api/v1/networks");
  }

  async getNetwork(networkId: string): Promise<Network> {
    return this.request<Network>(`/api/v1/networks/${networkId}`);
  }

  async getSequencers(networkId: string): Promise<Sequencer[]> {
    return this.request<Sequencer[]>(
      `/api/v1/networks/${networkId}/sequencers`,
    );
  }

  // Sequencer action endpoints

  async pauseSequencer(sequencerId: string): Promise<Sequencer> {
    return this.request<Sequencer>(
      `/api/v1/sequencers/${sequencerId}/pause`,
      { method: "POST" },
    );
  }

  async resumeSequencer(sequencerId: string): Promise<Sequencer> {
    return this.request<Sequencer>(
      `/api/v1/sequencers/${sequencerId}/resume`,
      { method: "POST" },
    );
  }

  async transferLeader(
    sequencerId: string,
    data: TransferLeaderRequest,
  ): Promise<{ message: string; target_id: string; target_addr: string }> {
    return this.request(
      `/api/v1/sequencers/${sequencerId}/transfer-leader`,
      {
        method: "POST",
        body: JSON.stringify(data),
      },
    );
  }

  async resignLeader(sequencerId: string): Promise<Sequencer> {
    return this.request<Sequencer>(
      `/api/v1/sequencers/${sequencerId}/resign-leader`,
      { method: "POST" },
    );
  }

  async overrideLeader(
    sequencerId: string,
    data: OverrideLeaderRequest,
  ): Promise<Sequencer> {
    return this.request<Sequencer>(
      `/api/v1/sequencers/${sequencerId}/override-leader`,
      {
        method: "POST",
        body: JSON.stringify(data),
      },
    );
  }

  async haltSequencer(sequencerId: string): Promise<Sequencer> {
    return this.request<Sequencer>(
      `/api/v1/sequencers/${sequencerId}/halt`,
      { method: "POST" },
    );
  }

  async forceActive(
    sequencerId: string,
    data: ForceActiveRequest,
  ): Promise<Sequencer> {
    return this.request<Sequencer>(
      `/api/v1/sequencers/${sequencerId}/force-active`,
      {
        method: "POST",
        body: JSON.stringify(data),
      },
    );
  }

  async removeMember(
    sequencerId: string,
    data: RemoveMemberRequest,
  ): Promise<void> {
    return this.request<void>(
      `/api/v1/sequencers/${sequencerId}/membership`,
      {
        method: "DELETE",
        body: JSON.stringify(data),
      },
    );
  }

  async updateMembership(
    sequencerId: string,
    data: UpdateMembershipRequest,
  ): Promise<Sequencer> {
    return this.request<Sequencer>(
      `/api/v1/sequencers/${sequencerId}/membership`,
      {
        method: "PUT",
        body: JSON.stringify(data),
      },
    );
  }

  // Health check
  async healthCheck(): Promise<{ status: string }> {
    return this.request<{ status: string }>("/health");
  }
}

// Create singleton instance
export const api = new ApiClient();

// Export default for convenience
export default api;
