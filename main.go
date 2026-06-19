package main

/*
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

typedef uint64_t sb_handle;

typedef struct {
	const char *base_path;
	const char *working_path;
	const char *temp_path;
	const char *locale;
	const char *command_secret;
	int32_t command_port;
	int32_t log_max_lines;
	bool debug;
	bool oom_killer_enabled;
	bool oom_killer_disabled;
	int64_t oom_memory_limit;
} sb_init_options;
*/
import "C"

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"unsafe"

	libbox "github.com/sagernet/sing-box/experimental/libbox"
)

func main() {}

//export sb_version
func sb_version() *C.char {
	return C.CString(libbox.Version())
}

//export sb_go_version
func sb_go_version() *C.char {
	return C.CString(libbox.GoVersion())
}

//export sb_free_string
func sb_free_string(ptr *C.char) {
	if ptr != nil {
		C.free(unsafe.Pointer(ptr))
	}
}

//export sb_init
func sb_init(raw *C.sb_init_options, errOut **C.char) C.int32_t {
	clearErr(errOut)
	if raw == nil {
		setErr(errOut, "sb_init: nil options")
		return -1
	}
	opts := fromInitOptions(raw)
	if opts.locale != "" {
		if err := libbox.SetLocale(opts.locale); err != nil {
			setErr(errOut, err.Error())
			return -1
		}
	}
	err := libbox.Setup(&libbox.SetupOptions{
		BasePath:                opts.basePath,
		WorkingPath:             opts.workingPath,
		TempPath:                opts.tempPath,
		CommandServerListenPort: opts.commandPort,
		CommandServerSecret:     opts.commandSecret,
		LogMaxLines:             int(opts.logMaxLines),
		Debug:                   opts.debug,
		CrashReportSource:       "LitheNet",
		OomKillerEnabled:        opts.oomKillerEnabled,
		OomKillerDisabled:       opts.oomKillerDisabled,
		OomMemoryLimit:          opts.oomMemoryLimit,
	})
	if err != nil {
		setErr(errOut, err.Error())
		return -1
	}
	return 0
}

//export sb_check_config
func sb_check_config(configJSON *C.char, errOut **C.char) C.int32_t {
	clearErr(errOut)
	if configJSON == nil {
		setErr(errOut, "sb_check_config: nil config")
		return -1
	}
	if err := libbox.CheckConfig(C.GoString(configJSON)); err != nil {
		setErr(errOut, err.Error())
		return -1
	}
	return 0
}

//export sb_start
func sb_start(configJSON *C.char, out *C.sb_handle, errOut **C.char) C.int32_t {
	clearErr(errOut)
	if configJSON == nil {
		setErr(errOut, "sb_start: nil config")
		return -1
	}
	if out == nil {
		setErr(errOut, "sb_start: nil handle output")
		return -1
	}
	server, err := libbox.NewCommandServer(desktopCommandHandler{}, desktopPlatform{})
	if err != nil {
		setErr(errOut, err.Error())
		return -1
	}
	if err := server.StartOrReloadService(C.GoString(configJSON), &libbox.OverrideOptions{}); err != nil {
		server.Close()
		setErr(errOut, err.Error())
		return -1
	}
	handle := nextHandle.Add(1)
	handlesMu.Lock()
	handles[handle] = &runtimeHandle{server: server}
	handlesMu.Unlock()
	*out = C.sb_handle(handle)
	return 0
}

//export sb_reload
func sb_reload(handle C.sb_handle, configJSON *C.char, errOut **C.char) C.int32_t {
	clearErr(errOut)
	if configJSON == nil {
		setErr(errOut, "sb_reload: nil config")
		return -1
	}
	runtime, ok := getHandle(uint64(handle))
	if !ok {
		setErr(errOut, "sb_reload: invalid handle")
		return -1
	}
	if err := runtime.server.StartOrReloadService(C.GoString(configJSON), &libbox.OverrideOptions{}); err != nil {
		setErr(errOut, err.Error())
		return -1
	}
	return 0
}

//export sb_stop
func sb_stop(handle C.sb_handle, errOut **C.char) C.int32_t {
	clearErr(errOut)
	runtime, ok := getHandle(uint64(handle))
	if !ok {
		setErr(errOut, "sb_stop: invalid handle")
		return -1
	}
	if err := runtime.server.CloseService(); err != nil {
		setErr(errOut, err.Error())
		return -1
	}
	return 0
}

//export sb_free_handle
func sb_free_handle(handle C.sb_handle) C.int32_t {
	handlesMu.Lock()
	runtime, ok := handles[uint64(handle)]
	if ok {
		delete(handles, uint64(handle))
	}
	handlesMu.Unlock()
	if !ok {
		return -1
	}
	runtime.server.Close()
	return 0
}

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

func fromInitOptions(raw *C.sb_init_options) initOptions {
	return initOptions{
		basePath:          cstr(raw.base_path),
		workingPath:       cstr(raw.working_path),
		tempPath:          cstr(raw.temp_path),
		locale:            cstr(raw.locale),
		commandSecret:     cstr(raw.command_secret),
		commandPort:       int32(raw.command_port),
		logMaxLines:       int32(raw.log_max_lines),
		debug:             bool(raw.debug),
		oomKillerEnabled:  bool(raw.oom_killer_enabled),
		oomKillerDisabled: bool(raw.oom_killer_disabled),
		oomMemoryLimit:    int64(raw.oom_memory_limit),
	}
}

func cstr(value *C.char) string {
	if value == nil {
		return ""
	}
	return C.GoString(value)
}

func clearErr(errOut **C.char) {
	if errOut != nil {
		*errOut = nil
	}
}

