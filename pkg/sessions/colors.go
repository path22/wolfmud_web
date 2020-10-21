package sessions

import (
	"strings"
)

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

var Colors = map[string][3]string{
	Reset:     {"white", "", ""},
	Bold:      {"", "", "bold"},
	Normal:    {"", "", "normal"},
	Black:     {"black", "", ""},
	Red:       {"red", "", ""},
	Green:     {"green", "", ""},
	Yellow:    {"yellow", "", ""},
	Blue:      {"blue", "", ""},
	Magenta:   {"magenta", "", ""},
	Cyan:      {"cyan", "", ""},
	White:     {"white", "", ""},
	BGBlack:   {"", "black", ""},
	BGRed:     {"", "red", ""},
	BGGreen:   {"", "green", ""},
	BGYellow:  {"", "yellow", ""},
	BGBlue:    {"", "blue", ""},
	BGMagenta: {"", "magenta", ""},
	BGCyan:    {"", "cyan", ""},
	BGWhite:   {"", "white", ""},
	//Brown     : "",
	//BGBrown   : "",
	//Good   : "",
	//Info   : "",
	//Bad    : "",
	//Prompt : "",
}

func replaceColors(str string) string {
	if str == "" {
		return ""
	}
	str = strings.TrimSpace(str)
	for pattern, style := range Colors {
		if strings.Contains(str, ESC) {
			textColor := style[0]
			bgColor := style[1]
			fontWeight := style[2]
			tag := `</span><span style="color:` + textColor + `;background-color:` + bgColor + `;font-weight:` + fontWeight + `;">`
			str = strings.Replace(str, pattern, tag, -1)
		}
	}
	str = strings.TrimPrefix(str, "</span>")
	str = strings.Replace(str, "\n", "<br>", -1)
	str = "<br>" + str + "</span> "
	return str
}
