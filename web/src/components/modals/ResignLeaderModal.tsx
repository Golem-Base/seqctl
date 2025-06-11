import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { useActionStore } from "@/stores/actionStore";
import { useSequencer } from "@/stores/networkStore";
import { AlertCircle, Crown } from "lucide-react";

export function ResignLeaderModal() {
  const {
    activeModal,
    targetSequencerId,
    closeModal,
    resignLeader,
    loading,
    error,
  } = useActionStore();
  const { sequencer } = useSequencer(targetSequencerId || "");

  const isOpen = activeModal === "resign-leader";

  const handleConfirm = async () => {
    try {
      await resignLeader();
    } catch (error) {
      // Error is handled in the store
    }
  };

  if (!sequencer) return null;

  return (
    <Dialog open={isOpen} onOpenChange={closeModal}>
      <DialogContent className="sm:max-w-[450px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Crown className="h-5 w-5 text-yellow-500" />
            Resign Leadership
          </DialogTitle>
          <DialogDescription>
            Make {sequencer.id} resign as the Raft leader
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="rounded-md bg-yellow-50 dark:bg-yellow-900/20 p-4">
            <div className="flex">
              <AlertCircle className="h-5 w-5 text-yellow-600 dark:text-yellow-400 mr-2 flex-shrink-0 mt-0.5" />
              <div className="text-sm text-yellow-800 dark:text-yellow-200">
                <p className="font-medium mb-1">Warning:</p>
                <p>
                  This will trigger a new leader election. The cluster will be
                  temporarily without a leader until a new one is elected.
                </p>
              </div>
            </div>
          </div>

          <div className="text-sm text-muted-foreground">
            <p>
              Current leader:{" "}
              <span className="font-medium text-foreground">
                {sequencer.id}
              </span>
            </p>
            <p>
              Raft address:{" "}
              <span className="font-mono">{sequencer.raft_addr}</span>
            </p>
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
            {loading ? "Resigning..." : "Resign Leadership"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
