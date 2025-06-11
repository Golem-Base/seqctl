package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/golem-base/seqctl/pkg/sequencer"
	"github.com/golem-base/seqctl/pkg/ui/tui/model"
	"github.com/golem-base/seqctl/pkg/ui/tui/styles"
	"github.com/rivo/tview"
)

// SequencerTable is a component that displays sequencers in a table
type SequencerTable struct {
	*tview.Table

	model              *model.AppModel
	theme              *styles.Theme
	icons              *styles.Icons
	onSelectionChanged func(int)

	// Cache
	lastSequencerCount int
}

// NewSequencerTable creates a new sequencer table component
func NewSequencerTable(appModel *model.AppModel, theme *styles.Theme) *SequencerTable {
	table := &SequencerTable{
		Table: tview.NewTable(),
		model: appModel,
		theme: theme,
		icons: styles.DefaultIcons(),
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

	// Show initial loading state
	table.showLoadingState()

	// Register as model listener
	appModel.AddListener(table)

	// Handle selection changes
	table.Table.SetSelectionChangedFunc(func(row, column int) {
		if row > 0 {
			table.model.SetSelectedIndex(row - 1)
			if table.onSelectionChanged != nil {
				table.onSelectionChanged(row - 1)
			}
		}
	})

	// Input handling will be managed by the parent view

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
func (t *SequencerTable) updateTable(sequencers []*sequencer.Sequencer) {
	// Show error if no sequencers
	if len(sequencers) == 0 {
		t.showError("No sequencers found - check Kubernetes connection and labels")
		return
	}

	// Update each row
	for i, seq := range sequencers {
		row := i + 1 // Account for header row

		// Create cells for each column
		cells := []struct {
			text      string
			expansion int
			align     int
			color     tcell.Color
		}{
			{
				text:      styles.FormatLeaderIcon(seq.Status.ConductorLeader, t.icons),
				expansion: 0,
				align:     tview.AlignCenter,
				color:     t.getLeaderColor(seq.Status.ConductorLeader),
			},
			{
				text:      seq.Config.ID,
				expansion: 3,
				align:     tview.AlignLeft,
				color:     t.theme.TableFg,
			},
			{
				text:      styles.FormatBooleanColored(seq.Status.ConductorActive, t.icons),
				expansion: 1,
				align:     tview.AlignCenter,
				color:     t.theme.TableFg,
			},
			{
				text:      styles.FormatBooleanColored(seq.Status.SequencerHealthy, t.icons),
				expansion: 1,
				align:     tview.AlignCenter,
				color:     t.theme.TableFg,
			},
			{
				text:      styles.FormatBooleanColored(seq.Status.SequencerActive, t.icons),
				expansion: 1,
				align:     tview.AlignCenter,
				color:     t.theme.TableFg,
			},
			{
				text:      styles.FormatBooleanColored(seq.Config.Voting, t.icons),
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

	// Remove extra rows if the count decreased
	t.trimTableRows(len(sequencers) + 1)
	t.lastSequencerCount = len(sequencers)

	// Maintain selection, or select first row if none selected
	selectedIndex := t.model.GetSelectedIndex()
	if selectedIndex >= 0 && selectedIndex < len(sequencers) {
		t.Table.Select(selectedIndex+1, 0)
	} else if len(sequencers) > 0 && selectedIndex < 0 {
		// Select first row if nothing is selected
		t.Table.Select(1, 0)
		t.model.SetSelectedIndex(0)
	}
}

// showError displays an error message in the table
func (t *SequencerTable) showError(message string) {
	// Clear leader icon column
	t.Table.SetCell(1, 0, tview.NewTableCell(""))

	// Show error in ID column
	t.Table.SetCell(1, 1, tview.NewTableCell(message).
		SetAlign(tview.AlignCenter).
		SetTextColor(t.theme.ErrorColor))

	// Clear other columns
	for col := 2; col < 6; col++ {
		t.Table.SetCell(1, col, tview.NewTableCell(""))
	}

	// Remove extra rows
	t.trimTableRows(2)
}

// trimTableRows removes rows beyond the specified count
func (t *SequencerTable) trimTableRows(targetRows int) {
	for t.Table.GetRowCount() > targetRows {
		t.Table.RemoveRow(t.Table.GetRowCount() - 1)
	}
}

// getLeaderColor returns the appropriate color for leader status
func (t *SequencerTable) getLeaderColor(isLeader bool) tcell.Color {
	if isLeader {
		return t.theme.LeaderColor
	}
	return t.theme.TableFg
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

// Implement model.AppListener interface
func (t *SequencerTable) OnDataChanged(sequencers []*sequencer.Sequencer) {
	t.updateTable(sequencers)
}

func (t *SequencerTable) OnSelectionChanged(seq *sequencer.Sequencer) {
	// Selection is handled internally, no need to update
}

func (t *SequencerTable) OnError(err error) {
	if err != nil {
		t.showError(fmt.Sprintf("Error: %s", err.Error()))
	}
}

func (t *SequencerTable) OnRefreshCompleted(time.Time) {
	// No action needed for table
}

// showLoadingState displays a loading message
func (t *SequencerTable) showLoadingState() {
	// Clear leader icon column
	t.Table.SetCell(1, 0, tview.NewTableCell(""))

	// Show loading in ID column
	t.Table.SetCell(1, 1, tview.NewTableCell("[yellow]Loading sequencers...[-]").
		SetAlign(tview.AlignCenter).
		SetExpansion(3))

	// Clear other columns
	for col := 2; col < 6; col++ {
		t.Table.SetCell(1, col, tview.NewTableCell(""))
	}

	// Remove any extra rows
	t.trimTableRows(2)
}
