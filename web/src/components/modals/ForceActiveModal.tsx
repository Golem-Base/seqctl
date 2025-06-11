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
import { AlertTriangle, Zap } from "lucide-react";

const forceActiveSchema = z.object({
  block_hash: z.string().optional(),
});

type ForceActiveForm = z.infer<typeof forceActiveSchema>;

export function ForceActiveModal() {
  const {
    activeModal,
    targetSequencerId,
    closeModal,
    forceActive,
    loading,
    error,
  } = useActionStore();
  const { sequencer } = useSequencer(targetSequencerId || "");

  const isOpen = activeModal === "force-active";

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ForceActiveForm>({
    resolver: zodResolver(forceActiveSchema),
  });

  const onSubmit = async (data: ForceActiveForm) => {
    try {
      await forceActive(data);
      reset();
    } catch (error) {
      // Error is handled in the store
    }
  };

  const handleClose = () => {
    closeModal();
    reset();
  };

  if (!sequencer) return null;

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Zap className="h-5 w-5 text-yellow-500" />
            Force Sequencer Active
          </DialogTitle>
          <DialogDescription>
            Force {sequencer.id} to become the active sequencer
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="rounded-md bg-orange-50 dark:bg-orange-900/20 p-4">
            <div className="flex">
              <AlertTriangle className="h-5 w-5 text-orange-600 dark:text-orange-400 mr-2 flex-shrink-0 mt-0.5" />
              <div className="text-sm text-orange-800 dark:text-orange-200">
                <p className="font-medium mb-1">Emergency Use Only!</p>
                <p className="mb-2">
                  This operation bypasses normal sequencer selection and can
                  cause:
                </p>
                <ul className="list-disc list-inside ml-2">
                  <li>State inconsistencies if not properly synchronized</li>
                  <li>Transaction conflicts with other active sequencers</li>
                  <li>Potential data loss if the wrong block hash is used</li>
                </ul>
              </div>
            </div>
          </div>

          <div className="space-y-4">
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
                  Currently Active:{" "}
                  <span className="font-medium text-foreground">
                    {sequencer.sequencer_active ? "Yes" : "No"}
                  </span>
                </p>
                <p>
                  Unsafe L2 Block:{" "}
                  <span className="font-mono text-foreground">
                    {sequencer.unsafe_l2}
                  </span>
                </p>
              </div>
            </div>

            <div>
              <Label htmlFor="block_hash">Starting Block Hash (Optional)</Label>
              <Input
                id="block_hash"
                {...register("block_hash")}
                placeholder="0x... (leave empty to use zero hash)"
                className="font-mono"
              />
              <p className="text-xs text-muted-foreground mt-1">
                Specify a block hash to start from, or leave empty to use the
                zero hash
              </p>
              {errors.block_hash && (
                <p className="text-sm text-destructive mt-1">
                  {errors.block_hash.message}
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
            <Button
              type="submit"
              variant="destructive"
              disabled={loading}
            >
              {loading ? "Forcing Active..." : "Force Active"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
