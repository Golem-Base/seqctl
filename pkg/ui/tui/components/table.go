package components

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/golem-base/seqctl/pkg/sequencer"
	"github.com/golem-base/seqctl/pkg/ui/tui/styles"
	"github.com/rivo/tview"
)

// SequencerTable is a component that displays sequencers in a table
type SequencerTable struct {
	*tview.Table

	theme              *styles.Theme
	icons              *styles.Icons
	onSelectionChanged func(int)

	// Current data
	sequencers    []*sequencer.Sequencer
	selectedIndex int
}

// NewSequencerTable creates a new sequencer table component
func NewSequencerTable(theme *styles.Theme) *SequencerTable {
	table := &SequencerTable{
		Table:         tview.NewTable(),
		theme:         theme,
		icons:         styles.DefaultIcons(),
		selectedIndex: -1,
	}

	// Configure base table
	table.Table.
		SetFixed(1, 0).
		SetSelectable(true, false).
		SetSeparator(' ').
		SetSelectedStyle(tcell.StyleDefault.
			Background(theme.SelectedBg).
			Foreground(theme.SelectedFg).
			Attributes(tcell.AttrNone))

	// Enable dynamic colors for the table cells
	table.Table.SetBackgroundColor(theme.BackgroundColor)

	// Set border properties
	table.Table.SetBorder(true).
		SetBorderAttributes(tcell.AttrBold).
		SetBorderPadding(0, 0, 1, 1).
		SetTitle(" Sequencers ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.BorderColor)

	// Setup headers
	table.setupHeaders()

	// Handle selection changes
	table.Table.SetSelectionChangedFunc(func(row, column int) {
		if row > 0 && row-1 < len(table.sequencers) {
			table.selectedIndex = row - 1
			if table.onSelectionChanged != nil {
				table.onSelectionChanged(row - 1)
			}
		}
	})

	return table
}

// SetOnSelectionChanged sets the callback for selection changes
func (t *SequencerTable) SetOnSelectionChanged(fn func(int)) {
	t.onSelectionChanged = fn
}

// setupHeaders creates the table headers
func (t *SequencerTable) setupHeaders() {
	headers := []struct {
		text      string
		expansion int
		align     int
	}{
		{"", 0, tview.AlignCenter},           // Leader icon
		{"ID", 3, tview.AlignLeft},           // ID
		{"Active", 1, tview.AlignCenter},     // Active
		{"Healthy", 1, tview.AlignCenter},    // Healthy
		{"Sequencing", 1, tview.AlignCenter}, // Sequencing
		{"Voting", 1, tview.AlignCenter},     // Voting
	}

	for col, header := range headers {
		cell := tview.NewTableCell(header.text).
			SetTextColor(t.theme.HeaderFg).
			SetAlign(header.align).
			SetExpansion(header.expansion).
			SetSelectable(false).
			SetAttributes(tcell.AttrBold)

		if t.theme.HeaderBg != tcell.ColorDefault {
			cell.SetBackgroundColor(t.theme.HeaderBg)
		}

		t.Table.SetCell(0, col, cell)
	}
}

// updateTable populates the table with sequencer data
func (t *SequencerTable) updateTable() {
	// Clear existing data rows (keep header)
	for t.Table.GetRowCount() > 1 {
		t.Table.RemoveRow(1)
	}

	// Update each row
	for i, seq := range t.sequencers {
		row := i + 1 // Account for header row

		// Create cells for each column
		cells := []struct {
			text      string
			expansion int
			align     int
			color     tcell.Color
		}{
			{
				text:      t.formatLeaderIcon(seq.Status.ConductorLeader),
				expansion: 0,
				align:     tview.AlignCenter,
				color: func() tcell.Color {
					if seq.Status.ConductorLeader {
						return t.theme.LeaderColor
					}
					return t.theme.TableFg
				}(),
			},
			{
				text:      seq.Config.ID,
				expansion: 3,
				align:     tview.AlignLeft,
				color:     t.theme.TableFg,
			},
			{
				text:      t.formatBoolean(seq.Status.ConductorActive),
				expansion: 1,
				align:     tview.AlignCenter,
				color:     t.theme.TableFg,
			},
			{
				text:      t.formatBoolean(seq.Status.SequencerHealthy),
				expansion: 1,
				align:     tview.AlignCenter,
				color:     t.theme.TableFg,
			},
			{
				text:      t.formatBoolean(seq.Status.SequencerActive),
				expansion: 1,
				align:     tview.AlignCenter,
				color:     t.theme.TableFg,
			},
			{
				text:      t.formatBoolean(seq.Config.Voting),
				expansion: 1,
				align:     tview.AlignCenter,
				color:     t.theme.TableFg,
			},
		}

		for col, cellData := range cells {
			cell := tview.NewTableCell(cellData.text).
				SetAlign(cellData.align).
				SetExpansion(cellData.expansion)

			// Only set color if the text doesn't already contain color codes
			if !strings.Contains(cellData.text, "[") {
				cell.SetTextColor(cellData.color)
			}

			t.Table.SetCell(row, col, cell)
		}
	}

	// Maintain or set selection
	if t.selectedIndex >= 0 && t.selectedIndex < len(t.sequencers) {
		t.Table.Select(t.selectedIndex+1, 0)
	} else if len(t.sequencers) > 0 {
		// Select first row if nothing is selected or selection is out of bounds
		t.selectedIndex = 0
		t.Table.Select(1, 0)
		if t.onSelectionChanged != nil {
			t.onSelectionChanged(0)
		}
	}
}

// Focus sets focus to the table
func (t *SequencerTable) Focus(delegate func(p tview.Primitive)) {
	row, _ := t.Table.GetSelection()
	if row == 0 && t.Table.GetRowCount() > 1 {
		t.Table.Select(1, 0)
	}
	t.Table.Focus(delegate)
}

// NavigateUp moves selection up
func (t *SequencerTable) NavigateUp() {
	row, col := t.Table.GetSelection()
	if row > 1 {
		t.Table.Select(row-1, col)
	}
}

// NavigateDown moves selection down
func (t *SequencerTable) NavigateDown() {
	row, col := t.Table.GetSelection()
	if row < t.Table.GetRowCount()-1 {
		t.Table.Select(row+1, col)
	}
}

// SetData updates the table with new sequencer data (called by MainView)
func (t *SequencerTable) SetData(sequencers []*sequencer.Sequencer) {
	t.sequencers = sequencers
	t.updateTable()
}

// SetSelectedIndex sets the selected sequencer index
func (t *SequencerTable) SetSelectedIndex(index int) {
	if index >= 0 && index < len(t.sequencers) {
		t.selectedIndex = index
		t.Table.Select(index+1, 0) // +1 for header row
	}
}

// GetSelectedIndex returns the current selected index
func (t *SequencerTable) GetSelectedIndex() int {
	return t.selectedIndex
}

// formatBoolean formats a boolean value with colored icon
func (t *SequencerTable) formatBoolean(status bool) string {
	if status {
		return fmt.Sprintf("[%s]%s[-]", t.theme.SuccessColor.String(), t.icons.Active)
	}
	return fmt.Sprintf("[%s]%s[-]", t.theme.ErrorColor.String(), t.icons.Inactive)
}

// formatLeaderIcon formats leader status for icon column (empty if not leader)
func (t *SequencerTable) formatLeaderIcon(isLeader bool) string {
	if isLeader {
		return t.icons.Leader
	}
	return ""
}
