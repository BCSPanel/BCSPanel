package mylog

import "log"

func init() {
	log.SetOutput(&Writer)
}
