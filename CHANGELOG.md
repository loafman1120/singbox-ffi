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
