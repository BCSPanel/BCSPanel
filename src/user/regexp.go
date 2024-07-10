package user

import "regexp"

var compiledRegexp_UsernameInputFormat = regexp.MustCompile(`^[0-9a-zA-Z\-_]{1,32}$`)
