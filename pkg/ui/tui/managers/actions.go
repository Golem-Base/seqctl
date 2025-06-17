package managers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/golem-base/seqctl/pkg/sequencer"
	"github.com/golem-base/seqctl/pkg/ui/tui/actions"
	"github.com/golem-base/seqctl/pkg/ui/tui/components"
	"github.com/golem-base/seqctl/pkg/ui/tui/model"
	"github.com/rivo/tview"
)

// ActionDispatcher handles action execution with proper error handling
type ActionDispatcher struct {
	appModel            *model.AppModel
	flashModel          *model.FlashModel
	app                 *tview.Application
	confirmationManager *components.ConfirmationManager
	refreshManager      *RefreshManager
	readOnlyMode        bool
	confirmDanger       bool
}

// NewActionDispatcher creates a new action dispatcher
func NewActionDispatcher(
	appModel *model.AppModel,
	flashModel *model.FlashModel,
	app *tview.Application,
	confirmationManager *components.ConfirmationManager,
	refreshManager *RefreshManager,
) *ActionDispatcher {
	return &ActionDispatcher{
		appModel:            appModel,
		flashModel:          flashModel,
		app:                 app,
		confirmationManager: confirmationManager,
		refreshManager:      refreshManager,
		readOnlyMode:        false,
		confirmDanger:       true,
	}
}

// Execute executes an action with all safety checks
func (d *ActionDispatcher) Execute(action *actions.Action, seq *sequencer.Sequencer) {
	if seq == nil {
		d.flashModel.Warning("No sequencer selected")
		return
	}

	// Check if action is enabled for this sequencer
	if action.Enabled != nil && !action.Enabled(seq) {
		d.flashModel.Warning(fmt.Sprintf("Action '%s' is not available for sequencer %s", action.Name, seq.Config.ID))
		return
	}

	// Check read-only mode (all actions are now considered write operations)
	if d.readOnlyMode {
		d.flashModel.Warning("Action not available in read-only mode")
		return
	}

	// Handle dangerous actions with confirmation
	if action.Dangerous && d.confirmDanger {
		d.showConfirmation(action, seq)
		return
	}

	// Execute the action
	d.perform(action, seq)
}

// SetReadOnlyMode sets the read-only mode
func (d *ActionDispatcher) SetReadOnlyMode(readOnly bool) {
	d.readOnlyMode = readOnly
}

// SetConfirmDanger sets whether dangerous actions require confirmation
func (d *ActionDispatcher) SetConfirmDanger(confirm bool) {
	d.confirmDanger = confirm
}

// showConfirmation shows confirmation dialog for dangerous actions
func (d *ActionDispatcher) showConfirmation(action *actions.Action, seq *sequencer.Sequencer) {
	networkName := d.appModel.GetNetwork().Name()

	d.confirmationManager.ShowActionConfirmation(
		action,
		seq,
		networkName,
		func() { d.perform(action, seq) },
		nil,
	)
}

// perform executes the action with proper error handling and feedback
func (d *ActionDispatcher) perform(action *actions.Action, seq *sequencer.Sequencer) {
	// Show feedback that action was triggered
	d.flashModel.Info(fmt.Sprintf("Executing %s...", action.Name))

	// Execute in background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		slog.Debug("Executing action",
			"action", action.Name,
			"sequencer", seq.Config.ID,
			"dangerous", action.Dangerous)

		if err := action.Handler(ctx, seq); err != nil {
			d.app.QueueUpdateDraw(func() {
				d.flashModel.Error(fmt.Sprintf("Failed to %s: %s", action.Name, err.Error()))
			})
			slog.Error("Action failed",
				"action", action.Name,
				"sequencer", seq.Config.ID,
				"error", err)
		} else {
			d.app.QueueUpdateDraw(func() {
				d.flashModel.Success(fmt.Sprintf("Successfully executed: %s", action.Name))
			})
			slog.Debug("Action completed",
				"action", action.Name,
				"sequencer", seq.Config.ID)
		}

		// Refresh data after action execution
		d.refreshManager.RefreshNow()
	}()
}
