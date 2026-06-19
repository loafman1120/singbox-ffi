## 0.1.0

- Package the Dart FFI bindings as the `singbox_ffi` Flutter FFI plugin.
- Add platform scaffolding for Windows, Linux, macOS, Android, and iOS native
  artifacts.
- Add `SingboxFfi.openBundled()` as the Flutter plugin default loader.
- Add `SingboxFfi.openDefault()` for resilient dynamic library lookup.
- Link iOS static archives for `DynamicLibrary.process()` and support
  Darwin framework/xcframework artifacts.
- Document static/process mode requirements for plugin-linked native symbols.
