package main

import "sync"

type initOptions struct {
	basePath          string
	workingPath       string
	tempPath          string
	locale            string
	commandSecret     string
	commandPort       int32
	logMaxLines       int32
	debug             bool
	oomKillerEnabled  bool
	oomKillerDisabled bool
	oomMemoryLimit    int64
}

var (
	initOptionsMu sync.RWMutex
	currentInit   initOptions
)

func setCurrentInitOptions(options initOptions) {
	initOptionsMu.Lock()
	currentInit = options
	initOptionsMu.Unlock()
}

func getCurrentInitOptions() initOptions {
	initOptionsMu.RLock()
	defer initOptionsMu.RUnlock()
	return currentInit
}
