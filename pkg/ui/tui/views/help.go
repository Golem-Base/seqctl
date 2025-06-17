package views

import (
	"fmt"
	"strings"

	"github.com/golem-base/seqctl/pkg/ui/tui/actions"
	"github.com/golem-base/seqctl/pkg/ui/tui/styles"
	"github.com/rivo/tview"
)

// HelpView displays help information
type HelpView struct {
	*tview.TextView

	theme *styles.Theme
}

// NewHelpView creates a new help view
func NewHelpView(theme *styles.Theme) *HelpView {
	view := &HelpView{
		TextView: tview.NewTextView(),
		theme:    theme,
	}

	// Configure TextView
	view.TextView.
		SetDynamicColors(true).
		SetScrollable(true).
		SetBorderPadding(1, 1, 2, 2)

	// Set help content
	view.updateContent()

	return view
}

// updateContent updates the help text
func (v *HelpView) updateContent() {
	var help strings.Builder

	help.WriteString("[::b]GB Conductor Ops - Keyboard Shortcuts[::-]\n\n")

	// Navigation section
	help.WriteString(fmt.Sprintf("[%s]Navigation:[-]\n", v.theme.PrimaryColor.String()))
	help.WriteString("  ↑/↓       Move selection up/down\n")
	help.WriteString("  j/k       Move selection down/up (vim-style)\n")
	help.WriteString("  Enter     Show quick actions for selected sequencer\n")
	help.WriteString("  i         Toggle info panel visibility\n\n")

	// Operations section
	help.WriteString(fmt.Sprintf("[%s]Sequencer Operations:[-]\n", v.theme.PrimaryColor.String()))
	actions := actions.AllActions
	for _, action := range actions {
		color := v.theme.SecondaryColor.String()
		if action.Dangerous {
			color = v.theme.DangerColor.String()
		}

		help.WriteString(fmt.Sprintf("  [%s]%c[-]         %s\n", color, action.Key, action.Description))
	}
	help.WriteString("\n")

	// General section
	help.WriteString(fmt.Sprintf("[%s]General:[-]\n", v.theme.PrimaryColor.String()))
	help.WriteString("  r         Refresh data\n")
	help.WriteString("  a         Toggle auto-refresh\n")
	help.WriteString("  ?         Show this help\n")
	help.WriteString("  q         Quit application\n")
	help.WriteString("  Ctrl+C    Force quit\n\n")

	// Notes
	help.WriteString(fmt.Sprintf("[%s]Notes:[-]\n", v.theme.SecondaryColor.String()))
	help.WriteString(fmt.Sprintf("[%s]- Operations apply to the currently highlighted sequencer[-]\n", v.theme.SecondaryColor.String()))
	help.WriteString(fmt.Sprintf("[%s]- Orange operations are potentially dangerous[-]\n", v.theme.SecondaryColor.String()))
	help.WriteString(fmt.Sprintf("[%s]- Some operations may be disabled based on sequencer state[-]\n", v.theme.SecondaryColor.String()))

	v.TextView.SetText(help.String())
}
