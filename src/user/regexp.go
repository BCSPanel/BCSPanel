package user

import "regexp"

var compiledRegexp_UsernameInputFormat = regexp.MustCompile(`^[\w\-]{1,32}$`)
