# LitheNet

LitheNet is a lightweight Flutter proxy client core powered by sing-box. It uses
a small C ABI around `github.com/sagernet/sing-box/experimental/libbox`, then
loads that native library from Dart with `dart:ffi`.

This repository is the first usable slice of the product: build the native core,
run a Dart process, and get a real local `mixed` proxy on `127.0.0.1:2080`.

## Why

- Flutter UI, native sing-box engine
- Tiny and stable FFI boundary
- No dependency on the unavailable upstream `sing-ffi` generator
- Desktop-first local proxy mode before TUN/system-proxy work

## Build Core

Windows with MSYS2 UCRT64 GCC:

```powershell
$env:CGO_ENABLED = "1"
$env:CC = "C:\msys64\ucrt64\bin\gcc.exe"
$env:PATH = "C:\msys64\ucrt64\bin;$env:PATH"

go build -trimpath -buildmode=c-shared `
  -tags "with_gvisor,with_quic,with_wireguard,with_utls,with_naive_outbound,with_purego,with_clash_api,badlinkname,tfogo_checklinkname0" `
  -ldflags "-s -w -buildid= -checklinkname=0" `
  -o build\lithenetcore.dll .
```

Outputs:

```text
build/lithenetcore.dll
build/lithenetcore.h
```

Linux/macOS use the same `go build -buildmode=c-shared` flow and output
`liblithenetcore.so` or `liblithenetcore.dylib`.

## Run The Dart Proxy

```powershell
cd examples\flutter
dart pub get
dart run bin\proxy.dart ..\..\build\lithenetcore.dll
```

In another terminal:

```powershell
curl.exe -x socks5h://127.0.0.1:2080 https://example.com
```

Press `Ctrl+C` in the Dart process to stop the proxy.

## Flutter Integration

Reuse `examples/flutter/lib/singbox_ffi.dart` in a Flutter desktop app:

```dart
final core = SingboxFfi.open('lithenetcore.dll');
core.init();
final service = core.start(configJson);

// Later:
service.close();
```

Ship the native library beside the Flutter executable, or pass an absolute path
from your installer/app data directory.

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

## Status

Implemented:

- config validation
- local mixed/SOCKS/HTTP proxy start, reload, stop
- desktop stub platform interface
- Dart FFI binding
- Dart proxy runner

Next:

- Flutter UI shell
- profile import and subscription update
- log/event draining
- system proxy toggle
- TUN mode

## License Notice

This project links against `github.com/sagernet/sing-box/experimental/libbox`.
sing-box is distributed under the GNU General Public License, version 3 or later,
with the additional upstream naming restriction copied in `LICENSE.sing-box`.

Distributions of this wrapper, linked binaries, and applications embedding the
produced native library should carry the corresponding GPL notice and must not
use the sing-box name or imply association with the upstream application without
prior consent.
