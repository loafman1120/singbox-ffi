# singbox_ffi

Flutter/Dart FFI bindings for a small native wrapper around
[`github.com/sagernet/sing-box/experimental/libbox`](https://github.com/SagerNet/sing-box).

Use this package when a Flutter app needs to embed a sing-box core through FFI:
validate sing-box JSON configuration, start a local proxy service, reload it,
stop it, and query the native core version.

This package is intentionally only the Dart/Flutter plugin plus native packaging
scaffolding. The large native libraries are distributed separately through this
project's GitHub Releases.

## What It Can Do

- Load a bundled or explicitly provided `singboxffi` native library.
- Query the sing-box/libbox version and Go runtime version.
- Initialize libbox with app paths, locale, command server options, logging
  limits, debug mode, and OOM options.
- Validate a sing-box JSON config before starting it.
- Start a sing-box service from JSON config and receive an opaque handle.
- Reload a running service with new JSON config.
- Stop and free a running service handle.
- Drain or stream service logs for UI log panels.
- Expose the same ABI to C, Dart raw FFI bindings, and a small Dart wrapper.

The wrapper currently supports proxy-style sing-box configs such as local
`mixed`, SOCKS, or HTTP inbounds with normal sing-box outbounds and routing.

## Current Limits

These platform integrations are not implemented by this wrapper yet:

- TUN mode (`OpenTun` returns an unsupported error).
- System proxy toggling.
- General event draining APIs beyond logs.
- SSH agent, platform shell, SFTP, user lookup, and connection-owner lookup.

If your app needs those features, treat this package as a lower-level starting
point rather than a complete sing-box GUI runtime.

## Install

Add the package:

```yaml
dependencies:
  singbox_ffi: ^0.1.2
```

Then add the native artifact for each platform you build. The pub.dev package
does not include native binaries.

| Platform | Expected artifact location |
| --- | --- |
| Windows | `windows/artifacts/x64/singboxffi.dll` |
| Linux | `linux/artifacts/x86_64/libsingboxffi.so` |
| macOS | `macos/Libraries/libsingboxffi.dylib`, `macos/Libraries/libsingboxffi.a`, or a vendored framework/xcframework |
| Android | `android/src/main/jniLibs/<abi>/libsingboxffi.so` |
| iOS | `ios/Libraries/libsingboxffi.a` or a vendored framework/xcframework |

Download prebuilt artifacts from the matching GitHub Release, or build them
locally from this repository. Desktop CMake files fail fast when a required
desktop artifact is missing so the problem is visible during `flutter build`.

## Quick Start

```dart
import 'package:singbox_ffi/singbox_ffi.dart';

const configJson = '''
{
  "log": {"level": "info"},
  "inbounds": [
    {
      "type": "mixed",
      "tag": "mixed-in",
      "listen": "127.0.0.1",
      "listen_port": 2080
    }
  ],
  "outbounds": [
    {"type": "direct", "tag": "direct"}
  ],
  "route": {"final": "direct"}
}
''';

Future<void> main() async {
  final singbox = SingboxFfi.openBundled();

  print('sing-box: ${singbox.version()}');
  print('Go: ${singbox.goVersion()}');

  singbox.init();
  singbox.checkConfig(configJson);

  final service = singbox.start(configJson);
  try {
    // The mixed proxy is now listening on 127.0.0.1:2080.
    await for (final event in service.logs()) {
      if (event.isReset) {
        // Clear your visible log list.
      } else {
        print('[${event.levelName}] ${event.message}');
      }
    }
  } finally {
    service.close();
  }
}
```

For Flutter apps, `SingboxFfi.openBundled()` is the recommended loader. It
matches the package's platform packaging:

1. Use the explicit path when you pass one.
2. On iOS, use `DynamicLibrary.process()` because the native archive/framework
   is linked into the app.
3. On Android, Windows, Linux, and the default macOS flow, open the platform
   dynamic library from normal package/executable search paths.

For command-line smoke tests or custom packaging, pass a path explicitly:

```dart
final singbox = SingboxFfi.openBundled(r'C:\path\to\singboxffi.dll');
```

## API Overview

High-level Dart API:

```dart
final core = SingboxFfi.openBundled();

core.version();
core.goVersion();
core.init(SingboxInitOptions(...));
core.checkConfig(configJson);

final service = core.start(configJson);
service.logs();
service.drainLogs();
service.clearLogs();
service.reload(nextConfigJson);
service.close();
```

Lower-level APIs are also exported for advanced integrations:

- `SingboxRawBindings` exposes typed Dart FFI function pointers.
- `SingboxNativeSymbols` contains the exported native symbol names.
- `SbInitOptions` and native typedefs mirror the C ABI.
- `SingboxFfi.fromLibrary()` lets you provide an already-opened
  `DynamicLibrary`.
- `SingboxFfi.process()` works only when the native symbols are already linked
  into the current process.

## Initialization Options

`SingboxInitOptions` maps directly to libbox setup options:

| Dart option | Meaning |
| --- | --- |
| `basePath` | Base directory used by libbox. |
| `workingPath` | Working directory for runtime state. |
| `tempPath` | Temporary directory for runtime files. |
| `locale` | Optional locale passed to libbox before setup. |
| `commandSecret` | Secret for libbox command server support. |
| `commandPort` | Command server listen port. Use `0` for automatic/default behavior. |
| `logMaxLines` | Maximum log lines retained by libbox internals. |
| `debug` | Enable libbox debug setup. |
| `oomKillerEnabled` | Enable libbox OOM killer behavior. |
| `oomKillerDisabled` | Disable libbox OOM killer behavior. Defaults to `true`. |
| `oomMemoryLimit` | Optional OOM memory limit. |

## Native ABI

The native library exports a small C ABI:

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

int32_t sb_drain_logs(sb_handle handle, int32_t max_entries,
                      char **json_out, char **err_out);
int32_t sb_clear_logs(sb_handle handle, char **err_out);
```

Strings returned by the core must be released with `sb_free_string`. Handles
returned by `sb_start` should be stopped with `sb_stop` and released with
`sb_free_handle`. The Dart `SingboxService.close()` helper does both.

`sb_drain_logs` returns a JSON array of log events. A reset event means libbox
cleared its internal log history, usually because a service started or reloaded.
Dart callers normally use `SingboxService.logs()` instead of parsing this JSON
directly.

## Build Native Artifacts Locally

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

Static archive builds use the same ABI:

```powershell
go build -trimpath -buildmode=c-archive `
  -tags "with_gvisor,with_quic,with_wireguard,with_utls,with_naive_outbound,with_purego,with_clash_api,badlinkname,tfogo_checklinkname0" `
  -ldflags "-s -w -buildid= -checklinkname=0" `
  -o build\singboxffi.a .
```

Static/process loading is for apps that link the native symbols into the app
binary or Flutter runner. It is not a general fallback for a missing dynamic
library.

## Run The Smoke Proxy

The repository includes a Dart smoke example that starts a real local `mixed`
proxy on `127.0.0.1:2080`.

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

## Repository Notes

- The repository root is the publishable Flutter FFI plugin.
- `example/` is the pub.dev-facing example.
- `examples/flutter` and `examples/c` are repository smoke tests.
- Release automation builds native artifacts and attaches complete packages to
  GitHub Releases.
- The lightweight pub.dev package excludes generated native binaries.

## License Notice

This project links against `github.com/sagernet/sing-box/experimental/libbox`.
sing-box is distributed under the GNU General Public License, version 3 or
later, with the additional upstream naming restriction copied in
`LICENSE.sing-box`.

Distributions of this wrapper, linked binaries, and applications embedding the
produced native library should carry the corresponding GPL notice and must not
use the sing-box name or imply association with the upstream application without
prior consent.
