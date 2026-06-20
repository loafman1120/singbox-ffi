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
	"context"
	"encoding/json"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/sagernet/sing-box/daemon"
	libbox "github.com/sagernet/sing-box/experimental/libbox"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
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
		CrashReportSource:       "singbox-ffi",
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
	runtime := newRuntimeHandle(server)
	runtime.startLogSubscription()
	handle := nextHandle.Add(1)
	handlesMu.Lock()
	handles[handle] = runtime
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
		runtime.setLastError(err.Error())
		setErr(errOut, err.Error())
		return -1
	}
	runtime.setState(serviceStateRunning)
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
	if runtime.state() == serviceStateStopped {
		return 0
	}
	if err := runtime.server.CloseService(); err != nil {
		runtime.setLastError(err.Error())
		setErr(errOut, err.Error())
		return -1
	}
	runtime.setState(serviceStateStopped)
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
	runtime.close()
	return 0
}

//export sb_drain_logs
func sb_drain_logs(handle C.sb_handle, maxEntries C.int32_t, jsonOut **C.char, errOut **C.char) C.int32_t {
	clearErr(errOut)
	if jsonOut == nil {
		setErr(errOut, "sb_drain_logs: nil output")
		return -1
	}
	*jsonOut = nil
	runtime, ok := getHandle(uint64(handle))
	if !ok {
		setErr(errOut, "sb_drain_logs: invalid handle")
		return -1
	}
	entries := runtime.logs.drain(int(maxEntries))
	if entries == nil {
		entries = []logEvent{}
	}
	payload, err := json.Marshal(entries)
	if err != nil {
		setErr(errOut, err.Error())
		return -1
	}
	*jsonOut = C.CString(string(payload))
	return 0
}

//export sb_clear_logs
func sb_clear_logs(handle C.sb_handle, errOut **C.char) C.int32_t {
	clearErr(errOut)
	runtime, ok := getHandle(uint64(handle))
	if !ok {
		setErr(errOut, "sb_clear_logs: invalid handle")
		return -1
	}
	runtime.logs.clear()
	if _, err := runtime.server.StartedService.ClearLogs(context.Background(), &emptypb.Empty{}); err != nil {
		setErr(errOut, err.Error())
		return -1
	}
	return 0
}

//export sb_service_state
func sb_service_state(handle C.sb_handle, jsonOut **C.char, errOut **C.char) C.int32_t {
	clearErr(errOut)
	if jsonOut == nil {
		setErr(errOut, "sb_service_state: nil output")
		return -1
	}
	*jsonOut = nil
	runtime, ok := getHandle(uint64(handle))
	if !ok {
		setErr(errOut, "sb_service_state: invalid handle")
		return -1
	}
	payload, err := json.Marshal(runtime.snapshot())
	if err != nil {
		setErr(errOut, err.Error())
		return -1
	}
	*jsonOut = C.CString(string(payload))
	return 0
}

func newRuntimeHandle(server *libbox.CommandServer) *runtimeHandle {
	ctx, cancel := context.WithCancel(context.Background())
	return &runtimeHandle{
		server:    server,
		logs:      newLogBuffer(),
		logCtx:    ctx,
		logCancel: cancel,
		nowState:  serviceStateRunning,
	}
}

func (h *runtimeHandle) close() {
	if h.logCancel != nil {
		h.logCancel()
	}
	h.setState(serviceStateClosed)
	h.server.Close()
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
	server    *libbox.CommandServer
	logs      *logBuffer
	logCtx    context.Context
	logCancel context.CancelFunc
	stateMu   sync.RWMutex
	nowState  serviceState
	lastError string
}

type serviceState string

const (
	serviceStateRunning serviceState = "running"
	serviceStateStopped serviceState = "stopped"
	serviceStateClosed  serviceState = "closed"
)

type serviceSnapshot struct {
	State     serviceState `json:"state"`
	Running   bool         `json:"running"`
	Closed    bool         `json:"closed"`
	LastError string       `json:"lastError,omitempty"`
}

func (h *runtimeHandle) state() serviceState {
	h.stateMu.RLock()
	defer h.stateMu.RUnlock()
	return h.nowState
}

func (h *runtimeHandle) setState(state serviceState) {
	h.stateMu.Lock()
	h.nowState = state
	h.stateMu.Unlock()
}

func (h *runtimeHandle) setLastError(message string) {
	h.stateMu.Lock()
	h.lastError = message
	h.stateMu.Unlock()
}

func (h *runtimeHandle) snapshot() serviceSnapshot {
	h.stateMu.RLock()
	defer h.stateMu.RUnlock()
	return serviceSnapshot{
		State:     h.nowState,
		Running:   h.nowState == serviceStateRunning,
		Closed:    h.nowState == serviceStateClosed,
		LastError: h.lastError,
	}
}

