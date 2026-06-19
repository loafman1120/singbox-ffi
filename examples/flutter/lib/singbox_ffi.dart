import 'dart:ffi';
import 'dart:io';

import 'package:ffi/ffi.dart';

final class SbInitOptions extends Struct {
  external Pointer<Utf8> basePath;
  external Pointer<Utf8> workingPath;
  external Pointer<Utf8> tempPath;
  external Pointer<Utf8> locale;
  external Pointer<Utf8> commandSecret;

  @Int32()
  external int commandPort;

  @Int32()
  external int logMaxLines;

  @Bool()
  external bool debug;

  @Bool()
  external bool oomKillerEnabled;

  @Bool()
  external bool oomKillerDisabled;

  @Int64()
  external int oomMemoryLimit;
}

typedef SbHandle = int;

typedef SbVersionNative = Pointer<Utf8> Function();
typedef SbVersionDart = Pointer<Utf8> Function();

typedef SbGoVersionNative = Pointer<Utf8> Function();
typedef SbGoVersionDart = Pointer<Utf8> Function();

typedef SbFreeStringNative = Void Function(Pointer<Utf8>);
typedef SbFreeStringDart = void Function(Pointer<Utf8>);

typedef SbInitNative = Int32 Function(
  Pointer<SbInitOptions>,
  Pointer<Pointer<Utf8>>,
);
typedef SbInitDart = int Function(
  Pointer<SbInitOptions>,
  Pointer<Pointer<Utf8>>,
);

typedef SbCheckConfigNative = Int32 Function(
  Pointer<Utf8>,
  Pointer<Pointer<Utf8>>,
);
typedef SbCheckConfigDart = int Function(
  Pointer<Utf8>,
  Pointer<Pointer<Utf8>>,
);

typedef SbStartNative = Int32 Function(
  Pointer<Utf8>,
  Pointer<Uint64>,
  Pointer<Pointer<Utf8>>,
);
typedef SbStartDart = int Function(
  Pointer<Utf8>,
  Pointer<Uint64>,
  Pointer<Pointer<Utf8>>,
);

typedef SbReloadNative = Int32 Function(
  Uint64,
  Pointer<Utf8>,
  Pointer<Pointer<Utf8>>,
);
typedef SbReloadDart = int Function(
  int,
  Pointer<Utf8>,
  Pointer<Pointer<Utf8>>,
);

typedef SbStopNative = Int32 Function(Uint64, Pointer<Pointer<Utf8>>);
typedef SbStopDart = int Function(int, Pointer<Pointer<Utf8>>);

typedef SbFreeHandleNative = Int32 Function(Uint64);
typedef SbFreeHandleDart = int Function(int);

final class SingboxNativeSymbols {
  const SingboxNativeSymbols._();

  static const version = 'sb_version';
  static const goVersion = 'sb_go_version';
  static const freeString = 'sb_free_string';
  static const init = 'sb_init';
  static const checkConfig = 'sb_check_config';
  static const start = 'sb_start';
  static const reload = 'sb_reload';
  static const stop = 'sb_stop';
  static const freeHandle = 'sb_free_handle';
}

class SingboxRawBindings {
  SingboxRawBindings(this.library)
      : sbVersion = library.lookupFunction<SbVersionNative, SbVersionDart>(
          SingboxNativeSymbols.version,
        ),
        sbGoVersion =
            library.lookupFunction<SbGoVersionNative, SbGoVersionDart>(
          SingboxNativeSymbols.goVersion,
        ),
        sbFreeString =
            library.lookupFunction<SbFreeStringNative, SbFreeStringDart>(
          SingboxNativeSymbols.freeString,
        ),
        sbInit = library.lookupFunction<SbInitNative, SbInitDart>(
          SingboxNativeSymbols.init,
        ),
        sbCheckConfig =
            library.lookupFunction<SbCheckConfigNative, SbCheckConfigDart>(
          SingboxNativeSymbols.checkConfig,
        ),
        sbStart = library.lookupFunction<SbStartNative, SbStartDart>(
          SingboxNativeSymbols.start,
        ),
        sbReload = library.lookupFunction<SbReloadNative, SbReloadDart>(
          SingboxNativeSymbols.reload,
        ),
        sbStop = library.lookupFunction<SbStopNative, SbStopDart>(
          SingboxNativeSymbols.stop,
        ),
        sbFreeHandle =
            library.lookupFunction<SbFreeHandleNative, SbFreeHandleDart>(
          SingboxNativeSymbols.freeHandle,
        );

