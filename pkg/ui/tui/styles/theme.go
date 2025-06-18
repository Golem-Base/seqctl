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
	MarkColor   tcell.Color
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
		MarkColor:   tcell.ColorPurple,
	}
}

// CatppuccinMocha returns the Catppuccin Mocha theme
// A soothing pastel theme with warm colors on a dark background
func CatppuccinMocha() *Theme {
	return &Theme{
		// Background color - Catppuccin Base
		BackgroundColor: tcell.NewHexColor(0x1e1e2e),

		// Border colors - Catppuccin Surface 1 and Lavender
		BorderColor:      tcell.NewHexColor(0x45475a), // Surface 1
		BorderFocusColor: tcell.NewHexColor(0xb4befe), // Lavender

		// Selection colors - Catppuccin Surface 2 and Text
		SelectedBg: tcell.NewHexColor(0x585b70), // Surface 2
		SelectedFg: tcell.NewHexColor(0xcdd6f4), // Text

		// Table colors - Catppuccin Text and Base
		TableFg:  tcell.NewHexColor(0xcdd6f4), // Text
		TableBg:  tcell.NewHexColor(0x1e1e2e), // Base
		HeaderFg: tcell.NewHexColor(0xb4befe), // Lavender
		HeaderBg: tcell.NewHexColor(0x1e1e2e), // Base

		// Status colors - Catppuccin themed
		SuccessColor: tcell.NewHexColor(0xa6e3a1), // Green
		ErrorColor:   tcell.NewHexColor(0xf38ba8), // Red
		WarningColor: tcell.NewHexColor(0xfab387), // Peach
		InfoColor:    tcell.NewHexColor(0x89dceb), // Sky

		// Text styling colors - Catppuccin themed
		PrimaryColor:   tcell.NewHexColor(0x89b4fa), // Blue
		SecondaryColor: tcell.NewHexColor(0xa6adc8), // Subtext 0
		DangerColor:    tcell.NewHexColor(0xf38ba8), // Red

		// Special colors - Catppuccin themed
		LeaderColor: tcell.NewHexColor(0xf9e2af), // Yellow
		MarkColor:   tcell.NewHexColor(0xcba6f7), // Mauve
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
