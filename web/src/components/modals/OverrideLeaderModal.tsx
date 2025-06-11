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
import { useActionStore } from "@/stores/actionStore";
import { useSequencer } from "@/stores/networkStore";
import { AlertTriangle, Shield } from "lucide-react";

export function OverrideLeaderModal() {
  const {
    activeModal,
    targetSequencerId,
    closeModal,
    overrideLeader,
    loading,
    error,
  } = useActionStore();
  const { sequencer } = useSequencer(targetSequencerId || "");
  const [override, setOverride] = useState(true);

  const isOpen = activeModal === "override-leader";

  const handleConfirm = async () => {
    try {
      await overrideLeader({ override });
    } catch (error) {
      // Error is handled in the store
    }
  };

  if (!sequencer) return null;

  return (
    <Dialog open={isOpen} onOpenChange={closeModal}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5 text-orange-500" />
            Override Leader Status
          </DialogTitle>
          <DialogDescription>
            Force override the leader status of {sequencer.id}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="rounded-md bg-red-50 dark:bg-red-900/20 p-4">
            <div className="flex">
              <AlertTriangle className="h-5 w-5 text-red-600 dark:text-red-400 mr-2 flex-shrink-0 mt-0.5" />
              <div className="text-sm text-red-800 dark:text-red-200">
                <p className="font-medium mb-1">Danger: Split-Brain Risk!</p>
                <p className="mb-2">
                  This operation can cause a split-brain scenario where multiple
                  nodes believe they are the leader.
                </p>
                <p>Only use this in emergency situations when:</p>
                <ul className="list-disc list-inside mt-1 ml-2">
                  <li>The cluster is in an unrecoverable state</li>
                  <li>Normal leader election has failed</li>
                  <li>You fully understand the consequences</li>
                </ul>
              </div>
            </div>
          </div>

          <div className="space-y-3">
            <div>
              <Label>Current Status</Label>
              <div className="text-sm text-muted-foreground mt-1">
                <p>
                  Sequencer:{" "}
                  <span className="font-medium text-foreground">
                    {sequencer.id}
                  </span>
                </p>
                <p>
                  Is Leader:{" "}
                  <span className="font-medium text-foreground">
                    {sequencer.conductor_leader ? "Yes" : "No"}
                  </span>
                </p>
                <p>
                  Is Active:{" "}
                  <span className="font-medium text-foreground">
                    {sequencer.conductor_active ? "Yes" : "No"}
                  </span>
                </p>
              </div>
            </div>

            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="override"
                checked={override}
                onChange={(e) => setOverride(e.target.checked)}
                className="rounded border-gray-300"
              />
              <Label htmlFor="override" className="text-sm font-normal">
                Set as leader (unchecking will remove leader status)
              </Label>
            </div>
          </div>

          {error && (
            <div className="rounded-md bg-destructive/15 p-3">
              <p className="text-sm text-destructive">{error}</p>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={closeModal}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleConfirm}
            disabled={loading}
          >
            {loading ? "Overriding..." : "Override Leader Status"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
