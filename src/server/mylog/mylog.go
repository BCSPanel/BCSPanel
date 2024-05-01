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

	log.SetFlags(log.Ltime)
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
func logger_ln(prefix string, v ...any) {
	logLock.Lock()
	defer logLock.Unlock()
	UpdateWriter()
	log.SetPrefix(prefix)
	log.Println(v...)
	log.SetPrefix("---- ")
}

func INFOln(v ...any) {
	logger_ln("INFO ", v...)
}

func WARNln(v ...any) {
	logger_ln("WARN ", v...)
}

func ERRORln(v ...any) {
	logger_ln("ERROR ", v...)
}

// f
func logger_f(prefix string, format string, v ...any) {
	logLock.Lock()
	defer logLock.Unlock()
	UpdateWriter()
	log.SetPrefix(prefix)
	log.Printf(format, v...)
	log.SetPrefix("---- ")
}

func INFOf(format string, v ...any) {
	logger_f("INFO ", format, v...)
}

func WARNf(format string, v ...any) {
	logger_f("WARN ", format, v...)
}

func ERRORf(format string, v ...any) {
	logger_f("ERROR ", format, v...)
}

// .
func logger(prefix string, v ...any) {
	logLock.Lock()
	defer logLock.Unlock()
	UpdateWriter()
	log.SetPrefix(prefix)
	log.Print(v...)
	log.SetPrefix("---- ")
}

func INFO(v ...any) {
	logger("INFO ", v...)
}

func WARN(v ...any) {
	logger("WARN ", v...)
}

func ERROR(v ...any) {
	logger("ERROR ", v...)
}
