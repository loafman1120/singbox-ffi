# singbox-ffi

`singbox-ffi` is a small C ABI wrapper around `github.com/sagernet/sing-box/experimental/libbox`.

The goal is to use the public Go API directly and expose a stable, tiny desktop ABI for C#, Rust, Flutter, Electron native modules, or any other host that can load a native library.

This project intentionally does not depend on upstream `sing-ffi`. See [SagerNet/sing-box#4016](https://github.com/SagerNet/sing-box/issues/4016): the upstream `lib_windows`, `lib_android_new`, and `lib_apple_new` Makefile targets are wired to `sing-ffi`, but that generator is not publicly available.

## Decision

Use an external Go project:

```text
singbox-ffi
  go.mod
  main.go
  platform_desktop.go      # future: libbox PlatformInterface
  handler.go               # future: CommandServerHandler
```

Import upstream `libbox` directly:

```go
import libbox "github.com/sagernet/sing-box/experimental/libbox"
```

Do not modify upstream `experimental/libbox`.

## Why Not Implement sing-ffi

The upstream repository already has a generation design in `experimental/libbox/ffi.json`:

- Android AAR
- Apple XCFramework
- Windows C# NuGet

However, the generator command is external:

```make
lib_windows:
	$(SING_FFI) generate --config $(LIBBOX_FFI_CONFIG) --platform-type csharp
```

Because `sing-ffi` is not public, a third party cannot reproduce that official pipeline today. This project takes the maintainable route: a hand-written narrow C ABI.

## ABI Principles

- Keep C ABI small.
- Keep Go side free to use `libbox` normally.
- Never pass Go pointers or Go objects to C.
- Cross the boundary with integers, booleans, UTF-8 strings, JSON, byte buffers, and opaque handles.
- Whoever allocates a string frees it. Strings returned by this DLL must be released with `sb_free_string`.
- Treat the C ABI as the stable contract. Treat `libbox` as an internal implementation detail.

## Public ABI

Phase 0, already scaffolded:

```c
const char *sb_version(void);
const char *sb_go_version(void);
void sb_free_string(char *ptr);

int32_t sb_init(const sb_init_options *opts, char **err_out);
int32_t sb_check_config(const char *config_json, char **err_out);
```

Target Phase 1:

```c
typedef uint64_t sb_handle;

int32_t sb_start(const char *config_json, sb_handle *out, char **err_out);
int32_t sb_reload(sb_handle handle, const char *config_json, char **err_out);
int32_t sb_stop(sb_handle handle, char **err_out);
int32_t sb_free_handle(sb_handle handle);
int32_t sb_drain_events(sb_handle handle, char **json_out, char **err_out);
```

## Lifecycle

Expected host flow:

```text
load DLL
  -> sb_init(paths, locale, command port, secret, log options)
  -> sb_check_config(config JSON)
  -> sb_start(config JSON)
  -> optional sb_reload(config JSON)
  -> optional sb_drain_events()
  -> sb_stop(handle)
  -> sb_free_handle(handle)
```

## libbox Mapping

The wrapper should call upstream APIs directly:

```go
libbox.SetLocale(...)
libbox.Setup(...)
libbox.CheckConfig(config)
libbox.NewCommandServer(handler, platform)
server.Start()
server.StartOrReloadService(config, options)
server.CloseService()
server.Close()
```

## Runtime Model

The host only receives an opaque `uint64_t` handle. Go maintains a registry:

```go
type runtimeHandle struct {
	server *libbox.CommandServer
	events chan event
}
```

Use a mutex-protected map:

```go
map[uint64]*runtimeHandle
```

This avoids leaking Go pointers across C ABI and keeps C#, Rust, and Flutter bindings simple.

## Phase 1: Local Proxy MVP

Start with local proxy mode only:

- `mixed`, `socks`, or `http` inbound
- no `tun` inbound
- no route manipulation
- no elevated privileges
- command server bound to `127.0.0.1:<port>`

For desktop apps, this is the fastest useful target. The host can optionally set system proxy itself.

## Phase 1 PlatformInterface

`libbox.NewCommandServer` needs a `libbox.PlatformInterface`. For the local proxy MVP, implement a desktop stub:

- `OpenTun`: return unsupported.
- `UseProcFS`: false on Windows.
- `GetInterfaces`: map `net.Interfaces()` into libbox network interface objects.
- `StartDefaultInterfaceMonitor`: no-op at first, then add polling later.
- `FindConnectionOwner`: unsupported at first.
- `SystemCertificates`: return empty at first.
- notifications, Wi-Fi, shell, neighbor monitor: no-op or unsupported.

This keeps the first release focused and prevents TUN from blocking the project.

## Phase 2: Observability

Add event draining:

```c
int32_t sb_drain_events(sb_handle handle, char **json_out, char **err_out);
```

Return a JSON array:

```json
[
  {"type":"log","level":"info","message":"started"},
  {"type":"status","uplink":1234,"downlink":5678},
  {"type":"error","message":"..."}
]
```

Polling is easier and safer than cross-thread callbacks for C#/Flutter/Rust.

## Phase 3: Host Bindings

C# P/Invoke:

```csharp
[DllImport("singboxffi", CallingConvention = CallingConvention.Cdecl)]
static extern IntPtr sb_version();
```

Rust:

```rust
extern "C" {
    fn sb_version() -> *mut std::os::raw::c_char;
}
```

Flutter:

```dart
final version = dylib.lookupFunction<Pointer<Utf8> Function(), Pointer<Utf8> Function()>('sb_version');
```

All bindings should call `sb_free_string` for returned strings after copying them.

## Phase 4: TUN Mode

Add TUN only after local proxy mode is stable.

Windows TUN mode needs separate engineering:

- Wintun or another supported TUN provider
- admin or service privileges
- route and DNS management
- rollback on crash
- default interface monitoring
- interface owner/process lookup if required

This should be a separate milestone because it changes permissions, installer design, and failure modes.

## Build

Requirements:

- Go 1.26 or newer.
- A C toolchain for `-buildmode=c-shared`.
- On Windows, install a GCC-compatible toolchain such as MSYS2/MinGW-w64 and make sure `gcc` is in `PATH`.

Windows:

```powershell
$env:CGO_ENABLED="1"
go build -trimpath -buildmode=c-shared `
  -tags "with_gvisor,with_quic,with_wireguard,with_utls,with_naive_outbound,with_purego,with_clash_api,badlinkname,tfogo_checklinkname0" `
  -ldflags "-s -w -buildid=" `
  -o build\singboxffi.dll .
```

Outputs:

```text
build/singboxffi.dll
build/singboxffi.h
```

Linux:

```bash
CGO_ENABLED=1 go build -trimpath -buildmode=c-shared \
  -tags "with_gvisor,with_quic,with_wireguard,with_utls,with_clash_api" \
  -ldflags "-s -w -buildid=" \
  -o build/libsingboxffi.so .
```

macOS:

```bash
CGO_ENABLED=1 go build -trimpath -buildmode=c-shared \
  -tags "with_gvisor,with_quic,with_wireguard,with_utls,with_clash_api" \
  -ldflags "-s -w -buildid=" \
  -o build/libsingboxffi.dylib .
```

## Version Strategy

Pin `github.com/sagernet/sing-box` to a release tag or commit.

This repository currently tracks the upstream testing line:

```bash
go get github.com/sagernet/sing-box@17852ccaaa494f664cbe90e59e2d1558c7f1db34
```

For a stable release line, switch to a released tag after checking that `experimental/libbox` and its transitive dependencies compile together:

```bash
go get github.com/sagernet/sing-box@v1.13.13
```

For testing upstream `testing`:

```go
replace github.com/sagernet/sing-box => C:\Users\kangj\Documents\Project\sing-box
```

Do not expose upstream Go API directly to host applications. Host applications should only depend on this C ABI.

## Milestones

1. Phase 0: `version`, `go_version`, `init`, `check_config`.
2. Phase 1: local proxy `start`, `reload`, `stop`, handle registry.
3. Phase 2: event queue and JSON draining.
4. Phase 3: C# and Flutter binding examples.
5. Phase 4: Windows TUN implementation.
6. Phase 5: package as NuGet / Flutter plugin / Rust crate helper.

## Safety Notes

- Bind command server to localhost only.
- Always use a command secret.
- Keep string ownership explicit.
- Never store host object references in Go callbacks.
- Prefer polling events over callbacks for the first desktop release.
- Make TUN mode opt-in and separate from local proxy mode.
