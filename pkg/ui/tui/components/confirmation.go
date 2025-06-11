package components

import (
	"fmt"

	"github.com/golem-base/seqctl/pkg/sequencer"
	"github.com/golem-base/seqctl/pkg/ui/tui/actions"
	"github.com/golem-base/seqctl/pkg/ui/tui/model"
	"github.com/golem-base/seqctl/pkg/ui/tui/styles"
	"github.com/rivo/tview"
)

// ConfirmationTemplate holds the template for an action confirmation
type ConfirmationTemplate struct {
	Title     string
	Message   string
	Dangerous bool
}

// ConfirmationManager handles dangerous action confirmations
type ConfirmationManager struct {
	pages      *tview.Pages
	dialog     *Dialog
	flashModel *model.FlashModel
	templates  map[string]ConfirmationTemplate
}

// NewConfirmationManager creates a new confirmation manager
func NewConfirmationManager(pages *tview.Pages, flashModel *model.FlashModel, theme *styles.Theme) *ConfirmationManager {
	cm := &ConfirmationManager{
		pages:      pages,
		dialog:     NewDialog(theme),
		flashModel: flashModel,
	}

	// Initialize confirmation templates
	cm.templates = map[string]ConfirmationTemplate{
		actions.ActionNamePause: {
			Title:     "Pause Conductor",
			Message:   "Network: %s\nSequencer: %s\n\nThis will pause the conductor.",
			Dangerous: false,
		},
		actions.ActionNameHaltSequencer: {
			Title:     "Halt Sequencer",
			Message:   "Network: %s\nSequencer: %s\n\n[red]⚠️  This will stop the sequencer from producing blocks[-]",
			Dangerous: true,
		},
		actions.ActionNameOverrideLeader: {
			Title:     "Override Leader",
			Message:   "Set leader override for sequencer %s?\n\nNetwork: %s\nSequencer: %s\n\n[orange]⚠️  This will force the sequencer to act as leader regardless of cluster state[-]",
			Dangerous: true,
		},
		actions.ActionNameTransferLeader: {
			Title:     "Transfer Leadership",
			Message:   "Transfer leadership to sequencer %s?\n\nNetwork: %s\nTarget: %s\n\n[orange]⚠️  This will change the current leader[-]",
			Dangerous: true,
		},
		actions.ActionNameForceActive: {
			Title:     "Force Active Sequencer",
			Message:   "Force sequencer %s to become active?\n\nNetwork: %s\nSequencer: %s\n\n[red]⚠️  This may disrupt consensus if another sequencer is active[-]",
			Dangerous: true,
		},
		actions.ActionNameRemoveServer: {
			Title:     "Remove Server",
			Message:   "Remove sequencer %s from cluster?\n\nNetwork: %s\nSequencer: %s\n\n[red]⚠️  This operation is irreversible and will permanently remove the server[-]",
			Dangerous: true,
		},
		actions.ActionNameUpdateMembership: {
			Title:     "Update Cluster Membership",
			Message:   "Update cluster membership from sequencer %s?\n\nNetwork: %s\nSequencer: %s (Leader)\n\n[orange]⚠️  This will reconfigure the entire cluster[-]",
			Dangerous: true,
		},
	}

	return cm
}

// ShowActionConfirmation displays appropriate confirmation for the action
func (cm *ConfirmationManager) ShowActionConfirmation(
	action *actions.Action,
	seq *sequencer.Sequencer,
	networkName string,
	onConfirm func(),
	onCancel func(),
) {
	// Create wrapped callbacks that handle dialog cleanup
	confirmCallback := cm.wrapCallback(onConfirm)
	cancelCallback := cm.wrapCallback(onCancel, func() {
		cm.flashModel.Info("Operation cancelled")
	})

	// Get template for this action
	if template, exists := cm.templates[action.Name]; exists {
		// Handle special case for override leader (toggle behavior)
		message := template.Message
		if action.Name == actions.ActionNameOverrideLeader && seq.Status.ConductorLeader {
			message = "Remove leader override for sequencer %s?\n\nNetwork: %s\nSequencer: %s\n\n[orange]⚠️  Removing the override requires manually restarting the op-node pod[-]"
		}

		// Handle special case for pause (multiple vs single)
		if action.Name == actions.ActionNamePause {
			cm.dialog.ShowPauseConfirm(seq.Config.ID, networkName, false, confirmCallback, cancelCallback)
		} else {
			// Format the message with sequencer and network info
			formattedMessage := fmt.Sprintf(message, seq.Config.ID, networkName, seq.Config.ID)
			cm.dialog.ShowConfirm(template.Title, formattedMessage, template.Dangerous, confirmCallback, cancelCallback)
		}
	} else {
		// Fallback for unknown actions
		message := fmt.Sprintf("Execute dangerous action '%s' on sequencer %s?\n\n[red]⚠️  This operation may affect network stability[-]",
			action.Description, seq.Config.ID)
		cm.dialog.ShowConfirm("Confirm Dangerous Action", message, true, confirmCallback, cancelCallback)
	}

	// Show the dialog
	cm.showDialog()
}

// wrapCallback wraps a callback to handle dialog cleanup
func (cm *ConfirmationManager) wrapCallback(callback func(), fallback ...func()) func() {
	return func() {
		cm.hideDialog()
		if callback != nil {
			callback()
		} else if len(fallback) > 0 && fallback[0] != nil {
			fallback[0]()
		}
	}
}

// showDialog displays the confirmation dialog
func (cm *ConfirmationManager) showDialog() {
	cm.pages.AddPage("confirmation", cm.dialog, false, true)
}

// hideDialog removes the confirmation dialog
func (cm *ConfirmationManager) hideDialog() {
	cm.pages.RemovePage("confirmation")
}

// IsVisible returns whether the confirmation dialog is currently visible
func (cm *ConfirmationManager) IsVisible() bool {
	frontPage, _ := cm.pages.GetFrontPage()
	return frontPage == "confirmation"
}
