# Windows Native Artifacts

Place prebuilt DLLs under architecture-specific folders before publishing or
consuming the plugin:

```text
windows/artifacts/x64/singboxffi.dll
windows/artifacts/arm64/singboxffi.dll
```

The Flutter Windows build reads `windows/CMakeLists.txt` and bundles the
matching DLL through `singbox_ffi_bundled_libraries`.
