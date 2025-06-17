package views

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/golem-base/seqctl/pkg/sequencer"
	"github.com/golem-base/seqctl/pkg/ui/tui/actions"
	"github.com/golem-base/seqctl/pkg/ui/tui/components"
	"github.com/golem-base/seqctl/pkg/ui/tui/model"
	"github.com/golem-base/seqctl/pkg/ui/tui/styles"
	"github.com/rivo/tview"
)

// Focus panel constants
const (
	FocusTable = iota
	FocusDetails
)

// MainView is the primary sequencer management view
type MainView struct {
	// Layout components
	container    *tview.Flex
	headerView   *tview.TextView
	footerView   *tview.TextView
	flashMessage *components.FlashMessage

	// Main components
	table          *components.SequencerTable
	detailsPanel   *components.DetailsPanel
	operationsView *tview.TextView
	infoPanel      *tview.Flex

	// Models
	appModel   *model.AppModel
	flashModel *model.FlashModel

	// State
	showDetails     bool
	focusedPanel    int
	focusablePanels []tview.Primitive
	theme           *styles.Theme
	icons           *styles.Icons
}

// NewMainView creates the main sequencer view
func NewMainView(appModel *model.AppModel, flashModel *model.FlashModel) *MainView {
	view := &MainView{
		appModel:     appModel,
		flashModel:   flashModel,
		showDetails:  true,
		focusedPanel: FocusTable,
		theme:        styles.Default(),
		icons:        styles.DefaultIcons(),
	}

	// Create components
	view.createComponents()

	// Setup layout
	view.setupLayout()

	// Initialize focusable panels array
	view.focusablePanels = []tview.Primitive{
		view.table,
		view.detailsPanel,
	}

	// Register as listener
	appModel.AddListener(view)

	return view
}


// createComponents creates all UI components
func (v *MainView) createComponents() {
	// Header
	v.headerView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	v.updateHeader()

	// Footer
	v.footerView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetText(v.getFooterText())

	// Flash messages
	v.flashMessage = components.NewFlashMessage(v.flashModel, v.theme)

	// Table
	v.table = components.NewSequencerTable(v.appModel, v.theme)

	// Details panel
	v.detailsPanel = components.NewDetailsPanel(v.appModel, v.theme)

	// Operations view
	v.operationsView = tview.NewTextView()
	v.operationsView.SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetBorderPadding(1, 1, 1, 1)
	v.updateOperationsView()
}

// setupLayout creates the layout structure
func (v *MainView) setupLayout() {
	// Create bordered sections
	detailsSection := v.createBorderedSection("Sequencer Info", v.detailsPanel)
	operationsSection := v.createBorderedSection("Operations", v.operationsView)

	// Info panel (right side)
	v.infoPanel = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(detailsSection, 0, 2, false).
		AddItem(operationsSection, 0, 1, false)

	// Main content area
	mainContent := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(v.table, 0, 7, true).
		AddItem(v.infoPanel, 0, 3, false)

	// Complete layout
	v.container = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(v.headerView, 1, 0, false).
		AddItem(v.flashMessage, 1, 0, false).
		AddItem(mainContent, 0, 1, true).
		AddItem(v.footerView, 1, 0, false)
}

// createBorderedSection creates a flex with border
func (v *MainView) createBorderedSection(title string, content tview.Primitive) *tview.Flex {
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(content, 0, 1, false)

	flex.SetBorder(true).
		SetTitle(title).
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(v.theme.BorderColor)

	return flex
}

// updateHeader updates the header with network statistics
func (v *MainView) updateHeader() {
	network := v.appModel.GetNetwork()
	lastUpdate := v.appModel.GetLastUpdate()

	// Connection status based on whether we have recent data
	connectionStatus := "[green]Connected[-]"
	if lastUpdate.IsZero() {
		connectionStatus = "[yellow]Connecting...[-]"
	} else if time.Since(lastUpdate) > 30*time.Second {
		connectionStatus = "[red]Disconnected[-]"
	}

	// Build header
	var header string
	if lastUpdate.IsZero() {
		header = fmt.Sprintf("%s Network: [aqua]%s[-] | Status: %s | [yellow]Loading data...[-]",
			v.icons.Network, network.Name(), connectionStatus)
	} else {
		header = fmt.Sprintf("%s Network: [aqua]%s[-] | Status: %s | Last Update: %s",
			v.icons.Network, network.Name(), connectionStatus,
			lastUpdate.Format("15:04:05"),
		)
	}

	v.headerView.SetText(header)
}

