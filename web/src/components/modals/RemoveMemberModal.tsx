import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useActionStore } from "@/stores/actionStore";
import { useSequencer } from "@/stores/networkStore";
import { AlertTriangle, UserMinus } from "lucide-react";

export function RemoveMemberModal() {
  const {
    activeModal,
    targetSequencerId,
    closeModal,
    removeMember,
    loading,
    error,
  } = useActionStore();
  const { sequencer, network } = useSequencer(targetSequencerId || "");
  const [selectedServerId, setSelectedServerId] = useState("");

  const isOpen = activeModal === "remove-member";

  const handleConfirm = async () => {
    if (!selectedServerId) return;

    try {
      await removeMember({ server_id: selectedServerId });
      setSelectedServerId("");
    } catch (error) {
      // Error is handled in the store
    }
  };

  const handleClose = () => {
    closeModal();
    setSelectedServerId("");
  };

  if (!sequencer || !network) return null;

  // Get removable members (all except the current leader)
  const removableMembers = network.sequencers.filter((s) =>
    s.id !== sequencer.id
  );

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <UserMinus className="h-5 w-5 text-red-500" />
            Remove Cluster Member
          </DialogTitle>
          <DialogDescription>
            Remove a server from the Raft cluster
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="rounded-md bg-red-50 dark:bg-red-900/20 p-4">
            <div className="flex">
              <AlertTriangle className="h-5 w-5 text-red-600 dark:text-red-400 mr-2 flex-shrink-0 mt-0.5" />
              <div className="text-sm text-red-800 dark:text-red-200">
                <p className="font-medium mb-1">Warning:</p>
                <p>Removing a member from the cluster will:</p>
                <ul className="list-disc list-inside mt-1 ml-2">
                  <li>Stop replication to that server</li>
                  <li>Remove its voting rights (if applicable)</li>
                  <li>Potentially affect cluster fault tolerance</li>
                </ul>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <div>
              <Label>Current Cluster Size</Label>
              <p className="text-sm text-muted-foreground mt-1">
                {network.sequencers.length}{" "}
                members ({network.sequencers.filter((s) => s.voting).length}
                {" "}
                voting)
              </p>
            </div>

            <div>
              <Label htmlFor="member">Select Member to Remove</Label>
              <Select
                value={selectedServerId}
                onValueChange={setSelectedServerId}
              >
                <SelectTrigger id="member">
                  <SelectValue placeholder="Select a member" />
                </SelectTrigger>
                <SelectContent>
                  {removableMembers.map((member) => (
                    <SelectItem key={member.id} value={member.id}>
                      <div className="flex items-center justify-between w-full">
                        <span className="font-mono">{member.id}</span>
                        <span className="text-xs text-muted-foreground ml-2">
                          {member.voting ? "Voting" : "Non-voting"}
                          {member.conductor_leader && " • Leader"}
                        </span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {selectedServerId && (
              <div className="rounded-md bg-muted p-3">
                <p className="text-sm">
                  Selected:{" "}
                  <span className="font-mono font-medium">
                    {selectedServerId}
                  </span>
                </p>
                {removableMembers.find((m) => m.id === selectedServerId)
                  ?.conductor_leader && (
                  <p className="text-sm text-yellow-600 dark:text-yellow-400 mt-1">
                    ⚠️ This member is currently the leader. Consider
                    transferring leadership first.
                  </p>
                )}
              </div>
            )}
          </div>

          {error && (
            <div className="rounded-md bg-destructive/15 p-3">
              <p className="text-sm text-destructive">{error}</p>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleConfirm}
            disabled={loading || !selectedServerId}
          >
            {loading ? "Removing..." : "Remove Member"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
