import { create } from "zustand";
import { api, ApiError } from "@/api/client";
import type {
  ForceActiveRequest,
  OverrideLeaderRequest,
  RemoveMemberRequest,
  SequencerAction,
  TransferLeaderRequest,
  UpdateMembershipRequest,
} from "@/api/types";
import { useNetworkStore } from "./networkStore";
import { toast } from "@/hooks/use-toast";

interface ActionStore {
  // Modal state
  activeModal: SequencerAction | null;
  targetSequencerId: string | null;

  // Action state
  loading: boolean;
  error: string | null;

  // Modal actions
  openModal: (action: SequencerAction, sequencerId: string) => void;
  closeModal: () => void;

  // Complex actions
  transferLeader: (data: TransferLeaderRequest) => Promise<void>;
  resignLeader: () => Promise<void>;
  overrideLeader: (data: OverrideLeaderRequest) => Promise<void>;
  forceActive: (data: ForceActiveRequest) => Promise<void>;
  removeMember: (data: RemoveMemberRequest) => Promise<void>;
  updateMembership: (data: UpdateMembershipRequest) => Promise<void>;
}

export const useActionStore = create<ActionStore>((set, get) => ({
  // Initial state
  activeModal: null,
  targetSequencerId: null,
  loading: false,
  error: null,

  // Modal actions
  openModal: (action, sequencerId) => {
    set({
      activeModal: action,
      targetSequencerId: sequencerId,
      error: null,
    });
  },

  closeModal: () => {
    set({
      activeModal: null,
      targetSequencerId: null,
      error: null,
      loading: false,
    });
  },

  // Complex actions
  transferLeader: async (data) => {
    const { targetSequencerId } = get();
    if (!targetSequencerId) throw new Error("No sequencer selected");

    set({ loading: true, error: null });

    try {
      await api.transferLeader(targetSequencerId, data);
      toast({
        title: "Leadership Transferred",
        description:
          `Leadership transfer to ${data.target_id} initiated successfully`,
      });
      await useNetworkStore.getState().fetchNetworks();
      get().closeModal();
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.data.detail || error.message
        : "Failed to transfer leadership";
      set({ error: errorMessage, loading: false });
      toast({
        title: "Transfer Failed",
        description: errorMessage,
        variant: "destructive",
      });
      throw error;
    }
  },

  resignLeader: async () => {
    const { targetSequencerId } = get();
    if (!targetSequencerId) throw new Error("No sequencer selected");

    set({ loading: true, error: null });

    try {
      await api.resignLeader(targetSequencerId);
      await useNetworkStore.getState().fetchNetworks();
      get().closeModal();
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.data.detail || error.message
        : "Failed to resign leadership";
      set({ error: errorMessage, loading: false });
      throw error;
    }
  },

  overrideLeader: async (data) => {
    const { targetSequencerId } = get();
    if (!targetSequencerId) throw new Error("No sequencer selected");

    set({ loading: true, error: null });

    try {
      await api.overrideLeader(targetSequencerId, data);
      await useNetworkStore.getState().fetchNetworks();
      get().closeModal();
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.data.detail || error.message
        : "Failed to override leader";
      set({ error: errorMessage, loading: false });
      throw error;
    }
  },

  forceActive: async (data) => {
    const { targetSequencerId } = get();
    if (!targetSequencerId) throw new Error("No sequencer selected");

    set({ loading: true, error: null });

    try {
      await api.forceActive(targetSequencerId, data);
      await useNetworkStore.getState().fetchNetworks();
      get().closeModal();
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.data.detail || error.message
        : "Failed to force sequencer active";
      set({ error: errorMessage, loading: false });
      throw error;
    }
  },

  removeMember: async (data) => {
    const { targetSequencerId } = get();
    if (!targetSequencerId) throw new Error("No sequencer selected");

    set({ loading: true, error: null });

    try {
      await api.removeMember(targetSequencerId, data);
      await useNetworkStore.getState().fetchNetworks();
      get().closeModal();
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.data.detail || error.message
        : "Failed to remove member";
      set({ error: errorMessage, loading: false });
      throw error;
    }
  },

  updateMembership: async (data) => {
    const { targetSequencerId } = get();
    if (!targetSequencerId) throw new Error("No sequencer selected");

    set({ loading: true, error: null });

    try {
      await api.updateMembership(targetSequencerId, data);
      await useNetworkStore.getState().fetchNetworks();
      get().closeModal();
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.data.detail || error.message
        : "Failed to update membership";
      set({ error: errorMessage, loading: false });
      throw error;
    }
  },
}));
