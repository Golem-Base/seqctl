package components

import (
	"fmt"

	"github.com/golem-base/seqctl/pkg/ui/tui/styles"
	"github.com/rivo/tview"
)

// LoadingState displays a loading message
type LoadingState struct {
	*tview.TextView
	theme *styles.Theme
}

// NewLoadingState creates a new loading state component
func NewLoadingState(theme *styles.Theme) *LoadingState {
	loading := &LoadingState{
		TextView: tview.NewTextView(),
		theme:    theme,
	}

	// Configure TextView
	loading.TextView.
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetBorder(true).
		SetTitle(" Status ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.BorderColor).
		SetBackgroundColor(theme.BackgroundColor)

	return loading
}

// ShowLoading displays a loading message
func (l *LoadingState) ShowLoading(message string) {
	if message == "" {
		message = "Loading sequencers..."
	}
	l.TextView.SetText(fmt.Sprintf("[%s]%s[-]", l.theme.WarningColor.String(), message))
}

// ShowEmpty displays an empty state message
func (l *LoadingState) ShowEmpty(message string) {
	if message == "" {
		message = "No sequencers found"
	}
	l.TextView.SetText(fmt.Sprintf("[%s]%s[-]", l.theme.SecondaryColor.String(), message))
}
