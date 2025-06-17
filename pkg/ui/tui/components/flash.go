package components

import (
	"fmt"

	"github.com/golem-base/seqctl/pkg/ui/tui/model"
	"github.com/golem-base/seqctl/pkg/ui/tui/styles"
	"github.com/rivo/tview"
)

// FlashMessage displays temporary status messages
type FlashMessage struct {
	*tview.TextView

	flashModel *model.FlashModel
	theme      *styles.Theme
}

// NewFlashMessage creates a new flash message component
func NewFlashMessage(flashModel *model.FlashModel, theme *styles.Theme) *FlashMessage {
	flash := &FlashMessage{
		TextView:   tview.NewTextView(),
		flashModel: flashModel,
		theme:      theme,
	}

	// Configure TextView
	flash.TextView.
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetBorderPadding(0, 0, 1, 1)

	// Register as listener
	flashModel.AddListener(flash)

	// Hide initially
	flash.Clear()

	return flash
}

// OnFlashMessage handles new flash messages
func (f *FlashMessage) OnFlashMessage(msg model.FlashMessage) {
	var color string
	var prefix string

	switch msg.Level {
	case model.FlashInfo:
		color = f.theme.InfoColor.String()
		prefix = "ℹ️"
	case model.FlashSuccess:
		color = f.theme.SuccessColor.String()
		prefix = "✓"
	case model.FlashWarning:
		color = f.theme.WarningColor.String()
		prefix = "⚠️"
	case model.FlashError:
		color = f.theme.ErrorColor.String()
		prefix = "✗"
	}

	text := fmt.Sprintf("[%s]%s %s[-]", color, prefix, msg.Message)
	f.TextView.SetText(text)
}

// OnFlashCleared handles when all messages are cleared
func (f *FlashMessage) OnFlashCleared() {
	f.Clear()
}

// Clear removes the message
func (f *FlashMessage) Clear() {
	f.TextView.SetText("")
}

// GetHeight returns the required height for the flash message
func (f *FlashMessage) GetHeight() int {
	if f.TextView.GetText(false) == "" {
		return 0
	}
	return 1
}
