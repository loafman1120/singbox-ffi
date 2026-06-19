package main

/*
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

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
