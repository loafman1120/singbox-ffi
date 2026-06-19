# singbox-ffi

`singbox-ffi` is the native core package for apps such as LitheNet. It builds a
small C ABI wrapper around `github.com/sagernet/sing-box/experimental/libbox`
and publishes native artifacts for GUI apps to download.

LitheNet should not compile this Go code directly. Its Flutter build should
depend on this Flutter FFI plugin, use prebuilt `singbox-ffi` artifacts, and
build only the Flutter UI.

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

## Flutter FFI Plugin

The repository root is now a publishable Flutter FFI plugin named
`singbox_ffi`. The plugin owns platform packaging and linking for the native
artifacts:

- Windows: bundle `windows/artifacts/x64/singboxffi.dll`.
- Linux: bundle `linux/artifacts/x86_64/libsingboxffi.so`.
- macOS: link/package `macos/Libraries/libsingboxffi.dylib`,
  `macos/Libraries/libsingboxffi.a`, or a vendored framework/xcframework.
- Android: package ABI-specific `.so` files from
  `android/src/main/jniLibs/<abi>/libsingboxffi.so`.
- iOS: link `ios/Libraries/libsingboxffi.a` or a vendored framework/xcframework so
  `DynamicLibrary.process()` can resolve native symbols.

The package expects prebuilt native artifacts to be generated before publishing
or consuming the plugin. Flutter build files intentionally fail fast when a
desktop artifact is missing.

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
flutter pub get
cd examples\flutter
flutter pub get
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
SingboxRawBindings.openBundled([path])
SingboxRawBindings.openDefault([path])
```

High-level Dart API:

```dart
SingboxFfi.open([path])
SingboxFfi.openBundled([path])
SingboxFfi.openDefault([path])
SingboxFfi.fromLibrary(library)
SingboxFfi.process()
SingboxFfi.defaultLibraryName
SingboxFfi.openBundledLibrary([path])
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

Use the bundled plugin default in Flutter apps:

```dart
final core = SingboxFfi.openBundled();
```

`SingboxFfi.openBundled([path])` follows the packaging strategy used by the
plugin:

1. Use the explicit `path`, when provided.
2. On iOS, use `DynamicLibrary.process()` because the plugin links the static
   archive or framework into the app.
3. On Android, Windows, Linux, and default macOS builds, use dynamic loading via
   `SingboxFfi.openDefault()`.

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
mode for the Dart API alone. It only works after the native library has
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
- Linux desktop: package `libsingboxffi.so` through CMake.
- macOS desktop: package `libsingboxffi.dylib` by default. Static
  archive/framework builds are supported, but app code should call
  `SingboxFfi.process()` only for those builds.
- Android: package ABI-specific `.so` files through the Flutter FFI plugin.
- iOS: link a static archive or framework through CocoaPods; `openBundled()`
  uses `DynamicLibrary.process()` so Dart can resolve `sb_version`.

Mobile and static mode are handled by the Flutter FFI plugin scaffolding in
this package. It owns Windows/macOS/Linux native library linking or packaging,
Android ABI `.so` packaging, iOS/macOS static archive or framework linking, and
the platform setup required for `DynamicLibrary.process()` to find
`sb_version`.

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

The repository root is the `singbox_ffi` Flutter FFI plugin. LitheNet should
depend on it directly:

```yaml
dependencies:
  singbox_ffi:
    path: ../singbox-ffi
```

`examples/flutter` is only a smoke/example package. It depends on the root
plugin with `path: ../..` and should not be consumed as the public API.

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
