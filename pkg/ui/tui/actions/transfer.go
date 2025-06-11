package actions

import (
	"context"

	"github.com/golem-base/seqctl/pkg/sequencer"
)

const ActionNameTransferLeader = "transfer-leader"

// TransferLeaderAction creates the transfer leader action
func TransferLeaderAction() *Action {
	return &Action{
		Key:         't',
		Name:        ActionNameTransferLeader,
		Description: "Transfer leadership to this sequencer",
		Category:    "Leadership",
		Handler:     transferLeaderHandler,
		Enabled: func(seq *sequencer.Sequencer) bool {
			return seq != nil && !seq.Status.ConductorLeader
		},
		Opts: ActionOpts{
			Visible:   true,
			Shared:    false,
			Dangerous: true,
			ReadOnly:  false,
		},
	}
}

// transferLeaderHandler implements the transfer leader operation
func transferLeaderHandler(ctx context.Context, seq *sequencer.Sequencer) error {
	// Transfer leadership to this sequencer
	// This requires the current leader to transfer to this sequencer
	// For the TUI context, we need the sequencer ID and address
	return seq.TransferLeaderToServer(ctx, seq.Config.ID, seq.Config.RaftAddr)
}
