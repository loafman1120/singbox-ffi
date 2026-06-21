//go:build linux

package main

import (
	"errors"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	systemProxyMu    sync.Mutex
	savedSystemProxy *linuxSystemProxySnapshot
)

type linuxSystemProxySnapshot struct {
	mode      string
	httpHost  string
	httpPort  string
	httpsHost string
	httpsPort string
	ignore    string
}

func getSystemProxyStatus() (systemProxyStatus, error) {
	systemProxyMu.Lock()
	defer systemProxyMu.Unlock()
	if _, err := exec.LookPath("gsettings"); err != nil {
		return systemProxyStatus{
			Platform:      runtime.GOOS,
			Supported:     false,
			HasSavedState: savedSystemProxy != nil,
		}, nil
	}
	mode, err := gsettingsGet("org.gnome.system.proxy", "mode")
	if err != nil {
		return systemProxyStatus{}, err
	}
	host, _ := gsettingsGet("org.gnome.system.proxy.http", "host")
	port, _ := gsettingsGet("org.gnome.system.proxy.http", "port")
	ignore, _ := gsettingsGet("org.gnome.system.proxy", "ignore-hosts")
	server := ""
	if trimGVariantString(host) != "" && port != "0" {
		server = trimGVariantString(host) + ":" + port
	}
	return systemProxyStatus{
		Platform:      runtime.GOOS,
		Supported:     true,
		Enabled:       trimGVariantString(mode) == "manual",
		Server:        server,
		Bypass:        ignore,
		HasSavedState: savedSystemProxy != nil,
	}, nil
}

func setSystemProxy(enabled bool, host string, port int32, bypass string) error {
	systemProxyMu.Lock()
	defer systemProxyMu.Unlock()
	if _, err := exec.LookPath("gsettings"); err != nil {
		return errors.New("gsettings is not available")
	}
	if enabled {
		if host == "" {
			return errors.New("system proxy host is required")
		}
		if port <= 0 || port > 65535 {
			return errors.New("system proxy port out of range")
		}
		if savedSystemProxy == nil {
			snapshot, err := linuxSnapshotSystemProxy()
			if err != nil {
				return err
			}
			savedSystemProxy = snapshot
		}
		portText := strconv.Itoa(int(port))
		for _, item := range []struct {
			schema string
			key    string
			value  string
		}{
			{"org.gnome.system.proxy.http", "host", quoteGVariantString(host)},
			{"org.gnome.system.proxy.http", "port", portText},
			{"org.gnome.system.proxy.https", "host", quoteGVariantString(host)},
			{"org.gnome.system.proxy.https", "port", portText},
			{"org.gnome.system.proxy", "ignore-hosts", linuxIgnoreHosts(bypass)},
			{"org.gnome.system.proxy", "mode", quoteGVariantString("manual")},
		} {
			if err := gsettingsSet(item.schema, item.key, item.value); err != nil {
				return err
			}
		}
		return nil
	}
	if savedSystemProxy != nil {
		err := linuxRestoreSystemProxy(savedSystemProxy)
		savedSystemProxy = nil
		return err
	}
	return gsettingsSet("org.gnome.system.proxy", "mode", quoteGVariantString("none"))
}

func linuxSnapshotSystemProxy() (*linuxSystemProxySnapshot, error) {
	mode, err := gsettingsGet("org.gnome.system.proxy", "mode")
	if err != nil {
		return nil, err
	}
	httpHost, _ := gsettingsGet("org.gnome.system.proxy.http", "host")
	httpPort, _ := gsettingsGet("org.gnome.system.proxy.http", "port")
	httpsHost, _ := gsettingsGet("org.gnome.system.proxy.https", "host")
	httpsPort, _ := gsettingsGet("org.gnome.system.proxy.https", "port")
	ignore, _ := gsettingsGet("org.gnome.system.proxy", "ignore-hosts")
	return &linuxSystemProxySnapshot{
		mode:      mode,
		httpHost:  httpHost,
		httpPort:  httpPort,
		httpsHost: httpsHost,
		httpsPort: httpsPort,
		ignore:    ignore,
	}, nil
}

func linuxRestoreSystemProxy(snapshot *linuxSystemProxySnapshot) error {
	for _, item := range []struct {
		schema string
		key    string
		value  string
	}{
		{"org.gnome.system.proxy.http", "host", snapshot.httpHost},
		{"org.gnome.system.proxy.http", "port", snapshot.httpPort},
		{"org.gnome.system.proxy.https", "host", snapshot.httpsHost},
		{"org.gnome.system.proxy.https", "port", snapshot.httpsPort},
		{"org.gnome.system.proxy", "ignore-hosts", snapshot.ignore},
		{"org.gnome.system.proxy", "mode", snapshot.mode},
	} {
		if strings.TrimSpace(item.value) == "" {
			continue
		}
		if err := gsettingsSet(item.schema, item.key, item.value); err != nil {
			return err
		}
	}
	return nil
}

func gsettingsGet(schema string, key string) (string, error) {
	output, err := exec.Command("gsettings", "get", schema, key).CombinedOutput()
	if err != nil {
		return "", errors.New(strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

func gsettingsSet(schema string, key string, value string) error {
	output, err := exec.Command("gsettings", "set", schema, key, value).CombinedOutput()
	if err != nil {
		return errors.New(strings.TrimSpace(string(output)))
	}
	return nil
}

func trimGVariantString(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "'")
	value = strings.Trim(value, "\"")
	return value
}

func quoteGVariantString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "\\'") + "'"
}

func linuxIgnoreHosts(bypass string) string {
	items := []string{"'localhost'", "'127.0.0.1'", "'::1'"}
	for _, item := range strings.FieldsFunc(bypass, func(r rune) bool {
		return r == ',' || r == ';' || r == ' '
	}) {
		item = strings.TrimSpace(item)
		if item == "" || item == "<local>" {
			continue
		}
		items = append(items, quoteGVariantString(item))
	}
	return "[" + strings.Join(items, ", ") + "]"
}
