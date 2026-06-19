# Android Native Artifacts

Place prebuilt `libsingboxffi.so` files in ABI-specific folders before
publishing or consuming the plugin:

```text
android/src/main/jniLibs/arm64-v8a/libsingboxffi.so
android/src/main/jniLibs/armeabi-v7a/libsingboxffi.so
android/src/main/jniLibs/x86_64/libsingboxffi.so
```

Flutter packages these files into the APK/AAB. `SingboxFfi.openDefault()` can
then load `libsingboxffi.so`; static/process mode is not the Android default.
