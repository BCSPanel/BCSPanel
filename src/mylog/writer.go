package mylog

import (
	"fmt"
	"io/fs"
	"os"
	"sync"
	"time"
)

var Writer writer

const FileMode = fs.FileMode(0644)

type writer struct {
	f          *os.File
	updateLock sync.Mutex

	y int
	m time.Month
	d int
}

func (w *writer) updateFile() (ok bool) {
	w.updateLock.Lock()
	defer w.updateLock.Unlock()

	y, m, d := time.Now().Date()
	if w.f != nil {
		if d == w.d && m == w.m && y == w.y {
			return true
		}
		w.f.Close()
	}
	w.y, w.m, w.d = y, m, d

	name := fmt.Sprintf("log/%d-%d-%d.txt", y, m, d)
	exist := false
	stat, err := os.Stat(name)
	if err == nil {
		if stat.IsDir() {
			err = os.RemoveAll(name)
			if err != nil {
				fmt.Printf("mylog remove dir %v error: %v\n", name, err)
				return false
			}
		}
		exist = true
	}

	os.Mkdir("log", FileMode)
	w.f, err = os.OpenFile(
		name,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		FileMode,
	)
	if err != nil {
		fmt.Printf("mylog open %v error: %v\n", name, err)
		return false
	}
	if exist {
		w.f.Write([]byte{'\n'})
	}
	return true
}

func (w *writer) Write(b []byte) (int, error) {
	n, err := os.Stdout.Write(b)
	if !w.updateFile() {
		return n, err
	}
	return w.f.Write(b)
}

func (w *writer) WriteString(s string) (int, error) {
	n, err := os.Stdout.WriteString(s)
	if !w.updateFile() {
		return n, err
	}
	return w.f.WriteString(s)
}
