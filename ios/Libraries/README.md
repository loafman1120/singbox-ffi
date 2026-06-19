# iOS Native Artifacts

Place the iOS static archive here before publishing or consuming the plugin:

```text
ios/Libraries/libsingboxffi.a
```

Alternatively, place a framework under `ios/Frameworks/`. The podspec links
vendored archives/frameworks into the app so `SingboxFfi.process()` can resolve
symbols such as `sb_version`.
