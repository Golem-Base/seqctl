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

// RefreshManager interface to avoid circular import
type RefreshManager interface {
	IsEnabled() bool
	SetEnabled(bool)
}

// MainView is the primary sequencer management view
type MainView struct {
	// Layout components
	container    *tview.Flex
	headerView   *tview.TextView
	footerView   *tview.TextView
	flashMessage *components.FlashMessage

	// Main components
	table          *components.SequencerTable
	loadingState   *components.LoadingState
	errorState     *components.ErrorState
	detailsPanel   *components.DetailsPanel
	operationsView *tview.TextView
	infoPanel      *tview.Flex

	// Content area (switches between table/loading/error)
	contentArea *tview.Flex

	// Models
	appModel       *model.AppModel
	flashModel     *model.FlashModel
	refreshManager RefreshManager

	// State
	showDetails     bool
	focusedPanel    int
	focusablePanels []tview.Primitive
	currentState    ViewState
	theme           *styles.Theme
	icons           *styles.Icons
}

// ViewState represents the current state of the main view
type ViewState int

const (
	StateLoading ViewState = iota
	StateError
	StateData
	StateEmpty
)

// NewMainView creates the main sequencer view
func NewMainView(appModel *model.AppModel, flashModel *model.FlashModel, refreshManager RefreshManager) *MainView {
	view := &MainView{
		appModel:       appModel,
		flashModel:     flashModel,
		refreshManager: refreshManager,
		showDetails:    true,
		focusedPanel:   FocusTable,
		currentState:   StateLoading,
		theme:          styles.Default(),
		icons:          styles.DefaultIcons(),
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

	// Table and state components
	v.table = components.NewSequencerTable(v.theme)
	v.loadingState = components.NewLoadingState(v.theme)
	v.errorState = components.NewErrorState(v.theme)

	// Setup table selection callback
	v.table.SetOnSelectionChanged(func(index int) {
		if index >= 0 {
			v.appModel.SetSelectedIndex(index)
		}
	})

	// Details panel
	v.detailsPanel = components.NewDetailsPanel(v.theme)

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

	// Content area that switches between states
	v.contentArea = tview.NewFlex().
		SetDirection(tview.FlexColumn)

	// Start with loading state
	v.showLoadingState()

	// Main content area (contentArea will manage the layout based on state)
	mainContent := v.contentArea

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
	connectionStatus := fmt.Sprintf("[%s]Connected[-]", v.theme.SuccessColor.String())
	if lastUpdate.IsZero() {
		connectionStatus = fmt.Sprintf("[%s]Connecting...[-]", v.theme.WarningColor.String())
	} else if time.Since(lastUpdate) > 30*time.Second {
		connectionStatus = fmt.Sprintf("[%s]Disconnected[-]", v.theme.ErrorColor.String())
	}

	// Build header
	var header string
	if lastUpdate.IsZero() {
		header = fmt.Sprintf("%s Network: [%s]%s[-] | Status: %s | [%s]Loading data...[-]",
			v.icons.Network, v.theme.PrimaryColor.String(), network.Name(), connectionStatus, v.theme.WarningColor.String())
	} else {
		header = fmt.Sprintf("%s Network: [%s]%s[-] | Status: %s | Last Update: %s",
			v.icons.Network, v.theme.PrimaryColor.String(), network.Name(), connectionStatus,
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

		color := v.theme.PrimaryColor.String()
		if action.Dangerous {
			color = v.theme.DangerColor.String()
		}
		if !enabled {
			color = v.theme.SecondaryColor.String()
		}

		text += fmt.Sprintf("[%s]%c[-] %s\n", color, action.Key, action.Description)
	}

	v.operationsView.SetText(text)
}

// getFooterText returns the footer help text
func (v *MainView) getFooterText() string {
	return fmt.Sprintf("[%s] 1: Table | 2: Details | Move: ↑↓/j/k | Refresh: r | Auto-refresh: a | Details: i | Help: ? | Quit: q[-]", v.theme.SecondaryColor.String())
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
	// Show loading state before starting refresh
	if v.currentState != StateLoading {
		v.showLoadingState()
	}

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
	enabled := !v.refreshManager.IsEnabled()
	v.refreshManager.SetEnabled(enabled)

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

	// If we're in data state, refresh the layout
	if v.currentState == StateData {
		v.showDataState()
	}
}

// Focus sets initial focus to the table
func (v *MainView) Focus() {
}

// Implement model.AppListener interface - MainView coordinates all UI updates
func (v *MainView) OnDataChanged(sequencers []*sequencer.Sequencer) {
	if len(sequencers) == 0 {
		v.showEmptyState()
	} else {
		// Update table data and show data state
		v.table.SetData(sequencers)
		v.showDataState()

		// Update details panel with current sequencers
		v.detailsPanel.UpdateData(sequencers)
	}

	// Update MainView-specific UI elements
	v.updateHeader()
	v.updateOperationsView()
}

func (v *MainView) OnSelectionChanged(seq *sequencer.Sequencer) {
	// Update details panel with selected sequencer
	v.detailsPanel.SetData(seq)

	// Update operations view
	v.updateOperationsView()
}

func (v *MainView) OnError(err error) {
	if err != nil {
		// Show error state
		v.showErrorState(err)

		// Also show error in flash message
		v.flashModel.Error(err.Error())
	}
}

func (v *MainView) OnRefreshCompleted(t time.Time) {
	v.updateHeader()
}

// State transition methods
func (v *MainView) showLoadingState() {
	v.currentState = StateLoading
	v.contentArea.Clear()
	v.contentArea.SetDirection(tview.FlexRow)
	v.contentArea.AddItem(v.loadingState, 0, 1, true)
	v.loadingState.ShowLoading("")
}

func (v *MainView) showErrorState(err error) {
	v.currentState = StateError
	v.contentArea.Clear()
	v.contentArea.SetDirection(tview.FlexRow)
	v.contentArea.AddItem(v.errorState, 0, 1, true)
	v.errorState.ShowError(err)
}

func (v *MainView) showDataState() {
	v.currentState = StateData
	v.contentArea.Clear()
	v.contentArea.SetDirection(tview.FlexColumn)

	if v.showDetails {
		v.contentArea.
			AddItem(v.table, 0, 7, true).
			AddItem(v.infoPanel, 0, 3, false)
	} else {
		v.contentArea.AddItem(v.table, 0, 1, true)
	}
}

func (v *MainView) showEmptyState() {
	v.currentState = StateEmpty
	v.contentArea.Clear()
	v.contentArea.SetDirection(tview.FlexRow)
	v.contentArea.AddItem(v.loadingState, 0, 1, true)
	v.loadingState.ShowEmpty("No sequencers found - check Kubernetes connection and labels")
}
