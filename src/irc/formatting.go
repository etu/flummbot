package irc

import "fmt"

//
// https://modern.ircdocs.horse/formatting.html
//
type Format struct {
	Bold          string // 0x02
	Color         string // 0x03
	HexColor      string // 0x04
	Italics       string // 0x1D
	Monospace     string // 0x11
	Reset         string // 0x0F
	ReverseColor  string // 0x16
	Strikethrough string // 0x1E
	Underline     string // 0x1F
	Colors        Colors // Struct with named colors
}

type Colors struct {
	White      string // 00
	Black      string // 01
	Blue       string // 02
	Green      string // 03
	Red        string // 04
	Brown      string // 05
	Magenta    string // 06
	Orange     string // 07
	Yellow     string // 08
	LightGreen string // 09
	Cyan       string // 10
	LightCyan  string // 11
	LightBlue  string // 12
	Pink       string // 13
	Gray       string // 14
	LightGray  string // 15
}

var formatting Format

func GetFormat() Format {
	if formatting.Bold == "" {
		formatting = Format{
			Bold:          fmt.Sprintf("%c", 0x02),
			Color:         fmt.Sprintf("%c", 0x03),
			HexColor:      fmt.Sprintf("%c", 0x04),
			Italics:       fmt.Sprintf("%c", 0x1D),
			Monospace:     fmt.Sprintf("%c", 0x11),
			Reset:         fmt.Sprintf("%c", 0x0F),
			ReverseColor:  fmt.Sprintf("%c", 0x16),
			Strikethrough: fmt.Sprintf("%c", 0x1E),
			Underline:     fmt.Sprintf("%c", 0x1F),
			Colors: Colors{
				White:      "00",
				Black:      "01",
				Blue:       "02",
				Green:      "03",
				Red:        "04",
				Brown:      "05",
				Magenta:    "06",
				Orange:     "07",
				Yellow:     "08",
				LightGreen: "09",
				Cyan:       "10",
				LightCyan:  "11",
				LightBlue:  "12",
				Pink:       "13",
				Gray:       "14",
				LightGray:  "15",
			},
		}
	}

	return formatting
}
