package myregexp

import "regexp"

func MatchString(pattern string, s string) bool {
	b, _ := regexp.MatchString(pattern, s)
	return b
}

func Match(pattern string, b []byte) bool {
	BOOL, _ := regexp.Match(pattern, b)
	return BOOL
}
