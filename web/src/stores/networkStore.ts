import { create } from "zustand";
import { api, ApiError } from "@/api/client";
import type { NetworksState } from "@/api/types";
import { toast } from "@/hooks/use-toast";

interface NetworkStore extends NetworksState {
  // Actions
  fetchNetworks: () => Promise<void>;
  fetchNetwork: (networkId: string) => Promise<void>;
  setRefreshInterval: (interval: number) => void;

  // Sequencer actions
  pauseSequencer: (sequencerId: string) => Promise<void>;
  resumeSequencer: (sequencerId: string) => Promise<void>;
  haltSequencer: (sequencerId: string) => Promise<void>;

  // Auto-refresh
  startAutoRefresh: () => void;
  stopAutoRefresh: () => void;

  // Internal
  _refreshTimer?: NodeJS.Timeout;
}

export const useNetworkStore = create<NetworkStore>((set, get) => ({
  // Initial state
  networks: [],
  loading: false,
  error: null,
  refreshInterval: 5000, // 5 seconds default

  // Actions
  fetchNetworks: async () => {
    set({ loading: true, error: null });

    try {
      const networks = await api.listNetworks();

      // Sort networks by ID to ensure consistent ordering
      const sortedNetworks = [...networks].sort((a, b) =>
        a.id.localeCompare(b.id)
      );

      set({ networks: sortedNetworks, loading: false });
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.data.detail || error.message
        : "Failed to fetch networks";
      set({ error: errorMessage, loading: false });
    }
  },

  fetchNetwork: async (networkId) => {
    try {
      const network = await api.getNetwork(networkId);

      set((state) => {
        // Update the specific network in the list
        const networks = state.networks.map((n) =>
          n.id === networkId ? network : n
        );

        // If network doesn't exist yet, add it
        if (!state.networks.find((n) => n.id === networkId)) {
          networks.push(network);
          networks.sort((a, b) => a.id.localeCompare(b.id));
        }

        return { networks };
      });
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.data.detail || error.message
        : "Failed to fetch network";
      set({ error: errorMessage });
      throw error;
    }
  },

  setRefreshInterval: (interval) => {
    set({ refreshInterval: interval });
    // Restart auto-refresh with new interval
    const { stopAutoRefresh, startAutoRefresh } = get();
    stopAutoRefresh();
    if (interval > 0) {
      startAutoRefresh();
    }
  },

  // Sequencer actions
  pauseSequencer: async (sequencerId) => {
    try {
      const sequencerResponse = await api.pauseSequencer(sequencerId);
      toast({
        title: "Sequencer Paused",
        description: `Successfully paused sequencer ${sequencerId}`,
      });
      // Refresh just this network
      if (sequencerResponse.network_id) {
        await get().fetchNetwork(sequencerResponse.network_id);
      }
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.data.detail || error.message
        : "Failed to pause sequencer";
      set({ error: errorMessage });
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
      throw error;
    }
  },

  resumeSequencer: async (sequencerId) => {
    try {
      const sequencerResponse = await api.resumeSequencer(sequencerId);
      toast({
        title: "Sequencer Resumed",
        description: `Successfully resumed sequencer ${sequencerId}`,
      });
      // Refresh just this network
      if (sequencerResponse.network_id) {
        await get().fetchNetwork(sequencerResponse.network_id);
      }
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.data.detail || error.message
        : "Failed to resume sequencer";
      set({ error: errorMessage });
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
      throw error;
    }
  },

  haltSequencer: async (sequencerId) => {
    try {
      const sequencerResponse = await api.haltSequencer(sequencerId);
      toast({
        title: "Sequencer Halted",
        description: `Successfully halted sequencer ${sequencerId}`,
      });
      // Refresh just this network
      if (sequencerResponse.network_id) {
        await get().fetchNetwork(sequencerResponse.network_id);
      }
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.data.detail || error.message
        : "Failed to halt sequencer";
      set({ error: errorMessage });
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
      throw error;
    }
  },

  // Auto-refresh
  startAutoRefresh: () => {
    const { refreshInterval, fetchNetworks, _refreshTimer } = get();

    // Clear existing timer
    if (_refreshTimer) {
      clearInterval(_refreshTimer);
    }

    // Start new timer
    if (refreshInterval > 0) {
      const timer = setInterval(() => {
        fetchNetworks();
      }, refreshInterval);

      set({ _refreshTimer: timer });
    }
  },

  stopAutoRefresh: () => {
    const { _refreshTimer } = get();
    if (_refreshTimer) {
      clearInterval(_refreshTimer);
      set({ _refreshTimer: undefined });
    }
  },
}));

// Helper hooks
export const useSequencer = (sequencerId: string) => {
  const { networks } = useNetworkStore();

  for (const network of networks) {
    const sequencer = network.sequencers.find((s) => s.id === sequencerId);
    if (sequencer) {
      return { sequencer, network };
    }
  }

  return { sequencer: null, network: null };
};