// updateOperationsView updates the operations panel
func (v *MainView) updateOperationsView() {
	selected := v.appModel.GetSelectedSequencer()

	var text string
	for _, action := range actions.GetVisibleActions() {
		enabled := action.Enabled == nil || (selected != nil && action.Enabled(selected))

		color := "aqua"
		if action.Dangerous {
			color = "orange"
		}
		if !enabled {
			color = "dim"
		}

		text += fmt.Sprintf("[%s]%c[-] %s\n", color, action.Key, action.Description)
	}

	v.operationsView.SetText(text)
}

// getFooterText returns the footer help text
func (v *MainView) getFooterText() string {
	return "[dim] 1: Table | 2: Details | Move: ↑↓/j/k | Refresh: r | Auto-refresh: a | Details: i | Help: ? | Quit: q[-]"
}

// GetContainer returns the root container
func (v *MainView) GetContainer() *tview.Flex {
	return v.container
}

// GetActionRegistry returns all actions (updated for simplified actions)
func (v *MainView) GetActionRegistry() []*actions.Action {
	return actions.AllActions
}

// GetTable returns the table component for focus management
func (v *MainView) GetTable() *components.SequencerTable {
	return v.table
}

// GetDetailsPanel returns the details panel component for focus management
func (v *MainView) GetDetailsPanel() *components.DetailsPanel {
	return v.detailsPanel
}

// SetFocusToPanel sets focus to a specific panel by index
func (v *MainView) SetFocusToPanel(app *tview.Application, panelIndex int) {
	if panelIndex < 0 || panelIndex >= len(v.focusablePanels) {
		return
	}

	// Special validation for details panel
	if panelIndex == FocusDetails && !v.showDetails {
		v.flashModel.Warning("Details panel is hidden")
		return
	}

	v.focusedPanel = panelIndex
	app.SetFocus(v.focusablePanels[panelIndex])
}

// HandleKey processes keyboard input (navigation and non-action keys only)
func (v *MainView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyRune {
		switch event.Rune() {
		case 'r', 'R':
			v.flashModel.Info("Refreshing data...")
			v.refresh()
			return nil
		case 'a', 'A':
			v.toggleAutoRefresh()
			return nil
		case 'i', 'I':
			v.toggleDetails()
			return nil
		case 'j', 'J':
			v.table.NavigateDown()
			return nil
		case 'k', 'K':
			v.table.NavigateUp()
			return nil
		}
	}

	// Handle arrow keys
	switch event.Key() {
	case tcell.KeyUp:
		v.table.NavigateUp()
		return nil
	case tcell.KeyDown:
		v.table.NavigateDown()
		return nil
	}

	return event
}

// refresh triggers a data refresh
func (v *MainView) refresh() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := v.appModel.Refresh(ctx); err != nil {
			v.flashModel.Error(fmt.Sprintf("Refresh failed: %s", err.Error()))
		}
	}()
}

// toggleAutoRefresh toggles auto-refresh
func (v *MainView) toggleAutoRefresh() {
	enabled := !v.appModel.IsAutoRefresh()
	v.appModel.SetAutoRefresh(enabled)

	if enabled {
		v.flashModel.Info("Auto-refresh enabled")
	} else {
		v.flashModel.Info("Auto-refresh disabled")
	}
}

// toggleDetails toggles the details panel
func (v *MainView) toggleDetails() {
	v.showDetails = !v.showDetails

	// If hiding details and currently focused on details, switch focus to table
	if !v.showDetails && v.focusedPanel == FocusDetails {
		v.focusedPanel = FocusTable
	}

	// Rebuild main content
	mainContent := v.container.GetItem(2).(*tview.Flex)
	mainContent.Clear()

	if v.showDetails {
		mainContent.
			AddItem(v.table, 0, 7, true).
			AddItem(v.infoPanel, 0, 3, false)
	} else {
		mainContent.AddItem(v.table, 0, 1, true)
	}
}

// Focus sets initial focus to the table
func (v *MainView) Focus() {
}

// Implement model.AppListener interface
func (v *MainView) OnDataChanged(sequencers []*sequencer.Sequencer) {
	v.updateHeader()
	v.updateOperationsView()
}

func (v *MainView) OnSelectionChanged(seq *sequencer.Sequencer) {
	v.updateOperationsView()
}

func (v *MainView) OnError(err error) {
	if err != nil {
		v.flashModel.Error(err.Error())
	}
}

func (v *MainView) OnRefreshCompleted(t time.Time) {
	v.updateHeader()
}
