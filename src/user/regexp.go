package user

import "regexp"

var RegexpUsernameFormat = regexp.MustCompile(`^[\w\-]{1,32}$`)
