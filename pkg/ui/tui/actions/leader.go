package actions

import (
	"context"

	"github.com/golem-base/seqctl/pkg/sequencer"
)

const ActionNameOverrideLeader = "override-leader"

// OverrideLeaderAction creates the override leader action
func OverrideLeaderAction() *Action {
	return &Action{
		Key:         'o',
		Name:        ActionNameOverrideLeader,
		Description: "Override leader status",
		Category:    "Leadership",
		Handler:     overrideLeaderHandler,
		Enabled: func(seq *sequencer.Sequencer) bool {
			return seq != nil
		},
		Opts: ActionOpts{
			Visible:   true,
			Shared:    false,
			Dangerous: true, // Leadership changes can affect consensus
			ReadOnly:  false,
		},
	}
}

// overrideLeaderHandler implements the override leader operation
func overrideLeaderHandler(ctx context.Context, seq *sequencer.Sequencer) error {
	// Toggle behavior: if already leader, remove override; otherwise set it
	removeOverride := seq.Status.ConductorLeader

	// Set/remove conductor leader override
	if err := seq.OverrideLeader(ctx, !removeOverride); err != nil {
		return err
	}

	// If setting override (not removing), also override node leader
	if !removeOverride {
		return seq.OverrideNodeLeader(ctx)
	}

	return nil
}
