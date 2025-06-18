package tui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/golem-base/seqctl/pkg/config"
	"github.com/golem-base/seqctl/pkg/network"
	"github.com/golem-base/seqctl/pkg/ui/tui/actions"
	"github.com/golem-base/seqctl/pkg/ui/tui/managers"
	"github.com/golem-base/seqctl/pkg/ui/tui/model"
	"github.com/golem-base/seqctl/pkg/ui/tui/styles"
	"github.com/golem-base/seqctl/pkg/ui/tui/views"
	"github.com/rivo/tview"
)

// TUI represents the main TUI application
type TUI struct {
	// Core components
	app *tview.Application

	// Models
	appModel   *model.AppModel
	flashModel *model.FlashModel

	// Views
	mainView *views.MainView
	helpView *views.HelpView

	// Managers
	navigation       *managers.NavigationManager
	refresh          *managers.RefreshManager
	actionDispatcher *managers.ActionDispatcher

	// Configuration
	uiConfig *config.UIConfig
	theme    *styles.Theme
	icons    *styles.Icons
}

// NewTUI creates a new TUI with clean architecture
func NewTUI(network *network.Network, uiConfig *config.UIConfig) *TUI {
	// Use default UI config if none provided
	if uiConfig == nil {
		cfg := config.New()
		uiConfig = &cfg.UI
	}

	// Get theme and icons from config
	theme, err := uiConfig.GetTheme()
	if err != nil {
		// Fallback to default theme if config is invalid
		theme = styles.Default()
	}

	icons, err := uiConfig.GetIcons()
	if err != nil {
		// Fallback to default icons if config is invalid
		icons = styles.DefaultIcons()
	}

	tui := &TUI{
		app:      tview.NewApplication(),
		uiConfig: uiConfig,
		theme:    theme,
		icons:    icons,
	}

	// Initialize models
	tui.appModel = model.NewAppModel(network)
	tui.flashModel = model.NewFlashModel()

	// Initialize refresh manager first (needed by MainView)
	tui.refresh = managers.NewRefreshManager(tui.appModel, tui.flashModel, tui.app)

	// Initialize views
	tui.mainView = views.NewMainView(tui.appModel, tui.flashModel, tui.refresh, tui.theme, tui.icons)
	tui.helpView = views.NewHelpView(tui.theme)

	// Initialize navigation manager
	tui.navigation = managers.NewNavigationManager(tui.app, tui.mainView, tui.helpView)

	// Initialize dialog manager and action dispatcher
	dialogManager := managers.NewDialogManager(tui.navigation.GetPages(), tui.flashModel, tui.theme)
	tui.actionDispatcher = managers.NewActionDispatcher(tui.appModel, tui.flashModel, tui.app, dialogManager, tui.refresh)

	// Setup key handling
	tui.setupKeyHandling()

	return tui
}

// setupKeyHandling configures global keyboard shortcuts
func (t *TUI) setupKeyHandling() {
	t.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			t.Stop()
			return nil
		}

		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'q', 'Q':
				t.Stop()
				return nil
			case '?':
				t.navigation.ToggleHelp()
				return nil
			case '1':
				t.navigation.SetFocusToPanel(managers.FocusTable)
				return nil
			case '2':
				t.navigation.SetFocusToPanel(managers.FocusDetails)
				return nil
			default:
				// Handle action keys if on main view
				if t.navigation.IsMainView() {
					if action := actions.GetActionByKey(event.Rune()); action != nil {
						seq := t.appModel.GetSelectedSequencer()
						t.actionDispatcher.Execute(action, seq)
						return nil
					}
				}
			}
		}

		// Delegate to current view for navigation
		if t.navigation.IsMainView() {
			return t.mainView.HandleKey(event)
		}

		return event
	})
}

// Run starts the TUI application
func (t *TUI) Run() error {
	// Start initial data loading
	t.refresh.InitialLoad()

	// Start auto-refresh
	t.refresh.Start()

	// Set initial focus to main page
	t.navigation.ShowMainView()

	// Run the application
	return t.app.Run()
}

// Stop gracefully stops the application
func (t *TUI) Stop() {
	t.refresh.Stop()
	t.app.Stop()
}

// SetAutoRefresh enables or disables auto-refresh
func (t *TUI) SetAutoRefresh(enabled bool) {
	t.refresh.SetEnabled(enabled)
}

// SetRefreshInterval sets the auto-refresh interval
func (t *TUI) SetRefreshInterval(interval time.Duration) {
	t.refresh.SetInterval(interval)
}

// SetReadOnlyMode sets the read-only mode
func (t *TUI) SetReadOnlyMode(readOnly bool) {
	t.actionDispatcher.SetReadOnlyMode(readOnly)
}

// SetConfirmDanger sets whether dangerous actions require confirmation
func (t *TUI) SetConfirmDanger(confirm bool) {
	t.actionDispatcher.SetConfirmDanger(confirm)
}

// GetUIConfig returns the current UI configuration
func (t *TUI) GetUIConfig() *config.UIConfig {
	return t.uiConfig
}
