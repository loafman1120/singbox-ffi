# iOS Native Artifacts

Place the iOS static archive here before publishing or consuming the plugin:

```text
ios/Libraries/libsingboxffi.a
```

The podspec force-loads this archive so symbols such as `sb_version` remain
visible to `DynamicLibrary.process()`.

Alternatively, place a framework or xcframework under `ios/Frameworks/`.
