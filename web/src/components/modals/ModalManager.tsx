import { useActionStore } from '@/stores/actionStore';
import { TransferLeaderModal } from './TransferLeaderModal';
import { ResignLeaderModal } from './ResignLeaderModal';
import { OverrideLeaderModal } from './OverrideLeaderModal';
import { ForceActiveModal } from './ForceActiveModal';
import { UpdateMembershipModal } from './UpdateMembershipModal';
import { RemoveMemberModal } from './RemoveMemberModal';

export function ModalManager() {
  const { activeModal } = useActionStore();

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