## 0.1.6

- Update the embedded sing-box/libbox dependency to `v1.14.0-alpha.36`.
- Start desktop services through `daemon.StartedService` so Windows TUN
  inbounds use sing-box/sing-tun's native adapter path instead of the libbox
  platform `OpenTun` callback.
- Split the native Go FFI runtime, logging, initialization, and desktop service
  code into focused files for easier maintenance.

## 0.1.5

- Update the embedded sing-box/libbox dependency to `v1.14.0-alpha.35`.
- Refresh related upstream runtime dependencies used by the native core.

## 0.1.4

- Add structured error metadata with `SingboxErrorKind` and optional error
  codes on `SingboxException`.
- Add native service state snapshots through `SingboxService.state()`.
- Make repeated native stop calls for the same handle idempotent.
- Add desktop system proxy status, enable, and restore support through the
  libbox system proxy callbacks and Dart wrapper APIs.

## 0.1.2

- Add Dart-facing log draining and streaming APIs.
- Add native `sb_drain_logs` and `sb_clear_logs` ABI functions backed by
  libbox's started-service log subscription.
- Add `SingboxLogEvent`, `SingboxService.logs()`, `drainLogs()`, and
  `clearLogs()` for UI log panels.

## 0.1.1

- Improve the pub.dev README with clearer install, artifact, API, and feature
  coverage documentation.
- Add pub.dev package metadata and a recognized `example/main.dart`.
- Add Dartdoc comments for the public FFI wrapper API.

## 0.1.0

- Package the Dart FFI bindings as the `singbox_ffi` Flutter FFI plugin.
- Add platform scaffolding for Windows, Linux, macOS, Android, and iOS native
  artifacts.
- Add `SingboxFfi.openBundled()` as the Flutter plugin default loader.
- Add `SingboxFfi.openDefault()` for resilient dynamic library lookup.
- Link iOS static archives for `DynamicLibrary.process()` and support
  Darwin framework/xcframework artifacts.
- Document static/process mode requirements for plugin-linked native symbols.
