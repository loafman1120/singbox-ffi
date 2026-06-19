# macOS Native Artifacts

For dynamic loading, place:

```text
macos/Libraries/libsingboxffi.dylib
```

For static/process mode, place:

```text
macos/Libraries/libsingboxffi.a
```

Frameworks can also be placed under `macos/Frameworks/`. The podspec links
vendored archives/frameworks and packages vendored dylibs.
