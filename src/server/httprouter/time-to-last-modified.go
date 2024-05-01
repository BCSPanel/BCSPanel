package httprouter

import "time"

// Mon, 02 Jan 2006 15:04:05 GMT
func TimeToLastModified(t time.Time) string {
	return t.UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
}
