package components

import (
	"fmt"

	"github.com/golem-base/seqctl/pkg/ui/tui/styles"
	"github.com/rivo/tview"
)

// ErrorState displays error messages
type ErrorState struct {
	*tview.TextView
	theme *styles.Theme
}

// NewErrorState creates a new error state component
func NewErrorState(theme *styles.Theme) *ErrorState {
	errorState := &ErrorState{
		TextView: tview.NewTextView(),
		theme:    theme,
	}

	// Configure TextView
	errorState.TextView.
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetBorder(true).
		SetTitle(" Error ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.ErrorColor).
		SetBackgroundColor(theme.BackgroundColor)

	return errorState
}

// ShowError displays an error message
func (e *ErrorState) ShowError(err error) {
	if err == nil {
		e.TextView.SetText("[red]Unknown error occurred[-]")
		return
	}
	
	message := fmt.Sprintf("[red]Error: %s[-]\n\n[dim]Press 'r' to retry[-]", err.Error())
	e.TextView.SetText(message)
}

// ShowConnectionError displays a connection-specific error
func (e *ErrorState) ShowConnectionError(message string) {
	if message == "" {
		message = "Failed to connect to Kubernetes cluster"
	}
	
	errorText := fmt.Sprintf("[red]Connection Error[-]\n\n%s\n\n[dim]Check your kubeconfig and cluster connection[-]", message)
	e.TextView.SetText(errorText)
}