package data

import "strings"

func GenericTrim(s string) string {
	// Characters to trim: space, tab, newline, carriage return, NUL byte, vertical tab
	const cutset = " \t\n\r\000\v"
	return strings.Trim(s, cutset)
}
