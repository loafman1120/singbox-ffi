# Linux Native Artifacts

Place prebuilt shared libraries under architecture-specific folders before
publishing or consuming the plugin:

```text
linux/artifacts/x86_64/libsingboxffi.so
linux/artifacts/aarch64/libsingboxffi.so
```

The Flutter Linux build reads `linux/CMakeLists.txt` and bundles the matching
library through `singbox_ffi_bundled_libraries`.
