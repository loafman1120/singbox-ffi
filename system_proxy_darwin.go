//go:build darwin

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
	savedSystemProxy *darwinSystemProxySnapshot
)

type darwinSystemProxySnapshot struct {
	services []darwinServiceProxySnapshot
}

type darwinServiceProxySnapshot struct {
	name          string
	webEnabled    bool
	webServer     string
	webPort       string
	secureEnabled bool
	secureServer  string
	securePort    string
	bypass        []string
}

func getSystemProxyStatus() (systemProxyStatus, error) {
	systemProxyMu.Lock()
	defer systemProxyMu.Unlock()
	services, err := darwinNetworkServices()
	if err != nil {
		return systemProxyStatus{}, err
	}
	status := systemProxyStatus{
		Platform:      runtime.GOOS,
		Supported:     true,
		HasSavedState: savedSystemProxy != nil,
	}
	for _, service := range services {
		web, _ := darwinReadProxy(service, "-getwebproxy")
		secure, _ := darwinReadProxy(service, "-getsecurewebproxy")
		if web.enabled || secure.enabled {
			status.Enabled = true
			if web.server != "" && web.port != "" {
				status.Server = web.server + ":" + web.port
			} else if secure.server != "" && secure.port != "" {
				status.Server = secure.server + ":" + secure.port
			}
			bypass, _ := darwinReadBypass(service)
			status.Bypass = strings.Join(bypass, ",")
			break
		}
	}
	return status, nil
}

func setSystemProxy(enabled bool, host string, port int32, bypass string) error {
	systemProxyMu.Lock()
	defer systemProxyMu.Unlock()
	if _, err := exec.LookPath("networksetup"); err != nil {
		return errors.New("networksetup is not available")
	}
	if enabled {
		if host == "" {
			return errors.New("system proxy host is required")
		}
		if port <= 0 || port > 65535 {
			return errors.New("system proxy port out of range")
		}
		if savedSystemProxy == nil {
			snapshot, err := darwinSnapshotSystemProxy()
			if err != nil {
				return err
			}
			savedSystemProxy = snapshot
		}
		services, err := darwinNetworkServices()
		if err != nil {
			return err
		}
		portText := strconv.Itoa(int(port))
		for _, service := range services {
			if err := runNetworkSetup("-setwebproxy", service, host, portText); err != nil {
				return err
			}
			if err := runNetworkSetup("-setsecurewebproxy", service, host, portText); err != nil {
				return err
			}
			bypassDomains := darwinBypassDomains(bypass)
			if len(bypassDomains) > 0 {
				args := append([]string{"-setproxybypassdomains", service}, bypassDomains...)
				if err := runNetworkSetup(args...); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if savedSystemProxy != nil {
		err := darwinRestoreSystemProxy(savedSystemProxy)
		savedSystemProxy = nil
		return err
	}
	services, err := darwinNetworkServices()
	if err != nil {
		return err
	}
	for _, service := range services {
		if err := runNetworkSetup("-setwebproxystate", service, "off"); err != nil {
			return err
		}
		if err := runNetworkSetup("-setsecurewebproxystate", service, "off"); err != nil {
			return err
		}
	}
	return nil
}

func darwinSnapshotSystemProxy() (*darwinSystemProxySnapshot, error) {
	services, err := darwinNetworkServices()
	if err != nil {
		return nil, err
	}
	snapshot := &darwinSystemProxySnapshot{
		services: make([]darwinServiceProxySnapshot, 0, len(services)),
	}
	for _, service := range services {
		web, _ := darwinReadProxy(service, "-getwebproxy")
		secure, _ := darwinReadProxy(service, "-getsecurewebproxy")
		bypass, _ := darwinReadBypass(service)
		snapshot.services = append(snapshot.services, darwinServiceProxySnapshot{
			name:          service,
			webEnabled:    web.enabled,
			webServer:     web.server,
			webPort:       web.port,
			secureEnabled: secure.enabled,
			secureServer:  secure.server,
			securePort:    secure.port,
			bypass:        bypass,
		})
	}
	return snapshot, nil
}

func darwinRestoreSystemProxy(snapshot *darwinSystemProxySnapshot) error {
	for _, service := range snapshot.services {
		if service.webServer != "" && service.webPort != "" {
			if err := runNetworkSetup("-setwebproxy", service.name, service.webServer, service.webPort); err != nil {
				return err
			}
		}
		if err := runNetworkSetup("-setwebproxystate", service.name, onOff(service.webEnabled)); err != nil {
			return err
		}
		if service.secureServer != "" && service.securePort != "" {
			if err := runNetworkSetup("-setsecurewebproxy", service.name, service.secureServer, service.securePort); err != nil {
				return err
			}
		}
		if err := runNetworkSetup("-setsecurewebproxystate", service.name, onOff(service.secureEnabled)); err != nil {
			return err
		}
		if len(service.bypass) > 0 {
			args := append([]string{"-setproxybypassdomains", service.name}, service.bypass...)
			if err := runNetworkSetup(args...); err != nil {
				return err
			}
		} else {
			_ = runNetworkSetup("-setproxybypassdomains", service.name, "Empty")
		}
	}
	return nil
}

type darwinProxyConfig struct {
	enabled bool
	server  string
	port    string
}

func darwinReadProxy(service string, command string) (darwinProxyConfig, error) {
	output, err := networkSetupOutput(command, service)
	if err != nil {
		return darwinProxyConfig{}, err
	}
	values := parseNetworkSetupMap(output)
	return darwinProxyConfig{
		enabled: strings.EqualFold(values["Enabled"], "Yes"),
		server:  values["Server"],
		port:    values["Port"],
	}, nil
}

func darwinReadBypass(service string) ([]string, error) {
	output, err := networkSetupOutput("-getproxybypassdomains", service)
	if err != nil {
		return nil, err
	}
	lines := splitNonEmptyLines(output)
	if len(lines) == 1 && strings.Contains(lines[0], "There aren't any bypass domains") {
		return nil, nil
	}
	return lines, nil
}

func darwinNetworkServices() ([]string, error) {
	output, err := networkSetupOutput("-listallnetworkservices")
	if err != nil {
		return nil, err
	}
	var services []string
	for _, line := range splitNonEmptyLines(output) {
		if strings.HasPrefix(line, "An asterisk") || strings.HasPrefix(line, "*") {
			continue
		}
		services = append(services, line)
	}
	if len(services) == 0 {
		return nil, errors.New("no macOS network services found")
	}
	return services, nil
}

func networkSetupOutput(args ...string) (string, error) {
	output, err := exec.Command("networksetup", args...).CombinedOutput()
	if err != nil {
		return "", errors.New(strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func runNetworkSetup(args ...string) error {
	_, err := networkSetupOutput(args...)
	return err
}

func parseNetworkSetupMap(output string) map[string]string {
	values := make(map[string]string)
	for _, line := range splitNonEmptyLines(output) {
		key, value, ok := strings.Cut(line, ":")
		if ok {
			values[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	return values
}

func splitNonEmptyLines(output string) []string {
	var lines []string
	for _, line := range strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func darwinBypassDomains(bypass string) []string {
	var domains []string
	for _, item := range strings.FieldsFunc(bypass, func(r rune) bool {
		return r == ',' || r == ';' || r == ' '
	}) {
		item = strings.TrimSpace(item)
		if item == "" || item == "<local>" {
			continue
		}
		domains = append(domains, item)
	}
	return domains
}

func onOff(value bool) string {
	if value {
		return "on"
	}
	return "off"
}
