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
import { useActionStore } from "@/stores/actionStore";
import { useSequencer } from "@/stores/networkStore";
import { Info, Users } from "lucide-react";

const updateMembershipSchema = z.object({
  server_id: z.string().min(1, "Server ID is required"),
  server_addr: z.string().min(1, "Server address is required"),
  voting: z.boolean(),
});

type UpdateMembershipForm = z.infer<typeof updateMembershipSchema>;

export function UpdateMembershipModal() {
  const {
    activeModal,
    targetSequencerId,
    closeModal,
    updateMembership,
    loading,
    error,
  } = useActionStore();
  const { sequencer, network } = useSequencer(targetSequencerId || "");

  const isOpen = activeModal === "update-member";

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm<UpdateMembershipForm>({
    resolver: zodResolver(updateMembershipSchema),
    defaultValues: {
      voting: true,
    },
  });

  const voting = watch("voting");

  const onSubmit = async (data: UpdateMembershipForm) => {
    try {
      await updateMembership(data);
      reset();
    } catch (error) {
      // Error is handled in the store
    }
  };

  const handleClose = () => {
    closeModal();
    reset();
  };

  if (!sequencer || !network) return null;

  // Get existing members for reference
  const existingMembers = network.sequencers.map((s) => ({
    id: s.id,
    addr: s.raft_addr,
    voting: s.voting,
  }));

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[550px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Users className="h-5 w-5 text-blue-500" />
            Update Cluster Membership
          </DialogTitle>
          <DialogDescription>
            Add a new server to the Raft cluster
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="rounded-md bg-blue-50 dark:bg-blue-900/20 p-4">
            <div className="flex">
              <Info className="h-5 w-5 text-blue-600 dark:text-blue-400 mr-2 flex-shrink-0 mt-0.5" />
              <div className="text-sm text-blue-800 dark:text-blue-200">
                <p className="font-medium mb-1">Membership Update</p>
                <p>
                  Adding a new server to the cluster. Voting members participate
                  in leader election and consensus. Non-voting members receive
                  replicated data but don't vote.
                </p>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <div>
              <Label>Current Members ({existingMembers.length})</Label>
              <div className="mt-2 space-y-1 max-h-32 overflow-y-auto">
                {existingMembers.map((member) => (
                  <div
                    key={member.id}
                    className="text-sm flex items-center justify-between p-2 bg-muted rounded"
                  >
                    <span className="font-mono">{member.id}</span>
                    <span className="text-xs text-muted-foreground">
                      {member.voting ? "Voting" : "Non-voting"}
                    </span>
                  </div>
                ))}
              </div>
            </div>

            <div>
              <Label htmlFor="server_id">Server ID</Label>
              <Input
                id="server_id"
                {...register("server_id")}
                placeholder="e.g., sequencer-3"
              />
              {errors.server_id && (
                <p className="text-sm text-destructive mt-1">
                  {errors.server_id.message}
                </p>
              )}
            </div>

            <div>
              <Label htmlFor="server_addr">Server Address</Label>
              <Input
                id="server_addr"
                {...register("server_addr")}
                placeholder="e.g., sequencer-3:2222"
                className="font-mono"
              />
              {errors.server_addr && (
                <p className="text-sm text-destructive mt-1">
                  {errors.server_addr.message}
                </p>
              )}
            </div>

            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="voting"
                {...register("voting")}
                className="rounded border-gray-300"
              />
              <Label htmlFor="voting" className="text-sm font-normal">
                Voting member (participates in leader election)
              </Label>
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
              {loading ? "Adding Member..." : "Add Member"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
