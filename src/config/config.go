package config

import (
	"os"
	"path/filepath"
	"sync"
)

const userConfigDir = "config"
const createFileMode = os.FileMode(0644)

var lock sync.Mutex

func Update() {
	UpdateHttp()
}

func create(name string, data []byte) {
	os.Mkdir(userConfigDir, createFileMode)
	os.WriteFile(filepath.Join(userConfigDir, name), data, createFileMode)
}