  factory SingboxRawBindings.open([String? path]) {
    return SingboxRawBindings(
      DynamicLibrary.open(path ?? SingboxFfi.defaultLibraryName),
    );
  }

  factory SingboxRawBindings.openDefault([String? path]) {
    return SingboxRawBindings(SingboxFfi.openDefaultLibrary(path));
  }

  factory SingboxRawBindings.process() {
    return SingboxRawBindings(DynamicLibrary.process());
  }

  final DynamicLibrary library;
  final SbVersionDart sbVersion;
  final SbGoVersionDart sbGoVersion;
  final SbFreeStringDart sbFreeString;
  final SbInitDart sbInit;
  final SbCheckConfigDart sbCheckConfig;
  final SbStartDart sbStart;
  final SbReloadDart sbReload;
  final SbStopDart sbStop;
  final SbFreeHandleDart sbFreeHandle;
}

class SingboxException implements Exception {
  SingboxException(this.message);

  final String message;

  @override
  String toString() => 'SingboxException: $message';
}

class SingboxInitOptions {
  const SingboxInitOptions({
    this.basePath = '.',
    this.workingPath = '.',
    this.tempPath = '.',
    this.locale,
    this.commandSecret = 'example-secret',
    this.commandPort = 0,
    this.logMaxLines = 300,
    this.debug = false,
    this.oomKillerEnabled = false,
    this.oomKillerDisabled = true,
    this.oomMemoryLimit = 0,
  });

  final String basePath;
  final String workingPath;
  final String tempPath;
  final String? locale;
  final String commandSecret;
  final int commandPort;
  final int logMaxLines;
  final bool debug;
  final bool oomKillerEnabled;
  final bool oomKillerDisabled;
  final int oomMemoryLimit;
}

class SingboxFfi {
  SingboxFfi._(this.raw);

  factory SingboxFfi.fromLibrary(DynamicLibrary library) {
    return SingboxFfi._(SingboxRawBindings(library));
  }

  factory SingboxFfi.open([String? path]) {
    return SingboxFfi._(SingboxRawBindings.open(path));
  }

  factory SingboxFfi.openDefault([String? path]) {
    return SingboxFfi._(SingboxRawBindings.openDefault(path));
  }

  factory SingboxFfi.process() {
    return SingboxFfi._(SingboxRawBindings.process());
  }

  static String get defaultLibraryName {
    if (Platform.isWindows) {
      return 'singboxffi.dll';
    }
    if (Platform.isMacOS || Platform.isIOS) {
      return 'libsingboxffi.dylib';
    }
    return 'libsingboxffi.so';
  }

  static DynamicLibrary openDefaultLibrary([String? path]) {
    if (path != null && path.isNotEmpty) {
      return DynamicLibrary.open(path);
    }

    final libraryName = defaultLibraryName;
    final searched = <String>{};
    final candidates = <String>[
      _joinPath(Directory.current.path, libraryName),
      _joinPath(File(Platform.executable).parent.path, libraryName),
      _joinPath(File(Platform.resolvedExecutable).parent.path, libraryName),
    ];

    Object? lastError;
    for (final candidate in candidates) {
      if (!searched.add(candidate)) {
        continue;
      }
      try {
        return DynamicLibrary.open(candidate);
      } catch (error) {
        lastError = error;
      }
    }

    try {
      return DynamicLibrary.open(libraryName);
    } catch (error) {
      throw SingboxException(
        'failed to open $libraryName from default search paths: '
        '${lastError ?? error}',
      );
    }
  }

  static String _joinPath(String directory, String fileName) {
    if (directory.isEmpty || directory == '.') {
      return fileName;
    }
    final separator = Platform.pathSeparator;
    if (directory.endsWith(separator) ||
        directory.endsWith('/') ||
        directory.endsWith('\\')) {
      return '$directory$fileName';
    }
    return '$directory$separator$fileName';
  }

  final SingboxRawBindings raw;

  String version() => takeString(raw.sbVersion());

