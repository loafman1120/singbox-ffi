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

Static archive builds for iOS, Android, and advanced desktop integrations use
the same native ABI and should publish platform-specific `singboxffi.a`
artifacts alongside the generated header.

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
dart pub get
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

`lib/singbox_ffi.dart` exports every native symbol in two layers.

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
SingboxRawBindings.openDefault([path])
```

High-level Dart API:

```dart
SingboxFfi.open([path])
SingboxFfi.openDefault([path])
SingboxFfi.fromLibrary(library)
SingboxFfi.process()
SingboxFfi.defaultLibraryName
SingboxFfi.openDefaultLibrary([path])
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
final core = SingboxFfi.openDefault('singboxffi.dll');
```

`SingboxFfi.openDefault([path])` searches in this order:

1. The explicit `path`, when provided.
2. The current working directory.
3. The executable directory from `Platform.executable`.
4. The resolved executable directory from `Platform.resolvedExecutable`.
5. The platform loader default via `DynamicLibrary.open(defaultLibraryName)`.

Static linking is possible, but `SingboxFfi.process()` is not a normal fallback
mode for this plain Dart package. It only works after the native library has
already been linked into the Flutter runner, app executable, or platform plugin
so that symbols such as `sb_version` are visible in the process. Then use:

```dart
final core = SingboxFfi.process();
```

Short term: do not expose `process()` to app users as an ordinary optional
mode unless `singbox-ffi` also provides the platform project that links the
native symbols into the app.

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
- Android: package ABI-specific `.so` files through a Flutter FFI plugin.
- iOS: static archive or framework linking should be provided by a Flutter FFI
  plugin; only then can `SingboxFfi.process()` resolve native symbols.

The next stage for mobile and static mode is to upgrade `singbox_ffi` from a
plain Dart package to a Flutter FFI plugin. That plugin should own
Windows/macOS/Linux native library linking or packaging, Android ABI `.so`
packaging, iOS/macOS static archive or framework linking, and the platform
setup required for `DynamicLibrary.process()` to find `sb_version`.

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

## Dart Package Layout

The repository root is the `singbox_ffi` Dart package. LitheNet should depend
on it directly:

```yaml
dependencies:
  singbox_ffi:
    path: ../singbox-ffi
```

`examples/flutter` is only a smoke/example package. It depends on the root
package with `path: ../..` and should not be consumed as the public API.

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
