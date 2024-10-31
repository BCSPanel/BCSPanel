package myregexp

import "regexp"

// [a-z]
var Compiled_az = regexp.MustCompile(`[a-z]`)

// [A-Z]
var Compiled_AZ = regexp.MustCompile(`[A-Z]`)

// [0-9]
var Compiled_09 = regexp.MustCompile(`[0-9]`)