  String goVersion() => takeString(raw.sbGoVersion());

  void init([SingboxInitOptions options = const SingboxInitOptions()]) {
    final opts = calloc<SbInitOptions>();
    final errOut = calloc<Pointer<Utf8>>();
    final allocations = <Pointer<Utf8>>[];

    Pointer<Utf8> nativeString(String? value) {
      if (value == null) {
        return nullptr;
      }
      final pointer = value.toNativeUtf8(allocator: calloc);
      allocations.add(pointer);
      return pointer;
    }

    try {
      opts.ref
        ..basePath = nativeString(options.basePath)
        ..workingPath = nativeString(options.workingPath)
        ..tempPath = nativeString(options.tempPath)
        ..locale = nativeString(options.locale)
        ..commandSecret = nativeString(options.commandSecret)
        ..commandPort = options.commandPort
        ..logMaxLines = options.logMaxLines
        ..debug = options.debug
        ..oomKillerEnabled = options.oomKillerEnabled
        ..oomKillerDisabled = options.oomKillerDisabled
        ..oomMemoryLimit = options.oomMemoryLimit;

      final code = raw.sbInit(opts, errOut);
      if (code != 0) {
        throw SingboxException(takeError(errOut));
      }
    } finally {
      for (final pointer in allocations) {
        calloc.free(pointer);
      }
      calloc.free(errOut);
      calloc.free(opts);
    }
  }

  void checkConfig(String configJson) {
    final config = configJson.toNativeUtf8(allocator: calloc);
    final errOut = calloc<Pointer<Utf8>>();
    try {
      final code = raw.sbCheckConfig(config, errOut);
      if (code != 0) {
        throw SingboxException(takeError(errOut));
      }
    } finally {
      calloc.free(config);
      calloc.free(errOut);
    }
  }

  SingboxService start(String configJson) {
    final config = configJson.toNativeUtf8(allocator: calloc);
    final handleOut = calloc<Uint64>();
    final errOut = calloc<Pointer<Utf8>>();
    try {
      final code = raw.sbStart(config, handleOut, errOut);
      if (code != 0) {
        throw SingboxException(takeError(errOut));
      }
      return SingboxService._(this, handleOut.value);
    } finally {
      calloc.free(config);
      calloc.free(handleOut);
      calloc.free(errOut);
    }
  }

  void reload(SbHandle handle, String configJson) {
    final config = configJson.toNativeUtf8(allocator: calloc);
    final errOut = calloc<Pointer<Utf8>>();
    try {
      final code = raw.sbReload(handle, config, errOut);
      if (code != 0) {
        throw SingboxException(takeError(errOut));
      }
    } finally {
      calloc.free(config);
      calloc.free(errOut);
    }
  }

  void stop(SbHandle handle) {
    final errOut = calloc<Pointer<Utf8>>();
    try {
      final code = raw.sbStop(handle, errOut);
      if (code != 0) {
        throw SingboxException(takeError(errOut));
      }
    } finally {
      calloc.free(errOut);
    }
  }

  void freeHandle(SbHandle handle) {
    final code = raw.sbFreeHandle(handle);
    if (code != 0) {
      throw SingboxException('invalid handle');
    }
  }

  void freeString(Pointer<Utf8> pointer) {
    if (pointer != nullptr) {
      raw.sbFreeString(pointer);
    }
  }

  String takeString(Pointer<Utf8> pointer) {
    if (pointer == nullptr) {
      return '';
    }
    try {
      return pointer.toDartString();
    } finally {
      freeString(pointer);
    }
  }

  String takeError(Pointer<Pointer<Utf8>> errOut) {
    final pointer = errOut.value;
    if (pointer == nullptr) {
      return 'unknown error';
    }
    return takeString(pointer);
  }
}

class SingboxService {
  SingboxService._(this._ffi, this.handle);

  final SingboxFfi _ffi;
  final SbHandle handle;
  bool _closed = false;

  void reload(String configJson) {
    if (_closed) {
      throw SingboxException('service is closed');
    }
    _ffi.reload(handle, configJson);
  }

  void close() {
    if (_closed) {
      return;
    }
    _closed = true;
    try {
      _ffi.stop(handle);
    } finally {
      _ffi.freeHandle(handle);
    }
  }
}
