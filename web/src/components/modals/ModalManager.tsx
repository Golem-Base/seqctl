import { TransferLeaderModal } from "./TransferLeaderModal";
import { ResignLeaderModal } from "./ResignLeaderModal";
import { OverrideLeaderModal } from "./OverrideLeaderModal";
import { ForceActiveModal } from "./ForceActiveModal";
import { UpdateMembershipModal } from "./UpdateMembershipModal";
import { RemoveMemberModal } from "./RemoveMemberModal";

export function ModalManager() {
  return (
    <>
      <TransferLeaderModal />
      <ResignLeaderModal />
      <OverrideLeaderModal />
      <ForceActiveModal />
      <UpdateMembershipModal />
      <RemoveMemberModal />
    </>
  );
}
