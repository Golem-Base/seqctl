package actions

import (
	"context"

	"github.com/golem-base/seqctl/pkg/sequencer"
)

const ActionNameResume = "resume"

// ResumeAction creates the resume action
func ResumeAction() *Action {
	return &Action{
		Key:         's',
		Name:        ActionNameResume,
		Description: "Resume conductor",
		Category:    "Control",
		Handler:     resumeHandler,
		Enabled: func(seq *sequencer.Sequencer) bool {
			return seq != nil && !seq.Status.ConductorActive
		},
		Opts: ActionOpts{
			Visible:   true,
			Shared:    false,
			Dangerous: false,
			ReadOnly:  false,
		},
	}
}

// resumeHandler implements the resume operation
func resumeHandler(ctx context.Context, seq *sequencer.Sequencer) error {
	return seq.Resume(ctx)
}
