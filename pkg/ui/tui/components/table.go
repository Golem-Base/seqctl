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

	// Multi-selection support
	marks map[string]struct{}
}

// NewSequencerTable creates a new sequencer table component
func NewSequencerTable(theme *styles.Theme) *SequencerTable {
	table := &SequencerTable{
		Table:         tview.NewTable(),
		theme:         theme,
		icons:         styles.DefaultIcons(),
		selectedIndex: -1,
		marks:         make(map[string]struct{}),
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

			// Set sequencer ID as reference in first column for selection tracking
			if col == 0 {
				cell.SetReference(seq.Config.ID)
			}

			// Apply marked styling if this sequencer is marked
			isMarked := t.IsMarked(seq.Config.ID)
			if isMarked {
				// Use theme's mark color for marked items
				cell.SetTextColor(t.theme.MarkColor)
			} else if !strings.Contains(cellData.text, "[") {
				// Only set color if the text doesn't already contain color codes and not marked
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

// GetRowID returns the sequencer ID at given row index
func (t *SequencerTable) GetRowID(index int) (string, bool) {
	if index <= 0 || index > len(t.sequencers) {
		return "", false
	}
	cell := t.Table.GetCell(index, 0)
	if cell == nil {
		return "", false
	}
	id, ok := cell.GetReference().(string)
	return id, ok
}

// GetSelectedItem returns the currently selected sequencer ID
func (t *SequencerTable) GetSelectedItem() string {
	row, _ := t.Table.GetSelection()
	if row <= 0 || row-1 >= len(t.sequencers) {
		return ""
	}
	return t.sequencers[row-1].Config.ID
}

// GetSelectedItems returns all marked sequencer IDs, or current selection if none marked
func (t *SequencerTable) GetSelectedItems() []string {
	if len(t.marks) == 0 {
		if item := t.GetSelectedItem(); item != "" {
			return []string{item}
		}
		return nil
	}

	items := make([]string, 0, len(t.marks))
	for item := range t.marks {
		items = append(items, item)
	}
	return items
}

// IsMarked returns true if the sequencer is marked for multi-selection
func (t *SequencerTable) IsMarked(id string) bool {
	_, ok := t.marks[id]
	return ok
}

// ToggleMark toggles the mark status of the currently selected sequencer
func (t *SequencerTable) ToggleMark() {
	sel := t.GetSelectedItem()
	if sel == "" {
		return
	}

	if _, ok := t.marks[sel]; ok {
		delete(t.marks, sel)
	} else {
		t.marks[sel] = struct{}{}
	}

	// Refresh the table to update visual indicators
	t.updateTable()
}

// ClearMarks removes all marks
func (t *SequencerTable) ClearMarks() {
	for k := range t.marks {
		delete(t.marks, k)
	}
	// Refresh the table to update visual indicators
	t.updateTable()
}

// SpanMark marks a range of sequencers from the last marked item to current selection
func (t *SequencerTable) SpanMark() {
	selIndex, _ := t.Table.GetSelection()
	if selIndex <= 0 {
		return
	}

	prev := -1
	// Look back to find previous mark
	for i := selIndex - 1; i > 0; i-- {
		id, ok := t.GetRowID(i)
		if !ok {
			break
		}
		if _, ok := t.marks[id]; ok {
			prev = i
			break
		}
	}

	if prev != -1 {
		t.markRange(prev, selIndex)
		return
	}

	// Look forward to see if we have a mark
	for i := selIndex; i < t.Table.GetRowCount(); i++ {
		id, ok := t.GetRowID(i)
		if !ok {
			break
		}
		if _, ok := t.marks[id]; ok {
			prev = i
			break
		}
	}
	t.markRange(prev, selIndex)
}

// markRange marks sequencers in the given range
func (t *SequencerTable) markRange(prev, curr int) {
	if prev < 0 {
		return
	}
	if prev > curr {
		prev, curr = curr, prev
	}

	for i := prev + 1; i <= curr; i++ {
		id, ok := t.GetRowID(i)
		if !ok {
			break
		}
		t.marks[id] = struct{}{}
	}

	// Refresh the table to update visual indicators
	t.updateTable()
}
