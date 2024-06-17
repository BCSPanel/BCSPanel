package httprouter

import (
	"regexp"
)

var (
	compiledRegExp_myRouter_moreSlash = regexp.MustCompile(`//+`)

	compiledRegExp_myRouter_notLoggedIn401 *regexp.Regexp = regexp.MustCompile(`^/(api|assets|icon)(/|$)`)
)

func updateRegexp() {
}
