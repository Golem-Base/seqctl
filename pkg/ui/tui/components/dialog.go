package components

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/golem-base/seqctl/pkg/ui/tui/styles"
	"github.com/rivo/tview"
)

// DialogType represents the type of dialog
type DialogType int

const (
	DialogTypeConfirm DialogType = iota
	DialogTypeInfo
	DialogTypeError
	DialogTypeWarning
)

// DialogConfig holds configuration for showing a dialog
type DialogConfig struct {
	Type      DialogType
	Title     string
	Message   string
	Dangerous bool
	OnConfirm func()
	OnCancel  func()
	OnClose   func()
}

// Dialog is a reusable dialog component
type Dialog struct {
	*tview.Modal
	theme *styles.Theme
}

// NewDialog creates a new dialog component
func NewDialog(theme *styles.Theme) *Dialog {
	dialog := &Dialog{
		Modal: tview.NewModal(),
		theme: theme,
	}

	dialog.Modal.SetBackgroundColor(tcell.ColorDefault)
	return dialog
}

// Show displays a dialog based on the provided configuration
func (d *Dialog) Show(config DialogConfig) {
	var text string
	var buttons []string
	var doneFunc func(int, string)

	// Format title based on type
	switch config.Type {
	case DialogTypeError:
		text = fmt.Sprintf("[%s][::b]%s[::-][-]\n\n%s", d.theme.ErrorColor.String(), config.Title, config.Message)
	case DialogTypeWarning:
		text = fmt.Sprintf("[%s][::b]%s[::-][-]\n\n%s", d.theme.WarningColor.String(), config.Title, config.Message)
	default:
		text = fmt.Sprintf("[::b]%s[::-]\n\n%s", config.Title, config.Message)
	}

	// Add danger warning for confirm dialogs
	if config.Type == DialogTypeConfirm && config.Dangerous {
		text += fmt.Sprintf("\n\n[%s]⚠️  This is a potentially dangerous operation[-]", d.theme.ErrorColor.String())
	}

	// Setup buttons and callbacks based on type
	switch config.Type {
	case DialogTypeConfirm:
		buttons = []string{"Confirm", "Cancel"}
		doneFunc = func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Confirm" && config.OnConfirm != nil {
				config.OnConfirm()
			} else if config.OnCancel != nil {
				config.OnCancel()
			}
		}
	default:
		buttons = []string{"OK"}
		doneFunc = func(buttonIndex int, buttonLabel string) {
			if config.OnClose != nil {
				config.OnClose()
			}
		}
	}

	d.Modal.
		SetText(text).
		ClearButtons().
		AddButtons(buttons).
		SetDoneFunc(doneFunc)
}

// Convenience methods for backward compatibility and cleaner API

// ShowConfirm displays a confirmation dialog
func (d *Dialog) ShowConfirm(title, message string, dangerous bool, onConfirm, onCancel func()) {
	d.Show(DialogConfig{
		Type:      DialogTypeConfirm,
		Title:     title,
		Message:   message,
		Dangerous: dangerous,
		OnConfirm: onConfirm,
		OnCancel:  onCancel,
	})
}

// ShowInfo displays an information dialog
func (d *Dialog) ShowInfo(title, message string, onClose func()) {
	d.Show(DialogConfig{
		Type:    DialogTypeInfo,
		Title:   title,
		Message: message,
		OnClose: onClose,
	})
}

// ShowError displays an error dialog
func (d *Dialog) ShowError(title, message string, onClose func()) {
	d.Show(DialogConfig{
		Type:    DialogTypeError,
		Title:   title,
		Message: message,
		OnClose: onClose,
	})
}

// ShowWarning displays a warning dialog
func (d *Dialog) ShowWarning(title, message string, onClose func()) {
	d.Show(DialogConfig{
		Type:    DialogTypeWarning,
		Title:   title,
		Message: message,
		OnClose: onClose,
	})
}

// Operation-specific confirmation methods

// ShowPauseConfirm shows a confirmation dialog for pause operation
func (d *Dialog) ShowPauseConfirm(sequencerID, networkName string, multiple bool, onConfirm, onCancel func()) {
	var message string
	if multiple {
		message = fmt.Sprintf("This will pause ALL conductors in network %s!\n\nAre you sure you want to continue?", networkName)
	} else {
		message = fmt.Sprintf("Network: %s\nSequencer: %s\n\nThis will pause the conductor.", networkName, sequencerID)
	}

	d.ShowConfirm("Pause Conductor", message, multiple, onConfirm, onCancel)
}

// ShowResumeConfirm shows a confirmation dialog for resume operation
func (d *Dialog) ShowResumeConfirm(sequencerID, networkName string, onConfirm, onCancel func()) {
	message := fmt.Sprintf("Network: %s\nSequencer: %s\n\nThis will resume the conductor.", networkName, sequencerID)
	d.ShowConfirm("Resume Conductor", message, false, onConfirm, onCancel)
}

// ShowOverrideLeaderConfirm shows a confirmation dialog for override leader operation
func (d *Dialog) ShowOverrideLeaderConfirm(sequencerID, networkName string, isLeader bool, onConfirm, onCancel func()) {
	var message string
	if isLeader {
		message = fmt.Sprintf("Remove leader override for sequencer %s?\n\nNetwork: %s\nSequencer: %s\n\n[%s]⚠️  Removing the override requires manually restarting the op-node pod[-]",
			sequencerID, networkName, sequencerID, d.theme.WarningColor.String())
	} else {
		message = fmt.Sprintf("Set leader override for sequencer %s?\n\nNetwork: %s\nSequencer: %s\n\n[%s]⚠️  This will force the sequencer to act as leader regardless of cluster state[-]",
			sequencerID, networkName, sequencerID, d.theme.WarningColor.String())
	}

	d.ShowConfirm("Override Leader", message, true, onConfirm, onCancel)
}

// ShowHaltConfirm shows a confirmation dialog for halt operation
func (d *Dialog) ShowHaltConfirm(sequencerID, networkName string, onConfirm, onCancel func()) {
	message := fmt.Sprintf("Network: %s\nSequencer: %s\n\n[%s]⚠️  This will stop the sequencer from producing blocks[-]",
		networkName, sequencerID, d.theme.ErrorColor.String())
	d.ShowConfirm("Halt Sequencer", message, true, onConfirm, onCancel)
}
