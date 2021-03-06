package ircf

import (
	"regexp"
	"strconv"
	"strings"
)

// This file is based on https://www.npmjs.com/package/irc-formatting 1.0.0-rc3
//
// The main difference is that the regex follows Daniel Oaks' IRC Formatting specification.

// Chars includes all the codes defined in https://modern.ircdocs.horse/formatting.html
const (
	CharBold          = '\x02'
	CharItalics       = '\x1D'
	CharUnderline     = '\x1F'
	CharStrikethrough = '\x1E'
	CharMonospace     = '\x11'
	CharColor         = '\x03'
	CharHex           = '\x04'
	CharReverseColor  = '\x16'
	CharReset         = '\x0F'
)

var colorRegex = regexp.MustCompile(`\x03(\d\d?)?(?:,(\d\d?))?`)
var replacer = strings.NewReplacer(
	string(CharBold), "",
	string(CharItalics), "",
	string(CharUnderline), "",
	string(CharStrikethrough), "",
	string(CharMonospace), "",
	string(CharColor), "",
	string(CharHex), "",
	string(CharReverseColor), "",
	string(CharReset), "",
)

var Keys = map[rune]string{
	CharBold:      "bold",
	CharItalics:   "italic",
	CharUnderline: "underline",
}

func StripCodes(text string) string {
	return replacer.Replace(colorRegex.ReplaceAllString(text, ""))
}

func StripColor(text string) string {
	return colorRegex.ReplaceAllString(text, "")
}

type color struct {
	foreground int
	background int
	strSize    int
}

func getIndexToColorMap(text string) map[int]color {
	indexToColor := make(map[int]color)
	matches := colorRegex.FindAllStringSubmatchIndex(text, -1)
	for _, match := range matches {
		// The index where the entire colour submatch starts/ends
		startIndex := match[0]
		endIndex := match[1]

		c := color{
			foreground: -1,
			background: -1,
			strSize:    endIndex - startIndex,
		}

		// Errors are impossible, our regex only matches numbers
		if match[2] != -1 {
			c.foreground, _ = strconv.Atoi(text[match[2]:match[3]])

			if match[4] != -1 {
				c.background, _ = strconv.Atoi(text[match[4]:match[5]])
			}
		}

		indexToColor[startIndex] = c
	}
	return indexToColor
}

func Parse(text string) (result []Block) {
	result = []Block{}
	prev := NewBlock("")
	startIndex := 0
	indexToColor := getIndexToColorMap(text)

	// Append a resetter to simplify code a bit
	text += string(CharReset)

	for i, ch := range text {
		current := prev
		updated := true
		nextStart := -1

		switch ch {
		// toggle style
		case CharBold, CharItalics, CharUnderline:
			current.SetField(ch, !prev.GetField(ch))

		// set the colors
		case CharColor:
			color := indexToColor[i]
			current.Foreground = color.foreground
			current.Background = color.background
			nextStart = i + color.strSize

		// reverse the colors
		case CharReverseColor:
			if prev.Foreground != -1 {
				current.Foreground = prev.Background
				current.Background = prev.Foreground

				if current.Foreground == -1 {
					current.Foreground = 0
				}
			}

			current.Reverse = !prev.Reverse

		// reset all formatting
		case CharReset:
			current = Empty

		default:
			updated = false
		}

		if updated {
			prev.Text = text[startIndex:i]

			if nextStart != -1 {
				startIndex = nextStart
			} else {
				startIndex = i + 1
			}

			if len(prev.Text) > 0 {
				result = append(result, prev)
			}

			prev = current
		}
	}

	return result
}
