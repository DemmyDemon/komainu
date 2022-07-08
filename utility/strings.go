package utility

import "unicode"

// UcFirst returns the given string with the first character modified to upper case.
func UcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

// Substring returns a string that is a subset of the given string.
func Substring(input string, start int, length int) string {
	runes := []rune(input)
	origLength := len(runes)
	if length <= 0 {
		return ""
	}
	if start >= origLength {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if start+length > origLength {
		length = origLength - start
	}
	return string(runes[start : start+length])
}
