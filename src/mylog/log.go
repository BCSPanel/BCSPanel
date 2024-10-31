package mylog

import (
	"fmt"
	"log"
)

// ln
func Logger_ln(prefix string, v ...any) {
	log.Println("[" + prefix + "] " + fmt.Sprint(v...))
}

func INFOln(v ...any) {
	Logger_ln("INFO", v...)
}

func WARNln(v ...any) {
	Logger_ln("WARN", v...)
}

func ERRORln(v ...any) {
	Logger_ln("ERROR", v...)
}

// f
func Logger_f(prefix string, format string, v ...any) {
	log.Printf("["+prefix+"] "+format, v...)
}

func INFOf(format string, v ...any) {
	Logger_f("INFO", format, v...)
}

func WARNf(format string, v ...any) {
	Logger_f("WARN", format, v...)
}

func ERRORf(format string, v ...any) {
	Logger_f("ERROR", format, v...)
}

// .
func Logger(prefix string, v ...any) {
	log.Print("[" + prefix + "] " + fmt.Sprint(v...))
}

func INFO(v ...any) {
	Logger("INFO", v...)
}

func WARN(v ...any) {
	Logger("WARN", v...)
}

func ERROR(v ...any) {
	Logger("ERROR", v...)
}
