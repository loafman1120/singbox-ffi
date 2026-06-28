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
	"unsafe"

	"github.com/sagernet/sing-box/daemon"
	libbox "github.com/sagernet/sing-box/experimental/libbox"
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
	setCurrentInitOptions(opts)
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
	service := newDesktopStartedService()
	if err := service.StartOrReloadService(C.GoString(configJSON), &daemon.OverrideOptions{}); err != nil {
		service.Close()
		setErr(errOut, err.Error())
		return -1
	}
	runtime := newRuntimeHandle(service)
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
	if err := runtime.service.StartOrReloadService(C.GoString(configJSON), &daemon.OverrideOptions{}); err != nil {
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
	if err := runtime.service.CloseService(); err != nil {
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
	if _, err := runtime.service.ClearLogs(context.Background(), &emptypb.Empty{}); err != nil {
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

//export sb_system_proxy_status
func sb_system_proxy_status(jsonOut **C.char, errOut **C.char) C.int32_t {
	clearErr(errOut)
	if jsonOut == nil {
		setErr(errOut, "sb_system_proxy_status: nil output")
		return -1
	}
	*jsonOut = nil
	status, err := getSystemProxyStatus()
	if err != nil {
		setErr(errOut, err.Error())
		return -1
	}
	payload, err := json.Marshal(status)
	if err != nil {
		setErr(errOut, err.Error())
		return -1
	}
	*jsonOut = C.CString(string(payload))
	return 0
}

//export sb_set_system_proxy
func sb_set_system_proxy(host *C.char, port C.int32_t, bypass *C.char, enabled C.bool, errOut **C.char) C.int32_t {
	clearErr(errOut)
	if err := setSystemProxy(bool(enabled), C.GoString(host), int32(port), C.GoString(bypass)); err != nil {
		setErr(errOut, err.Error())
		return -1
	}
	return 0
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
