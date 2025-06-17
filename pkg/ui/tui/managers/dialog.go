package managers

import (
	"fmt"

	"github.com/golem-base/seqctl/pkg/sequencer"
	"github.com/golem-base/seqctl/pkg/ui/tui/actions"
	"github.com/golem-base/seqctl/pkg/ui/tui/components"
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

// DialogManager handles dangerous action confirmations and other dialogs
type DialogManager struct {
	pages      *tview.Pages
	dialog     *components.Dialog
	flashModel *model.FlashModel
	theme      *styles.Theme
	templates  map[string]ConfirmationTemplate
}

// NewDialogManager creates a new dialog manager
func NewDialogManager(pages *tview.Pages, flashModel *model.FlashModel, theme *styles.Theme) *DialogManager {
	dm := &DialogManager{
		pages:      pages,
		dialog:     components.NewDialog(theme),
		flashModel: flashModel,
		theme:      theme,
	}

	// Initialize confirmation templates
	dm.templates = dm.createTemplates()

	return dm
}

// createTemplates creates confirmation templates using theme colors
func (dm *DialogManager) createTemplates() map[string]ConfirmationTemplate {
	return map[string]ConfirmationTemplate{
		actions.ActionNamePause: {
			Title:     "Pause Conductor",
			Message:   "Network: %s\nSequencer: %s\n\nThis will pause the conductor.",
			Dangerous: false,
		},
		actions.ActionNameHaltSequencer: {
			Title:     "Halt Sequencer",
			Message:   fmt.Sprintf("Network: %%s\nSequencer: %%s\n\n[%s]⚠️  This will stop the sequencer from producing blocks[-]", dm.theme.ErrorColor.String()),
			Dangerous: true,
		},
		actions.ActionNameOverrideLeader: {
			Title:     "Override Leader",
			Message:   fmt.Sprintf("Set leader override for sequencer %%s?\n\nNetwork: %%s\nSequencer: %%s\n\n[%s]⚠️  This will force the sequencer to act as leader regardless of cluster state[-]", dm.theme.WarningColor.String()),
			Dangerous: true,
		},
		actions.ActionNameTransferLeader: {
			Title:     "Transfer Leadership",
			Message:   fmt.Sprintf("Transfer leadership to sequencer %%s?\n\nNetwork: %%s\nTarget: %%s\n\n[%s]⚠️  This will change the current leader[-]", dm.theme.WarningColor.String()),
			Dangerous: true,
		},
		actions.ActionNameForceActive: {
			Title:     "Force Active Sequencer",
			Message:   fmt.Sprintf("Force sequencer %%s to become active?\n\nNetwork: %%s\nSequencer: %%s\n\n[%s]⚠️  This may disrupt consensus if another sequencer is active[-]", dm.theme.ErrorColor.String()),
			Dangerous: true,
		},
		actions.ActionNameRemoveServer: {
			Title:     "Remove Server",
			Message:   fmt.Sprintf("Remove sequencer %%s from cluster?\n\nNetwork: %%s\nSequencer: %%s\n\n[%s]⚠️  This operation is irreversible and will permanently remove the server[-]", dm.theme.ErrorColor.String()),
			Dangerous: true,
		},
		actions.ActionNameUpdateMembership: {
			Title:     "Update Cluster Membership",
			Message:   fmt.Sprintf("Update cluster membership from sequencer %%s?\n\nNetwork: %%s\nSequencer: %%s (Leader)\n\n[%s]⚠️  This will reconfigure the entire cluster[-]", dm.theme.WarningColor.String()),
			Dangerous: true,
		},
	}
}

// ShowActionConfirmation displays appropriate confirmation for the action
func (dm *DialogManager) ShowActionConfirmation(
	action *actions.Action,
	seq *sequencer.Sequencer,
	networkName string,
	onConfirm func(),
	onCancel func(),
) {
	// Create wrapped callbacks that handle dialog cleanup
	confirmCallback := dm.wrapCallback(onConfirm)
	cancelCallback := dm.wrapCallback(onCancel, func() {
		dm.flashModel.Info("Operation cancelled")
	})

	// Get template for this action
	if template, exists := dm.templates[action.Name]; exists {
		// Handle special case for override leader (toggle behavior)
		message := template.Message
		if action.Name == actions.ActionNameOverrideLeader && seq.Status.ConductorLeader {
			message = fmt.Sprintf("Remove leader override for sequencer %%s?\n\nNetwork: %%s\nSequencer: %%s\n\n[%s]⚠️  Removing the override requires manually restarting the op-node pod[-]", dm.theme.WarningColor.String())
		}

		// Handle special case for pause (multiple vs single)
		if action.Name == actions.ActionNamePause {
			dm.dialog.ShowPauseConfirm(seq.Config.ID, networkName, false, confirmCallback, cancelCallback)
		} else {
			// Format the message with sequencer and network info
			formattedMessage := fmt.Sprintf(message, seq.Config.ID, networkName, seq.Config.ID)
			dm.dialog.ShowConfirm(template.Title, formattedMessage, template.Dangerous, confirmCallback, cancelCallback)
		}
	} else {
		// Fallback for unknown actions
		message := fmt.Sprintf("Execute dangerous action '%s' on sequencer %s?\n\n[%s]⚠️  This operation may affect network stability[-]",
			action.Description, seq.Config.ID, dm.theme.ErrorColor.String())
		dm.dialog.ShowConfirm("Confirm Dangerous Action", message, true, confirmCallback, cancelCallback)
	}

	// Show the dialog
	dm.showDialog()
}

// wrapCallback wraps a callback to handle dialog cleanup
func (dm *DialogManager) wrapCallback(callback func(), fallback ...func()) func() {
	return func() {
		dm.hideDialog()
		if callback != nil {
			callback()
		} else if len(fallback) > 0 && fallback[0] != nil {
			fallback[0]()
		}
	}
}

// showDialog displays the confirmation dialog
func (dm *DialogManager) showDialog() {
	dm.pages.AddPage("confirmation", dm.dialog, false, true)
}

// hideDialog removes the confirmation dialog
func (dm *DialogManager) hideDialog() {
	dm.pages.RemovePage("confirmation")
}

// IsVisible returns whether the confirmation dialog is currently visible
func (dm *DialogManager) IsVisible() bool {
	frontPage, _ := dm.pages.GetFrontPage()
	return frontPage == "confirmation"
}
