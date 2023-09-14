package utils

import (
	"regexp"
	"unicode/utf8"
)

const minLen = 3
const maxLen = 48 // IMPORTANT: This is declared in the model as well

var re = regexp.MustCompile(`^[a-z]+[a-z0-9]*$`)

func IsValidTokenName(name string) bool {
	if !re.MatchString(name) {
		return false
	}
	if utf8.RuneCountInString(name) < minLen {
		return false
	}
	if utf8.RuneCountInString(name) > maxLen || len(name) > maxLen {
		return false
	}
	return true
}

func IsValidChannelName(name string) bool {
	return IsValidTokenName(name)
}
