package actions

import (
	"context"

	"github.com/golem-base/seqctl/pkg/sequencer"
)

const ActionNameForceActive = "force-active"

// ForceActiveSequencerAction creates the force active sequencer action
func ForceActiveSequencerAction() *Action {
	return &Action{
		Key:         'f',
		Name:        ActionNameForceActive,
		Description: "Force sequencer to become active",
		Handler:     forceActiveSequencerHandler,
		Enabled: func(seq *sequencer.Sequencer) bool {
			return seq != nil && !seq.Status.SequencerActive
		},
		Dangerous: true,
	}
}

// forceActiveSequencerHandler implements the force active sequencer operation
func forceActiveSequencerHandler(ctx context.Context, seq *sequencer.Sequencer) error {
	return seq.StartSequencer(ctx, seq.Status.UnsafeL2.Hash)
}
