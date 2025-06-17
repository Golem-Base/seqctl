package styles

import (
	"github.com/gdamore/tcell/v2"
)

// Theme defines the color scheme and styling for the TUI
type Theme struct {
	// Background color
	BackgroundColor tcell.Color

	// Border colors
	BorderColor      tcell.Color
	BorderFocusColor tcell.Color

	// Selection colors
	SelectedBg tcell.Color
	SelectedFg tcell.Color

	// Table colors
	TableFg  tcell.Color
	TableBg  tcell.Color
	HeaderFg tcell.Color
	HeaderBg tcell.Color

	// Status colors
	SuccessColor tcell.Color
	ErrorColor   tcell.Color
	WarningColor tcell.Color
	InfoColor    tcell.Color

	// Text styling colors
	PrimaryColor   tcell.Color
	SecondaryColor tcell.Color
	DangerColor    tcell.Color

	// Special colors
	LeaderColor tcell.Color
}

// Default returns the default theme
func Default() *Theme {
	return &Theme{
		// Background color
		BackgroundColor: tcell.ColorBlack,

		// Border colors
		BorderColor:      tcell.ColorDodgerBlue,
		BorderFocusColor: tcell.ColorAqua,

		// Selection colors
		SelectedBg: tcell.ColorAqua,
		SelectedFg: tcell.ColorBlack,

		// Table colors
		TableFg:  tcell.ColorBlue,
		TableBg:  tcell.ColorBlack,
		HeaderFg: tcell.ColorWhite,
		HeaderBg: tcell.ColorBlack,

		// Status colors
		SuccessColor: tcell.ColorGreen,
		ErrorColor:   tcell.ColorOrangeRed,
		WarningColor: tcell.ColorDarkOrange,
		InfoColor:    tcell.ColorLightSkyBlue,

		// Text styling colors
		PrimaryColor:   tcell.ColorAqua,
		SecondaryColor: tcell.ColorWhite,
		DangerColor:    tcell.ColorRed,

		// Special colors
		LeaderColor: tcell.ColorGold,
	}
}

// Icons defines the icons used in the UI
type Icons struct {
	Network  string
	Active   string
	Inactive string
	Healthy  string
	Leader   string
	Empty    string
}

// DefaultIcons returns the default icon set
func DefaultIcons() *Icons {
	return &Icons{
		Network:  "🌐",
		Active:   "✓",
		Inactive: "✗",
		Healthy:  "💚",
		Leader:   "👑",
		Empty:    "—",
	}
}
