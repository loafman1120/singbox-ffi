# singbox-ffi

`singbox-ffi` is the native core package for apps such as LitheNet. It builds a
small C ABI wrapper around `github.com/sagernet/sing-box/experimental/libbox`
and publishes native artifacts for GUI apps to download.

LitheNet should not compile this Go code directly. Its Flutter build should
download a prebuilt `singbox-ffi` artifact, copy the native library into the app
bundle, and build only the Flutter UI.

## Outputs

Windows:

```text
singboxffi.dll
singboxffi.h
```

Linux:

```text
libsingboxffi.so
singboxffi.h
```

macOS:

```text
libsingboxffi.dylib
singboxffi.h
```

## Build Locally

Windows with MSYS2 UCRT64 GCC:

```powershell
$env:CGO_ENABLED = "1"
$env:CC = "C:\msys64\ucrt64\bin\gcc.exe"
$env:PATH = "C:\msys64\ucrt64\bin;$env:PATH"

go build -trimpath -buildmode=c-shared `
  -tags "with_gvisor,with_quic,with_wireguard,with_utls,with_naive_outbound,with_purego,with_clash_api,badlinkname,tfogo_checklinkname0" `
  -ldflags "-s -w -buildid= -checklinkname=0" `
  -o build\singboxffi.dll .
```

## Run The Dart Smoke Proxy

The Dart example proves the FFI can start a real local `mixed` proxy on
`127.0.0.1:2080`.

```powershell
cd examples\flutter
dart pub get
dart run bin\proxy.dart ..\..\build\singboxffi.dll
```

In another terminal:

```powershell
curl.exe -x socks5h://127.0.0.1:2080 https://example.com
```

Press `Ctrl+C` in the Dart process to stop the proxy.

## Dart Exports

`examples/flutter/lib/singbox_ffi.dart` exports every native symbol in two
layers.

Raw C ABI bindings:

```dart
SbHandle
SbInitOptions

SbVersionNative / SbVersionDart
SbGoVersionNative / SbGoVersionDart
SbFreeStringNative / SbFreeStringDart
SbInitNative / SbInitDart
SbCheckConfigNative / SbCheckConfigDart
SbStartNative / SbStartDart
SbReloadNative / SbReloadDart
SbStopNative / SbStopDart
SbFreeHandleNative / SbFreeHandleDart

SingboxNativeSymbols
SingboxRawBindings
```

High-level Dart API:

```dart
SingboxFfi.open([path])
SingboxFfi.fromLibrary(library)
SingboxFfi.process()
SingboxFfi.defaultLibraryName
SingboxFfi.raw
SingboxFfi.version()
SingboxFfi.goVersion()
SingboxFfi.init([options])
SingboxFfi.checkConfig(configJson)
SingboxFfi.start(configJson)
SingboxFfi.reload(handle, configJson)
SingboxFfi.stop(handle)
SingboxFfi.freeHandle(handle)
SingboxFfi.freeString(pointer)
SingboxFfi.takeString(pointer)
SingboxFfi.takeError(errOut)

SingboxInitOptions
SingboxService.handle
SingboxService.reload(configJson)
SingboxService.close()
SingboxException
```

## Static Linking

Dynamic linking is the recommended desktop path:

```dart
final core = SingboxFfi.open('singboxffi.dll');
```

Static linking is possible, but the native library must be linked into the
Flutter runner, app executable, or a platform plugin first. Then use:

```dart
final core = SingboxFfi.process();
```

Build a static C archive instead of a dynamic library:

```powershell
go build -trimpath -buildmode=c-archive `
  -tags "with_gvisor,with_quic,with_wireguard,with_utls,with_naive_outbound,with_purego,with_clash_api,badlinkname,tfogo_checklinkname0" `
  -ldflags "-s -w -buildid= -checklinkname=0" `
  -o build\singboxffi.a .
```

Platform notes:

- Windows desktop: prefer `singboxffi.dll`; static linking a Go archive into
  Flutter's MSVC runner is possible only with extra native runner work and
  toolchain care.
- Linux/macOS desktop: static linking works through the native runner build, but
  a shared library is simpler to package and update.
- Android: link the Go archive into a plugin/shared object per ABI; Flutter
  still packages native code inside the APK/AAB.
- iOS: static linking is the normal route; expose symbols to the process and
  use `SingboxFfi.process()`.

## ABI

```c
typedef uint64_t sb_handle;

char *sb_version(void);
char *sb_go_version(void);
void sb_free_string(char *ptr);

int32_t sb_init(const sb_init_options *opts, char **err_out);
int32_t sb_check_config(char *config_json, char **err_out);

int32_t sb_start(char *config_json, sb_handle *out, char **err_out);
int32_t sb_reload(sb_handle handle, char *config_json, char **err_out);
int32_t sb_stop(sb_handle handle, char **err_out);
int32_t sb_free_handle(sb_handle handle);
```

Strings returned by the core must be released with `sb_free_string`. Handles
returned by `sb_start` must be stopped and released with `sb_stop` and
`sb_free_handle`.

## Repository Split

- `loafman1120/singbox-ffi`: builds and releases native core artifacts.
- `loafman1120/LitheNet`: Flutter GUI app; downloads `singbox-ffi` artifacts.

## Status

Implemented:

- config validation
- local mixed/SOCKS/HTTP proxy start, reload, stop
- desktop stub platform interface
- C and Dart smoke examples

Not implemented in this core yet:

- TUN mode
- system proxy toggling
- event/log draining

## License Notice

This project links against `github.com/sagernet/sing-box/experimental/libbox`.
sing-box is distributed under the GNU General Public License, version 3 or later,
with the additional upstream naming restriction copied in `LICENSE.sing-box`.

Distributions of this wrapper, linked binaries, and applications embedding the
produced native library should carry the corresponding GPL notice and must not
use the sing-box name or imply association with the upstream application without
prior consent.
