package managers

import (
	"github.com/golem-base/seqctl/pkg/ui/tui/views"
	"github.com/rivo/tview"
)

// FocusPanel represents focus panels in the main view
type FocusPanel int

const (
	FocusTable FocusPanel = iota
	FocusDetails
)

// NavigationManager handles page navigation and focus management
type NavigationManager struct {
	app      *tview.Application
	pages    *tview.Pages
	mainView *views.MainView
	helpView *views.HelpView
}

// NewNavigationManager creates a new navigation manager
func NewNavigationManager(app *tview.Application, mainView *views.MainView, helpView *views.HelpView) *NavigationManager {
	nav := &NavigationManager{
		app:      app,
		mainView: mainView,
		helpView: helpView,
	}

	nav.setupPages()
	return nav
}

// GetPages returns the pages container
func (n *NavigationManager) GetPages() *tview.Pages {
	return n.pages
}

// ShowMainView shows the main sequencer view
func (n *NavigationManager) ShowMainView() {
	n.pages.SwitchToPage("main")
	n.app.SetFocus(n.mainView.GetTable())
}

// ShowHelpView shows the help view
func (n *NavigationManager) ShowHelpView() {
	n.pages.SwitchToPage("help")
}

// ToggleHelp toggles between main and help view
func (n *NavigationManager) ToggleHelp() {
	frontPage, _ := n.pages.GetFrontPage()
	if frontPage == "main" {
		n.ShowHelpView()
	} else {
		n.ShowMainView()
	}
}

// SetFocusToPanel sets focus to a specific panel in the main view
func (n *NavigationManager) SetFocusToPanel(panel FocusPanel) {
	frontPage, _ := n.pages.GetFrontPage()
	if frontPage == "main" {
		n.mainView.SetFocusToPanel(n.app, int(panel))
	}
}

// GetCurrentPage returns the current front page name
func (n *NavigationManager) GetCurrentPage() string {
	frontPage, _ := n.pages.GetFrontPage()
	return frontPage
}

// IsMainView returns true if main view is currently shown
func (n *NavigationManager) IsMainView() bool {
	return n.GetCurrentPage() == "main"
}

// setupPages configures the page navigation
func (n *NavigationManager) setupPages() {
	n.pages = tview.NewPages()

	n.pages.AddPage("main", n.mainView.GetContainer(), true, true)
	n.pages.AddPage("help", n.helpView, true, false)

	n.app.SetRoot(n.pages, true).SetFocus(n.mainView.GetTable())
}
