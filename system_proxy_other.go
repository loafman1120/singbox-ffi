//go:build !windows && !darwin && !linux

package main

import (
	"errors"
	"runtime"
)

func getSystemProxyStatus() (systemProxyStatus, error) {
	return systemProxyStatus{
		Platform:  runtime.GOOS,
		Supported: false,
		Enabled:   false,
	}, nil
}

func setSystemProxy(bool, string, int32, string) error {
	return errors.New("system proxy is only implemented on Windows, macOS, and Linux")
}
