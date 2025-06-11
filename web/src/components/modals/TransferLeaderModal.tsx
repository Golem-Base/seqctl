import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useActionStore } from "@/stores/actionStore";
import { useNetworkStore, useSequencer } from "@/stores/networkStore";
import { AlertCircle, ArrowRightLeft } from "lucide-react";

const transferLeaderSchema = z.object({
  target_id: z.string().min(1, "Target ID is required"),
  target_addr: z.string().min(1, "Target address is required"),
});

type TransferLeaderForm = z.infer<typeof transferLeaderSchema>;

export function TransferLeaderModal() {
  const {
    activeModal,
    targetSequencerId,
    closeModal,
    transferLeader,
    loading,
    error,
  } = useActionStore();
  const { networks } = useNetworkStore();
  const { sequencer, network } = useSequencer(targetSequencerId || "");

  const isOpen = activeModal === "transfer-leader";

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    reset,
    formState: { errors },
  } = useForm<TransferLeaderForm>({
    resolver: zodResolver(transferLeaderSchema),
  });

  const selectedTargetId = watch("target_id");

  // Get available targets (non-leader sequencers in the same network)
  const availableTargets = network?.sequencers.filter(
    (s) => s.id !== sequencer?.id && !s.conductor_leader,
  ) || [];

  const onSubmit = async (data: TransferLeaderForm) => {
    try {
      await transferLeader(data);
      reset();
    } catch (error) {
      // Error is handled in the store
    }
  };

  const handleClose = () => {
    closeModal();
    reset();
  };

  const handleTargetSelect = (targetId: string) => {
    const target = availableTargets.find((s) => s.id === targetId);
    if (target) {
      setValue("target_id", target.id);
      setValue("target_addr", target.raft_addr);
    }
  };

  if (!sequencer || !network) return null;

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <ArrowRightLeft className="h-5 w-5 text-yellow-500" />
            Transfer Leadership
          </DialogTitle>
          <DialogDescription>
            Transfer Raft leadership from {sequencer.id} to another sequencer
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="rounded-md bg-yellow-50 dark:bg-yellow-900/20 p-4">
            <div className="flex">
              <AlertCircle className="h-5 w-5 text-yellow-600 dark:text-yellow-400 mr-2 flex-shrink-0 mt-0.5" />
              <div className="text-sm text-yellow-800 dark:text-yellow-200">
                <p className="font-medium mb-1">Important:</p>
                <p>
                  Leadership transfer will temporarily disrupt consensus
                  operations. Ensure the target sequencer is healthy and
                  properly connected.
                </p>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <div>
              <Label htmlFor="target">Target Sequencer</Label>
              <Select
                value={selectedTargetId}
                onValueChange={handleTargetSelect}
              >
                <SelectTrigger id="target">
                  <SelectValue placeholder="Select a target sequencer" />
                </SelectTrigger>
                <SelectContent>
                  {availableTargets.map((seq) => (
                    <SelectItem key={seq.id} value={seq.id}>
                      <div className="flex items-center justify-between w-full">
                        <span>{seq.id}</span>
                        {seq.voting && (
                          <span className="text-xs text-muted-foreground ml-2">
                            Voting
                          </span>
                        )}
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {errors.target_id && (
                <p className="text-sm text-destructive mt-1">
                  {errors.target_id.message}
                </p>
              )}
            </div>

            <div>
              <Label htmlFor="target_addr">Target Address</Label>
              <Input
                id="target_addr"
                {...register("target_addr")}
                placeholder="e.g., sequencer-1:2222"
                className="font-mono"
                readOnly={!!selectedTargetId}
              />
              {errors.target_addr && (
                <p className="text-sm text-destructive mt-1">
                  {errors.target_addr.message}
                </p>
              )}
            </div>
          </div>

          {error && (
            <div className="rounded-md bg-destructive/15 p-3">
              <p className="text-sm text-destructive">{error}</p>
            </div>
          )}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? "Transferring..." : "Transfer Leadership"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
