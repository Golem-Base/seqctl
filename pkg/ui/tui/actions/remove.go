package actions

import (
	"context"

	"github.com/golem-base/seqctl/pkg/sequencer"
)

const ActionNameRemoveServer = "remove-server"

// RemoveServerAction creates the remove server action
func RemoveServerAction() *Action {
	return &Action{
		Key:         'd',
		Name:        ActionNameRemoveServer,
		Description: "Remove sequencer from cluster",
		Handler:     removeServerHandler,
		Enabled: func(seq *sequencer.Sequencer) bool {
			return seq != nil && !seq.Status.ConductorLeader
		},
		Dangerous: true,
	}
}

// removeServerHandler implements the remove server operation
func removeServerHandler(ctx context.Context, seq *sequencer.Sequencer) error {
	// Remove this sequencer from the cluster
	// This should only be called on a leader sequencer
	return seq.RemoveServer(ctx, seq.Config.ID)
}