func (h *runtimeHandle) startLogSubscription() {
	go func() {
		_ = h.server.StartedService.SubscribeLog(&emptypb.Empty{}, &logStreamServer{
			ctx:    h.logCtx,
			buffer: h.logs,
		})
	}()
}

type logEvent struct {
	Sequence  uint64 `json:"sequence"`
	Level     int32  `json:"level,omitempty"`
	LevelName string `json:"levelName,omitempty"`
	Message   string `json:"message,omitempty"`
	Reset     bool   `json:"reset,omitempty"`
}

type logBuffer struct {
	mu      sync.Mutex
	nextSeq uint64
	entries []logEvent
}

func newLogBuffer() *logBuffer {
	return &logBuffer{}
}

func (b *logBuffer) appendReset() {
	b.mu.Lock()
	b.nextSeq++
	b.entries = append(b.entries, logEvent{
		Sequence: b.nextSeq,
		Reset:    true,
	})
	b.mu.Unlock()
}

func (b *logBuffer) appendMessage(level int32, message string) {
	b.mu.Lock()
	b.nextSeq++
	b.entries = append(b.entries, logEvent{
		Sequence:  b.nextSeq,
		Level:     level,
		LevelName: logLevelName(level),
		Message:   message,
	})
	b.mu.Unlock()
}

func (b *logBuffer) drain(maxEntries int) []logEvent {
	b.mu.Lock()
	defer b.mu.Unlock()
	if maxEntries <= 0 || maxEntries >= len(b.entries) {
		entries := b.entries
		b.entries = nil
		return entries
	}
	entries := append([]logEvent(nil), b.entries[:maxEntries]...)
	b.entries = append([]logEvent(nil), b.entries[maxEntries:]...)
	return entries
}

func (b *logBuffer) clear() {
	b.mu.Lock()
	b.entries = nil
	b.mu.Unlock()
}

func logLevelName(level int32) string {
	switch level {
	case -1:
		return "disabled"
	case 0:
		return "panic"
	case 1:
		return "fatal"
	case 2:
		return "error"
	case 3:
		return "warn"
	case 4:
		return "info"
	case 5:
		return "debug"
	case 6:
		return "trace"
	default:
		return "unknown"
	}
}

type logStreamServer struct {
	grpc.ServerStream
	ctx    context.Context
	buffer *logBuffer
}

func (s *logStreamServer) Send(message *daemon.Log) error {
	if message.Reset_ {
		s.buffer.appendReset()
	}
	for _, entry := range message.Messages {
		s.buffer.appendMessage(int32(entry.Level), entry.Message)
	}
	return nil
}

func (s *logStreamServer) SetHeader(metadata.MD) error  { return nil }
func (s *logStreamServer) SendHeader(metadata.MD) error { return nil }
func (s *logStreamServer) SetTrailer(metadata.MD)       {}
func (s *logStreamServer) Context() context.Context     { return s.ctx }
func (s *logStreamServer) SendMsg(any) error            { return nil }
func (s *logStreamServer) RecvMsg(any) error            { return nil }

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
	return errors.New("system proxy is not implemented by singbox-ffi")
}
func (desktopCommandHandler) TriggerNativeCrash() error { return errors.New("native crash disabled") }
func (desktopCommandHandler) WriteDebugMessage(string)  {}
func (desktopCommandHandler) ConnectSSHAgent() (int32, error) {
	return 0, errors.New("ssh agent is not implemented by singbox-ffi")
}

type desktopPlatform struct{}

func (desktopPlatform) LocalDNSTransport() libbox.LocalDNSTransport { return nil }
func (desktopPlatform) UsePlatformAutoDetectInterfaceControl() bool { return false }
func (desktopPlatform) AutoDetectInterfaceControl(int32) error      { return nil }
func (desktopPlatform) OpenTun(libbox.TunOptions) (int32, error) {
	return 0, errors.New("tun is not implemented by singbox-ffi")
}
func (desktopPlatform) UseProcFS() bool { return false }
func (desktopPlatform) FindConnectionOwner(int32, string, int32, string, int32) (*libbox.ConnectionOwner, error) {
	return nil, errors.New("connection owner lookup is not implemented by singbox-ffi")
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
	return errors.New("platform shell is not implemented by singbox-ffi")
}
func (desktopPlatform) OpenShellSession(*libbox.PlatformUser, string, libbox.StringIterator, string, int32, int32) (libbox.ShellSession, error) {
	return nil, errors.New("platform shell is not implemented by singbox-ffi")
}
func (desktopPlatform) LookupUser(string) (*libbox.PlatformUser, error) {
	return nil, errors.New("user lookup is not implemented by singbox-ffi")
}
func (desktopPlatform) LookupSFTPServer() (string, error) {
	return "", errors.New("sftp is not implemented by singbox-ffi")
}
func (desktopPlatform) ReadSystemSSHHostKey() (string, error) {
	return "", errors.New("ssh host key is not implemented by singbox-ffi")
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
