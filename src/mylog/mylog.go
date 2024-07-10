// 不要让日志异步记录，那样会导致日志顺序错乱

package mylog

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var logFile *os.File
var logWriter io.Writer
var logLock sync.Mutex
var UpdateWriterLock sync.Mutex
var logTY int
var logTM time.Month
var logTD int

// 故意的
func init() {
	gin.SetMode(gin.ReleaseMode)
}

func Init() {
	fmt.Println("Start BCSPanel mylog")
	gin.DisableConsoleColor()

	os.Mkdir("log", 0777)
	err := UpdateWriter()
	if err != nil {
		panic(fmt.Errorf("mylog error update writer: %v", err))
	}

	_, err = logWriter.Write([]byte{'\n'})
	if err != nil {
		panic(fmt.Errorf("mylog error writing file: %v", err))
	}

	// log.SetFlags(log.Ltime)
	INFOln("Start BCSPanel")
}

// 实现日志分日期
func UpdateWriter() (err error) {
	UpdateWriterLock.Lock()
	defer UpdateWriterLock.Unlock()
	timeNow := time.Now()
	y, m, d := timeNow.Date()
	if logFile != nil && d == logTD && m == logTM && y == logTY {
		return nil
	}
	logTY, logTM, logTD = y, m, d
	CloseFile()
	logFile, err = os.OpenFile(fmt.Sprintf("log/%s.txt", timeNow.Format("2006-01-02")), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return
	}
	logWriter = io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(logWriter)
	gin.DefaultWriter = logWriter
	return nil
}

func CloseFile() {
	if logFile != nil {
		logFile.Close()
	}
}

// ln
func Logger_ln(prefix string, v ...any) {
	logLock.Lock()
	defer logLock.Unlock()
	UpdateWriter()
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
	logLock.Lock()
	defer logLock.Unlock()
	UpdateWriter()
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
	logLock.Lock()
	defer logLock.Unlock()
	UpdateWriter()
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
