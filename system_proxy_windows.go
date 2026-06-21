//go:build windows

package main

import (
	"errors"
	"fmt"
	"runtime"
	"sync"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const internetSettingsKey = `Software\Microsoft\Windows\CurrentVersion\Internet Settings`

var (
	systemProxyMu         sync.Mutex
	savedSystemProxy      *systemProxySnapshot
	wininet               = windows.NewLazySystemDLL("wininet.dll")
	procInternetSetOption = wininet.NewProc("InternetSetOptionW")
)

type systemProxySnapshot struct {
	hadEnable bool
	enabled   bool
	hadServer bool
	server    string
	hadBypass bool
	bypass    string
}

func getSystemProxyStatus() (systemProxyStatus, error) {
	systemProxyMu.Lock()
	defer systemProxyMu.Unlock()
	enabled, server, bypass, err := readSystemProxy()
	if err != nil {
		return systemProxyStatus{}, err
	}
	return systemProxyStatus{
		Platform:      runtime.GOOS,
		Supported:     true,
		Enabled:       enabled,
		Server:        server,
		Bypass:        bypass,
		HasSavedState: savedSystemProxy != nil,
	}, nil
}

func setSystemProxy(enabled bool, host string, port int32, bypass string) error {
	systemProxyMu.Lock()
	defer systemProxyMu.Unlock()
	if enabled {
		if host == "" {
			return errors.New("system proxy host is required")
		}
		if port <= 0 || port > 65535 {
			return fmt.Errorf("system proxy port out of range: %d", port)
		}
		if savedSystemProxy == nil {
			snapshot, err := snapshotSystemProxy()
			if err != nil {
				return err
			}
			savedSystemProxy = snapshot
		}
		return writeSystemProxy(true, fmt.Sprintf("%s:%d", host, port), bypass)
	}
	if savedSystemProxy != nil {
		err := restoreSystemProxy(savedSystemProxy)
		savedSystemProxy = nil
		return err
	}
	return writeSystemProxy(false, "", "")
}

func openInternetSettings(access uint32) (registry.Key, error) {
	return registry.OpenKey(registry.CURRENT_USER, internetSettingsKey, access)
}

func readSystemProxy() (enabled bool, server string, bypass string, err error) {
	key, err := openInternetSettings(registry.QUERY_VALUE)
	if err != nil {
		return false, "", "", err
	}
	defer key.Close()
	if value, _, valueErr := key.GetIntegerValue("ProxyEnable"); valueErr == nil {
		enabled = value != 0
	}
	server, _, _ = key.GetStringValue("ProxyServer")
	bypass, _, _ = key.GetStringValue("ProxyOverride")
	return enabled, server, bypass, nil
}

func snapshotSystemProxy() (*systemProxySnapshot, error) {
	key, err := openInternetSettings(registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer key.Close()
	snapshot := new(systemProxySnapshot)
	if value, _, valueErr := key.GetIntegerValue("ProxyEnable"); valueErr == nil {
		snapshot.hadEnable = true
		snapshot.enabled = value != 0
	}
	if value, _, valueErr := key.GetStringValue("ProxyServer"); valueErr == nil {
		snapshot.hadServer = true
		snapshot.server = value
	}
	if value, _, valueErr := key.GetStringValue("ProxyOverride"); valueErr == nil {
		snapshot.hadBypass = true
		snapshot.bypass = value
	}
	return snapshot, nil
}

func writeSystemProxy(enabled bool, server string, bypass string) error {
	key, err := openInternetSettings(registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	if err := key.SetDWordValue("ProxyEnable", boolDWord(enabled)); err != nil {
		return err
	}
	if enabled {
		if err := key.SetStringValue("ProxyServer", server); err != nil {
			return err
		}
		if bypass != "" {
			if err := key.SetStringValue("ProxyOverride", bypass); err != nil {
				return err
			}
		} else {
			_ = key.DeleteValue("ProxyOverride")
		}
	}
	return notifySystemProxyChanged()
}

func restoreSystemProxy(snapshot *systemProxySnapshot) error {
	key, err := openInternetSettings(registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	if snapshot.hadEnable {
		if err := key.SetDWordValue("ProxyEnable", boolDWord(snapshot.enabled)); err != nil {
			return err
		}
	} else {
		_ = key.DeleteValue("ProxyEnable")
	}
	if snapshot.hadServer {
		if err := key.SetStringValue("ProxyServer", snapshot.server); err != nil {
			return err
		}
	} else {
		_ = key.DeleteValue("ProxyServer")
	}
	if snapshot.hadBypass {
		if err := key.SetStringValue("ProxyOverride", snapshot.bypass); err != nil {
			return err
		}
	} else {
		_ = key.DeleteValue("ProxyOverride")
	}
	return notifySystemProxyChanged()
}

func boolDWord(value bool) uint32 {
	if value {
		return 1
	}
	return 0
}

func notifySystemProxyChanged() error {
	const (
		internetOptionRefresh         = 37
		internetOptionSettingsChanged = 39
	)
	for _, option := range []uintptr{internetOptionSettingsChanged, internetOptionRefresh} {
		result, _, callErr := procInternetSetOption.Call(0, option, 0, 0)
		if result == 0 {
			return callErr
		}
	}
	return nil
}