func setErr(errOut **C.char, message string) {
	if errOut != nil {
		*errOut = C.CString(message)
	}
}

type runtimeHandle struct {
	server *libbox.CommandServer
}

var (
	nextHandle atomic.Uint64
	handlesMu  sync.Mutex
	handles    = map[uint64]*runtimeHandle{}
)

func getHandle(handle uint64) (*runtimeHandle, bool) {
	handlesMu.Lock()
	defer handlesMu.Unlock()
	runtime, ok := handles[handle]
	return runtime, ok
}

type desktopCommandHandler struct{}

func (desktopCommandHandler) ServiceStop() error   { return nil }
func (desktopCommandHandler) ServiceReload() error { return nil }
func (desktopCommandHandler) GetSystemProxyStatus() (*libbox.SystemProxyStatus, error) {
	return &libbox.SystemProxyStatus{Available: false, Enabled: false}, nil
}
func (desktopCommandHandler) SetSystemProxyEnabled(bool) error {
	return errors.New("system proxy is not implemented by LitheNet")
}
func (desktopCommandHandler) TriggerNativeCrash() error { return errors.New("native crash disabled") }
func (desktopCommandHandler) WriteDebugMessage(string)  {}
func (desktopCommandHandler) ConnectSSHAgent() (int32, error) {
	return 0, errors.New("ssh agent is not implemented by LitheNet")
}

type desktopPlatform struct{}

func (desktopPlatform) LocalDNSTransport() libbox.LocalDNSTransport { return nil }
func (desktopPlatform) UsePlatformAutoDetectInterfaceControl() bool { return false }
func (desktopPlatform) AutoDetectInterfaceControl(int32) error      { return nil }
func (desktopPlatform) OpenTun(libbox.TunOptions) (int32, error) {
	return 0, errors.New("tun is not implemented by LitheNet")
}
func (desktopPlatform) UseProcFS() bool { return false }
func (desktopPlatform) FindConnectionOwner(int32, string, int32, string, int32) (*libbox.ConnectionOwner, error) {
	return nil, errors.New("connection owner lookup is not implemented by LitheNet")
}
func (desktopPlatform) StartDefaultInterfaceMonitor(libbox.InterfaceUpdateListener) error {
	return nil
}
func (desktopPlatform) CloseDefaultInterfaceMonitor(libbox.InterfaceUpdateListener) error {
	return nil
}
func (desktopPlatform) GetInterfaces() (libbox.NetworkInterfaceIterator, error) {
	systemInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	interfaces := make([]*libbox.NetworkInterface, 0, len(systemInterfaces))
	for _, systemInterface := range systemInterfaces {
		addresses, _ := systemInterface.Addrs()
		addressStrings := make([]string, 0, len(addresses))
		for _, address := range addresses {
			addressStrings = append(addressStrings, address.String())
		}
		interfaces = append(interfaces, &libbox.NetworkInterface{
			Index:     int32(systemInterface.Index),
			MTU:       int32(systemInterface.MTU),
			Name:      systemInterface.Name,
			Addresses: &stringIterator{values: addressStrings},
			Flags:     int32(systemInterface.Flags),
			Type:      libbox.InterfaceTypeOther,
			DNSServer: &stringIterator{},
		})
	}
	return &networkInterfaceIterator{values: interfaces}, nil
}
func (desktopPlatform) UnderNetworkExtension() bool                              { return false }
func (desktopPlatform) IncludeAllNetworks() bool                                 { return false }
func (desktopPlatform) ReadWIFIState() *libbox.WIFIState                         { return nil }
func (desktopPlatform) SystemCertificates() libbox.StringIterator                { return &stringIterator{} }
func (desktopPlatform) ClearDNSCache()                                           {}
func (desktopPlatform) SendNotification(*libbox.Notification) error              { return nil }
func (desktopPlatform) StartNeighborMonitor(libbox.NeighborUpdateListener) error { return nil }
func (desktopPlatform) CloseNeighborMonitor(libbox.NeighborUpdateListener) error { return nil }
func (desktopPlatform) RegisterMyInterface(string)                               {}
func (desktopPlatform) UsePlatformShell() bool                                   { return false }
func (desktopPlatform) CheckPlatformShell() error {
	return errors.New("platform shell is not implemented by LitheNet")
}
func (desktopPlatform) OpenShellSession(*libbox.PlatformUser, string, libbox.StringIterator, string, int32, int32) (libbox.ShellSession, error) {
	return nil, errors.New("platform shell is not implemented by LitheNet")
}
func (desktopPlatform) LookupUser(string) (*libbox.PlatformUser, error) {
	return nil, errors.New("user lookup is not implemented by LitheNet")
}
func (desktopPlatform) LookupSFTPServer() (string, error) {
	return "", errors.New("sftp is not implemented by LitheNet")
}
func (desktopPlatform) ReadSystemSSHHostKey() (string, error) {
	return "", errors.New("ssh host key is not implemented by LitheNet")
}
func (desktopPlatform) TailscaleHostname() string { return "" }

type stringIterator struct {
	values []string
	index  int
}

func (i *stringIterator) Len() int32    { return int32(len(i.values)) }
func (i *stringIterator) HasNext() bool { return i.index < len(i.values) }
func (i *stringIterator) Next() string {
	if !i.HasNext() {
		return ""
	}
	value := i.values[i.index]
	i.index++
	return value
}

type networkInterfaceIterator struct {
	values []*libbox.NetworkInterface
	index  int
}

func (i *networkInterfaceIterator) HasNext() bool { return i.index < len(i.values) }
func (i *networkInterfaceIterator) Next() *libbox.NetworkInterface {
	if !i.HasNext() {
		return nil
	}
	value := i.values[i.index]
	i.index++
	return value
}
