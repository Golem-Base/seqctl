package actions

import (
	"context"
	"fmt"
	"sync"

	"github.com/golem-base/seqctl/pkg/sequencer"
)

// ActionHandler is a function that performs an action on a sequencer
type ActionHandler func(ctx context.Context, seq *sequencer.Sequencer) error

// ActionOpts tracks various action options
type ActionOpts struct {
	Visible   bool // Show in help/operations panel
	Shared    bool // Available across multiple views
	Dangerous bool // Requires confirmation
	ReadOnly  bool // Available in read-only mode
}

// Action represents an operation that can be performed on a sequencer
type Action struct {
	Key         rune
	Name        string
	Description string
	Category    string
	Handler     ActionHandler
	Enabled     func(*sequencer.Sequencer) bool
	Opts        ActionOpts
}

// ActionRegistry manages available actions
type ActionRegistry struct {
	actions map[string]*Action
	byKey   map[rune]*Action
	ordered []*Action
	mu      sync.RWMutex
}

// NewActionRegistry creates a new action registry
func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{
		actions: make(map[string]*Action),
		byKey:   make(map[rune]*Action),
		ordered: make([]*Action, 0),
	}
}

// Register adds an action to the registry
func (r *ActionRegistry) Register(action *Action) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.actions[action.Name]; exists {
		return fmt.Errorf("action %s already registered", action.Name)
	}

	if _, exists := r.byKey[action.Key]; exists {
		return fmt.Errorf("key %c already registered", action.Key)
	}

	r.actions[action.Name] = action
	r.byKey[action.Key] = action
	r.ordered = append(r.ordered, action)

	return nil
}

// GetByKey returns an action by its keyboard shortcut
func (r *ActionRegistry) GetByKey(key rune) *Action {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.byKey[key]
}

// GetByName returns an action by its name
func (r *ActionRegistry) GetByName(name string) *Action {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.actions[name]
}

// GetAll returns all registered actions in registration order
func (r *ActionRegistry) GetAll() []*Action {
	r.mu.RLock()
	defer r.mu.RUnlock()
	actions := make([]*Action, len(r.ordered))
	copy(actions, r.ordered)
	return actions
}

// GetEnabled returns all actions that are enabled for the given sequencer in registration order
func (r *ActionRegistry) GetEnabled(seq *sequencer.Sequencer) []*Action {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if seq == nil {
		return nil
	}

	enabled := make([]*Action, 0)
	for _, action := range r.ordered {
		if action.Enabled == nil || action.Enabled(seq) {
			enabled = append(enabled, action)
		}
	}
	return enabled
}

// GetCategories returns all unique action categories
func (r *ActionRegistry) GetCategories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categoryMap := make(map[string]bool)
	for _, action := range r.actions {
		if action.Category != "" {
			categoryMap[action.Category] = true
		}
	}

	categories := make([]string, 0, len(categoryMap))
	for category := range categoryMap {
		categories = append(categories, category)
	}
	return categories
}

// GetByCategory returns all actions in a specific category in registration order
func (r *ActionRegistry) GetByCategory(category string) []*Action {
	r.mu.RLock()
	defer r.mu.RUnlock()

	actions := make([]*Action, 0)
	for _, action := range r.ordered {
		if action.Category == category {
			actions = append(actions, action)
		}
	}
	return actions
}

// GetVisible returns all visible actions in registration order
func (r *ActionRegistry) GetVisible() []*Action {
	r.mu.RLock()
	defer r.mu.RUnlock()

	visible := make([]*Action, 0)
	for _, action := range r.ordered {
		if action.Opts.Visible {
			visible = append(visible, action)
		}
	}
	return visible
}
