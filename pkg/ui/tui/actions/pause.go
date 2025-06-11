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
		Category:    "Control",
		Handler:     pauseHandler,
		Enabled: func(seq *sequencer.Sequencer) bool {
			return seq != nil && seq.Status.ConductorActive
		},
		Opts: ActionOpts{
			Visible:   true,
			Shared:    false,
			Dangerous: true,
			ReadOnly:  false,
		},
	}
}

// pauseHandler implements the pause operation
func pauseHandler(ctx context.Context, seq *sequencer.Sequencer) error {
	return seq.Pause(ctx)
}
