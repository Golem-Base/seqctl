package actions

import (
	"context"

	"github.com/golem-base/seqctl/pkg/sequencer"
)

const ActionNameHaltSequencer = "halt-sequencer"

// HaltSequencerAction creates the halt sequencer action
func HaltSequencerAction() *Action {
	return &Action{
		Key:         'h',
		Name:        ActionNameHaltSequencer,
		Description: "Halt sequencer",
		Category:    "Control",
		Handler:     haltSequencerHandler,
		Enabled: func(seq *sequencer.Sequencer) bool {
			return seq != nil && seq.Status.SequencerActive
		},
		Opts: ActionOpts{
			Visible:   true,
			Shared:    false,
			Dangerous: true,
			ReadOnly:  false,
		},
	}
}

// haltSequencerHandler implements the halt sequencer operation
func haltSequencerHandler(ctx context.Context, seq *sequencer.Sequencer) error {
	_, err := seq.StopSequencer(ctx)
	return err
}
