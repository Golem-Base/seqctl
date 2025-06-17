package actions

import (
	"context"

	"github.com/golem-base/seqctl/pkg/sequencer"
)

// ActionHandler is a function that performs an action on a sequencer
type ActionHandler func(ctx context.Context, seq *sequencer.Sequencer) error

// Action represents an operation that can be performed on a sequencer
type Action struct {
	Key         rune
	Name        string
	Description string
	Handler     ActionHandler
	Enabled     func(*sequencer.Sequencer) bool
	Dangerous   bool // Requires confirmation
}

// All available actions for the sequencers
var AllActions = []*Action{
	PauseAction(),
	ResumeAction(),
	OverrideLeaderAction(),
	HaltSequencerAction(),
	TransferLeaderAction(),
	ForceActiveSequencerAction(),
	RemoveServerAction(),
	UpdateClusterMembershipAction(),
}

// GetActionByKey returns an action by its keyboard shortcut
func GetActionByKey(key rune) *Action {
	for _, action := range AllActions {
		if action.Key == key {
			return action
		}
	}
	return nil
}

// GetActionByName returns an action by its name
func GetActionByName(name string) *Action {
	for _, action := range AllActions {
		if action.Name == name {
			return action
		}
	}
	return nil
}

// GetVisibleActions returns all actions (all are visible by default)
func GetVisibleActions() []*Action {
	return AllActions
}

// GetEnabledActions returns all actions that are enabled for the given sequencer
func GetEnabledActions(seq *sequencer.Sequencer) []*Action {
	if seq == nil {
		return nil
	}

	enabled := make([]*Action, 0, len(AllActions))
	for _, action := range AllActions {
		if action.Enabled == nil || action.Enabled(seq) {
			enabled = append(enabled, action)
		}
	}
	return enabled
}
