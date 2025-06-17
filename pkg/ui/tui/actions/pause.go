package actions

import (
	"context"

	"github.com/golem-base/seqctl/pkg/sequencer"
)

const ActionNamePause = "pause"

// PauseAction creates the pause action
func PauseAction() *Action {
	return &Action{
		Key:         'p',
		Name:        ActionNamePause,
		Description: "Pause conductor",
		Handler:     pauseHandler,
		Enabled: func(seq *sequencer.Sequencer) bool {
			return seq != nil && seq.Status.ConductorActive
		},
		Dangerous: true,
	}
}

// pauseHandler implements the pause operation
func pauseHandler(ctx context.Context, seq *sequencer.Sequencer) error {
	return seq.Pause(ctx)
}
