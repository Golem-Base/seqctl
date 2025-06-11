package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/golem-base/seqctl/pkg/sequencer"
	"github.com/golem-base/seqctl/pkg/ui/tui/model"
	"github.com/golem-base/seqctl/pkg/ui/tui/styles"
	"github.com/rivo/tview"
)

// DetailsPanel displays detailed information about a selected sequencer
type DetailsPanel struct {
	*tview.TextView

	model   *model.AppModel
	theme   *styles.Theme
	current *sequencer.Sequencer
}

// NewDetailsPanel creates a new details panel component
func NewDetailsPanel(appModel *model.AppModel, theme *styles.Theme) *DetailsPanel {
	panel := &DetailsPanel{
		TextView: tview.NewTextView(),
		model:    appModel,
		theme:    theme,
	}

	// Configure TextView
	panel.TextView.
		SetDynamicColors(true).
		SetScrollable(true).
		SetBorderPadding(1, 1, 1, 1).
		SetBackgroundColor(theme.BackgroundColor)

	// Set initial text
	panel.updateContent(nil)

	// Register as model listener
	appModel.AddListener(panel)

	return panel
}

// updateContent updates the panel content
func (d *DetailsPanel) updateContent(seq *sequencer.Sequencer) {
	if seq == nil {
		d.TextView.SetText("[dim]No sequencer selected[-]")
		d.current = nil
		return
	}

	d.current = seq

	var details strings.Builder

	// Basic info
	details.WriteString(fmt.Sprintf("[aqua]ID:[-] %s\n", seq.Config.ID))

	// Status section
	details.WriteString("\n[aqua]Status:[-]\n")
	statusItems := []struct {
		label string
		value bool
	}{
		{"Conductor Active", seq.Status.ConductorActive},
		{"Conductor Leader", seq.Status.ConductorLeader},
		{"Sequencer Healthy", seq.Status.SequencerHealthy},
		{"Sequencer Active", seq.Status.SequencerActive},
	}

	for _, item := range statusItems {
		details.WriteString(fmt.Sprintf("  %s: %s\n", item.label, d.formatBooleanStatus(item.value)))
	}

	// Configuration section
	details.WriteString("\n[aqua]Configuration:[-]\n")
	details.WriteString(fmt.Sprintf("  Voting: %s\n", d.formatBooleanStatus(seq.Config.Voting)))
	details.WriteString(fmt.Sprintf("  Timeout: %s\n", seq.Config.Timeout.String()))

	// Network endpoints
	details.WriteString("\n[aqua]Network Endpoints:[-]\n")
	details.WriteString(fmt.Sprintf("  Conductor RPC: %s\n", seq.Config.ConductorRPCURL))
	details.WriteString(fmt.Sprintf("  Node RPC: %s\n", seq.Config.NodeRPCURL))
	details.WriteString(fmt.Sprintf("  Raft Address: %s\n", seq.Config.RaftAddr))

	// Block information if available
	if seq.Status.UnsafeL2 != nil {
		details.WriteString("\n[aqua]Block Information:[-]\n")
		details.WriteString(fmt.Sprintf("  Number: %d\n", seq.Status.UnsafeL2.Number))
		details.WriteString(fmt.Sprintf("  Hash: %s\n", seq.Status.UnsafeL2.Hash.String()))
		details.WriteString(fmt.Sprintf("  Parent Hash: %s\n", seq.Status.UnsafeL2.ParentHash.String()))
		details.WriteString(fmt.Sprintf("  L1 Origin: %s\n", seq.Status.UnsafeL2.L1Origin.Hash.String()))
		details.WriteString(fmt.Sprintf("  L1 Origin Number: %d\n", seq.Status.UnsafeL2.L1Origin.Number))
		details.WriteString(fmt.Sprintf("  Timestamp: %s\n", time.Unix(int64(seq.Status.UnsafeL2.Time), 0).Format(time.RFC3339)))
	}

	// Timing information
	if !seq.Status.LastUpdateTime.IsZero() {
		details.WriteString("\n[aqua]Timing:[-]\n")
		details.WriteString(fmt.Sprintf("  Last Update: %s\n", seq.Status.LastUpdateTime.Format(time.RFC3339)))
		details.WriteString(fmt.Sprintf("  Time Since Update: %s\n", time.Since(seq.Status.LastUpdateTime).Round(time.Second)))
	}

	d.TextView.SetText(details.String())
}

// formatBooleanStatus formats a boolean with color
func (d *DetailsPanel) formatBooleanStatus(status bool) string {
	if status {
		return "[green]✓ Yes[-]"
	}
	return "[red]✗ No[-]"
}

// Implement model.AppListener interface
func (d *DetailsPanel) OnDataChanged(sequencers []*sequencer.Sequencer) {
	// Update current sequencer if it still exists
	if d.current != nil {
		for _, seq := range sequencers {
			if seq.Config.ID == d.current.Config.ID {
				d.updateContent(seq)
				return
			}
		}
		// Current sequencer no longer exists
		d.updateContent(nil)
	}
}

func (d *DetailsPanel) OnSelectionChanged(seq *sequencer.Sequencer) {
	d.updateContent(seq)
}

func (d *DetailsPanel) OnError(err error) {
	// Details panel doesn't handle errors directly
}

func (d *DetailsPanel) OnRefreshCompleted(time.Time) {
	// No action needed
}
