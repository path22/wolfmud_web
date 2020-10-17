package sessions

import "strings"

const (
	ESC       = "\x1b"
	CSI       = ESC + "[" // Control Sequence Introducer
	Reset     = CSI + "0m"
	Bold      = CSI + "1m"
	Normal    = CSI + "22m"
	Black     = CSI + "30m"
	Red       = CSI + "31m"
	Green     = CSI + "32m"
	Yellow    = CSI + "33m"
	Blue      = CSI + "34m"
	Magenta   = CSI + "35m"
	Cyan      = CSI + "36m"
	White     = CSI + "37m"
	BGBlack   = CSI + "40m"
	BGRed     = CSI + "41m"
	BGGreen   = CSI + "42m"
	BGYellow  = CSI + "43m"
	BGBlue    = CSI + "44m"
	BGMagenta = CSI + "45m"
	BGCyan    = CSI + "46m"
	BGWhite   = CSI + "47m"

	// Setup brown as an alias for yellow
	Brown   = Yellow
	BGBrown = BGYellow

	// WolfMUD specific meta colors
	Good   = Green
	Info   = Yellow
	Bad    = Red
	Prompt = Magenta
)

var Colors = map[string]string{
	Reset:     "",
	Bold:      "",
	Normal:    "",
	Black:     "",
	Red:       "",
	Green:     "",
	Yellow:    "",
	Blue:      "",
	Magenta:   "",
	Cyan:      "",
	White:     "",
	BGBlack:   "",
	BGRed:     "",
	BGGreen:   "",
	BGYellow:  "",
	BGBlue:    "",
	BGMagenta: "",
	BGCyan:    "",
	BGWhite:   "",
	//Brown     : "",
	//BGBrown   : "",
	//Good   : "",
	//Info   : "",
	//Bad    : "",
	//Prompt : "",
}

func replaceColors(str string) string {
	for color, hexColor := range Colors {
		if strings.Contains(str, ESC) {
			str = strings.Replace(str, color, hexColor, -1)
		}

	}
	return str
}
