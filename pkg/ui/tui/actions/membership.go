package actions

import (
	"context"

	"github.com/golem-base/seqctl/pkg/sequencer"
)

const ActionNameUpdateMembership = "update-membership"

// UpdateClusterMembershipAction creates the update cluster membership action
func UpdateClusterMembershipAction() *Action {
	return &Action{
		Key:         'u',
		Name:        ActionNameUpdateMembership,
		Description: "Update cluster membership",
		Handler:     updateClusterMembershipHandler,
		Enabled: func(seq *sequencer.Sequencer) bool {
			return seq != nil && seq.Status.ConductorLeader
		},
		Dangerous: true,
	}
}

// updateClusterMembershipHandler implements the update cluster membership operation
func updateClusterMembershipHandler(ctx context.Context, seq *sequencer.Sequencer) error {
	// TODO: Implement full cluster membership update when we have network context
	return nil
}
